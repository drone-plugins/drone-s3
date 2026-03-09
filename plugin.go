package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/mattn/go-zglob"
	log "github.com/sirupsen/logrus"
)

var errSkip = fmt.Errorf("skip")

// Plugin defines the S3 plugin parameters.
type Plugin struct {
	Endpoint              string
	Key                   string
	Secret                string
	AssumeRole            string
	AssumeRoleSessionName string
	Bucket                string
	UserRoleArn           string
	UserRoleExternalID    string

	// if not "", enable server-side encryption
	// valid values are:
	//     AES256
	//     aws:kms
	Encryption string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string

	// if true, plugin is set to download mode, which means `source` from the bucket will be downloaded
	Download bool

	// Indicates the files ACL, which should be one
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	Access string

	// Sets the content type on each uploaded object based on a extension map
	ContentType map[string]string

	// Sets the content encoding on each uploaded object based on a extension map
	ContentEncoding map[string]string

	// Sets the Cache-Control header on each uploaded object based on a extension map
	CacheControl map[string]string

	// Sets the storage class, affects the storage backend costs
	StorageClass string

	// Copies the files from the specified directory.
	// Regexp matching will apply to match multiple
	// files
	//
	// Examples:
	//    /path/to/file
	//    /path/to/*.txt
	//    /path/to/*/*.txt
	//    /path/to/**
	Source string
	Target string

	// Strip the prefix from the target path (supports wildcards)
	StripPrefix string

	// Exclude files matching this pattern.
	Exclude []string

	// Use path style instead of domain style.
	//
	// Should be true for minio and false for AWS.
	PathStyle bool
	// Dry run without uploading/
	DryRun bool

	// set externalID for assume role
	ExternalID string

	// set OIDC ID Token to retrieve temporary credentials
	IdToken string

	// AWS session token for temporary credentials (e.g., from EKS Pod Identity, IRSA, STS)
	SessionToken string
}

// Exec runs the plugin
func (p *Plugin) Exec() error {
	if p.Download {
		p.Source = normalizePath(p.Source)
		p.Target = normalizePath(p.Target)
	} else {
		p.Target = strings.TrimPrefix(p.Target, "/")
	}

	ctx := context.Background()

	client := p.createS3Client(ctx)

	if p.Download {
		sourceDir := normalizePath(p.Source)
		return p.downloadS3Objects(ctx, client, sourceDir)
	}

	log.WithFields(log.Fields{
		"region":   p.Region,
		"endpoint": p.Endpoint,
		"bucket":   p.Bucket,
	}).Info("Attempting to upload")

	matches, err := matches(p.Source, p.Exclude)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Could not match files")
		return err
	}

	normalizedStrip := strings.ReplaceAll(p.StripPrefix, "\\", "/")
	if p.StripPrefix != "" && strings.HasPrefix(normalizedStrip, "/") {
		if err := validateStripPrefix(p.StripPrefix); err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"pattern": p.StripPrefix,
			}).Error("Invalid strip_prefix pattern")
			return err
		}
	}

	var compiled *regexp.Regexp
	if normalizedStrip != "" && strings.HasPrefix(normalizedStrip, "/") && strings.ContainsAny(normalizedStrip, "*?") {
		var err error
		compiled, err = patternToRegex(normalizedStrip)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"pattern": p.StripPrefix,
			}).Error("Failed to compile strip_prefix pattern")
			return err
		}
	}

	anyMatched := false

	for _, match := range matches {
		if err := isDir(match, matches); err != nil {
			if err == errSkip {
				continue
			}
			log.WithFields(log.Fields{
				"error": err,
				"match": match,
			}).Error("Directory specified without glob pattern")
			return err
		}

		stripped := match
		matched := false
		if normalizedStrip != "" {
			if strings.HasPrefix(normalizedStrip, "/") {
				var err error
				stripped, matched, err = stripWildcardPrefixWithRegex(match, normalizedStrip, compiled)
				if err != nil {
					log.WithFields(log.Fields{
						"error":   err,
						"path":    match,
						"pattern": p.StripPrefix,
					}).Warn("Failed to strip prefix, using original path")
					stripped = match
				}
			} else {
				m := filepath.ToSlash(match)
				trimmed := strings.TrimPrefix(m, normalizedStrip)
				if trimmed != m {
					matched = true
					stripped = trimmed
				} else {
					stripped = match
				}
			}
		}
		if matched {
			anyMatched = true
		}

		var target string
		if normalizedStrip != "" && !strings.HasPrefix(normalizedStrip, "/") {
			target = resolveKey(p.Target, filepath.ToSlash(match), p.StripPrefix)
		} else {
			rel := strings.TrimPrefix(filepath.ToSlash(stripped), "/")
			target = filepath.ToSlash(filepath.Join(p.Target, rel))
		}

		contentType := matchExtension(match, p.ContentType)
		contentEncoding := matchExtension(match, p.ContentEncoding)
		cacheControl := matchExtension(match, p.CacheControl)

		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(match))
			if contentType == "" {
				contentType = "application/octet-stream"
			}
		}

		log.WithFields(log.Fields{
			"name":   match,
			"bucket": p.Bucket,
			"target": target,
		}).Info("Uploading file")

		if p.DryRun {
			removed := ""
			if matched {
				orig := filepath.ToSlash(match)
				rem := strings.TrimSuffix(orig, filepath.ToSlash(stripped))
				removed = rem
			}
			log.WithFields(log.Fields{
				"name":           match,
				"bucket":         p.Bucket,
				"target":         target,
				"strip_pattern":  p.StripPrefix,
				"removed_prefix": removed,
			}).Info("Dry-run: would upload")
			continue
		}

		f, err := os.Open(match)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"file":  match,
			}).Error("Problem opening file")
			return err
		}
		defer f.Close()

		putObjectInput := &s3.PutObjectInput{
			Body:   f,
			Bucket: &(p.Bucket),
			Key:    &target,
		}

		if contentType != "" {
			putObjectInput.ContentType = aws.String(contentType)
		}

		if contentEncoding != "" {
			putObjectInput.ContentEncoding = aws.String(contentEncoding)
		}

		if cacheControl != "" {
			putObjectInput.CacheControl = aws.String(cacheControl)
		}

		if p.Encryption != "" {
			putObjectInput.ServerSideEncryption = s3types.ServerSideEncryption(p.Encryption)
		}

		if p.StorageClass != "" {
			putObjectInput.StorageClass = s3types.StorageClass(p.StorageClass)
		}

		if p.Access != "" {
			putObjectInput.ACL = s3types.ObjectCannedACL(p.Access)
		}

		_, err = client.PutObject(ctx, putObjectInput)

		if err != nil {
			log.WithFields(log.Fields{
				"name":   match,
				"bucket": p.Bucket,
				"target": target,
				"error":  err,
			}).Error("Could not upload file")

			return err
		}
		f.Close()
	}

	if normalizedStrip != "" && !anyMatched {
		log.WithFields(log.Fields{
			"pattern": p.StripPrefix,
		}).Warn("strip_prefix did not match any paths; keys will include original path")
	}

	return nil
}

func matches(include string, exclude []string) ([]string, error) {
	matches, err := zglob.Glob(include)
	if err != nil {
		return nil, err
	}
	if len(exclude) == 0 {
		return matches, nil
	}

	excludem := map[string]bool{}
	for _, pattern := range exclude {
		excludes, err := zglob.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range excludes {
			excludem[match] = true
		}
	}

	var included []string
	for _, include := range matches {
		_, ok := excludem[include]
		if ok {
			continue
		}
		included = append(included, include)
	}
	return included, nil
}

func matchExtension(match string, stringMap map[string]string) string {
	for pattern := range stringMap {
		matched, err := regexp.MatchString(pattern, match)
		if err != nil {
			panic(err)
		}
		if matched {
			return stringMap[pattern]
		}
	}
	return ""
}

func assumeRole(ctx context.Context, roleArn, roleSessionName, externalID, region string) aws.CredentialsProvider {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load AWS config for assume role: %v", err)
	}
	stsSvc := sts.NewFromConfig(cfg)
	duration := time.Hour * 1
	provider := stscreds.NewAssumeRoleProvider(stsSvc, roleArn, func(o *stscreds.AssumeRoleOptions) {
		o.Duration = duration
		o.RoleSessionName = roleSessionName
		if externalID != "" {
			o.ExternalID = &externalID
		}
	})
	return aws.NewCredentialsCache(provider)
}

func resolveKey(target, srcPath, stripPrefix string) string {
	key := filepath.Join(target, strings.TrimPrefix(srcPath, filepath.ToSlash(stripPrefix)))
	key = filepath.ToSlash(key)
	key = strings.TrimPrefix(key, "/")
	return key
}

func resolveSource(sourceDir, source, stripPrefix string) string {
	path := strings.TrimPrefix(strings.TrimPrefix(source, sourceDir), "/")
	return stripPrefix + path
}

func normalizeEndpoint(endpoint string) string {
	if endpoint == "" || strings.Contains(endpoint, "://") {
		return endpoint
	}
	return "https://" + endpoint
}

func isDir(source string, matches []string) error {
	stat, err := os.Stat(source)
	if err != nil {
		return errSkip
	}
	if stat.IsDir() {
		count := 0
		for _, match := range matches {
			if strings.HasPrefix(match, source) {
				count++
			}
		}
		if count <= 1 {
			return fmt.Errorf("directory '%s' specified without glob pattern. Use a pattern like '%s/*' or '%s/**' to upload directory contents", source, source, source)
		}
		return errSkip
	}
	return nil
}

func normalizePath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(path), "/")
}

func (p *Plugin) downloadS3Object(ctx context.Context, client *s3.Client, sourceDir, key, target string) error {
	log.WithFields(log.Fields{
		"bucket": p.Bucket,
		"key":    key,
	}).Info("Getting S3 object")

	obj, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.Bucket,
		Key:    &key,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"bucket": p.Bucket,
			"key":    key,
		}).Error("Cannot get S3 object")
		return err
	}
	defer obj.Body.Close()

	destination := filepath.Join(p.Target, target)
	dir := filepath.Dir(destination)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	f, err := os.Create(destination)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"file":  destination,
		}).Error("Failed to create file")
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, obj.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"file":  destination,
		}).Error("Failed to write file")
		return err
	}

	return nil
}

func (p *Plugin) downloadS3Objects(ctx context.Context, client *s3.Client, sourceDir string) error {
	log.WithFields(log.Fields{
		"bucket": p.Bucket,
		"dir":    sourceDir,
	}).Info("Listing S3 directory")

	list, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &p.Bucket,
		Prefix: &sourceDir,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"bucket": p.Bucket,
			"dir":    sourceDir,
		}).Error("Cannot list S3 directory")
		return err
	}

	for _, item := range list.Contents {
		target := resolveSource(sourceDir, *item.Key, p.StripPrefix)
		if err := p.downloadS3Object(ctx, client, sourceDir, *item.Key, target); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) createS3Client(ctx context.Context) *s3.Client {
	optFns := []func(*config.LoadOptions) error{
		config.WithRegion(p.Region),
	}

	if p.Key != "" && p.Secret != "" {
		if p.SessionToken != "" {
			log.Info("Using static credentials with session token (temporary credentials)")
		} else {
			log.Info("Using static credentials (access key and secret key)")
		}
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(p.Key, p.Secret, p.SessionToken),
		))
	} else if p.IdToken != "" && p.AssumeRole != "" {
		creds, err := assumeRoleWithWebIdentity(ctx, p.AssumeRole, p.AssumeRoleSessionName, p.IdToken, p.Region)
		if err != nil {
			log.Fatalf("failed to assume role with web identity: %v", err)
		}
		optFns = append(optFns, config.WithCredentialsProvider(creds))
	} else if p.AssumeRole != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			assumeRole(ctx, p.AssumeRole, p.AssumeRoleSessionName, p.ExternalID, p.Region),
		))
	} else {
		// No explicit credentials provided, falling back to the default AWS SDK credential chain.
		// The SDK will check: env vars -> shared credentials -> container credentials -> EC2 IMDS
		if containerCredsURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI"); containerCredsURI != "" {
			log.WithField("uri", containerCredsURI).Info(
				"No explicit credentials provided; AWS SDK will use EKS Pod Identity / container credentials")
		} else if os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI") != "" {
			log.Info("No explicit credentials provided; AWS SDK will use ECS container credentials")
		} else if os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" {
			log.Info("No explicit credentials provided; AWS SDK will use IRSA (Web Identity Token)")
		} else {
			log.Warn("No AWS credentials provided and no container/identity credential source detected. " +
				"Falling back to EC2 instance metadata (IMDS). This may fail if not running on EC2.")
		}
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Opts := []func(*s3.Options){}

	if p.Endpoint != "" {
		endpoint := normalizeEndpoint(p.Endpoint)
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = p.PathStyle
			// S3-compatible services (MinIO, Spaces, B2, etc.) may not support the
			// CRC32 checksums that SDK v2 sends by default with PutObject.
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		})
	} else if p.PathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(cfg, s3Opts...)

	if len(p.UserRoleArn) > 0 {
		log.WithField("UserRoleArn", p.UserRoleArn).Info("Using user role ARN")

		stsSvc := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsSvc, p.UserRoleArn, func(o *stscreds.AssumeRoleOptions) {
			if p.UserRoleExternalID != "" {
				o.ExternalID = aws.String(p.UserRoleExternalID)
			}
		})

		cfg.Credentials = aws.NewCredentialsCache(provider)
		client = s3.NewFromConfig(cfg, s3Opts...)
	}

	return client
}

func assumeRoleWithWebIdentity(ctx context.Context, roleArn, roleSessionName, idToken, region string) (aws.CredentialsProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}
	stsSvc := sts.NewFromConfig(cfg)
	result, err := stsSvc.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleArn),
		RoleSessionName:  aws.String(roleSessionName),
		WebIdentityToken: aws.String(idToken),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role with web identity: %w", err)
	}
	if result.Credentials == nil {
		return nil, fmt.Errorf("STS AssumeRoleWithWebIdentity returned nil credentials")
	}
	return credentials.NewStaticCredentialsProvider(
		*result.Credentials.AccessKeyId,
		*result.Credentials.SecretAccessKey,
		*result.Credentials.SessionToken,
	), nil
}

func validateStripPrefix(pattern string) error {
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	if !strings.HasPrefix(pattern, "/") {
		return fmt.Errorf("strip_prefix must start with '/'")
	}

	if len(pattern) >= 2 && pattern[1] == ':' {
		return fmt.Errorf("strip_prefix must be an absolute POSIX-style path (e.g. '/root/...'), drive letters are not supported")
	}

	if len(pattern) > 256 {
		return fmt.Errorf("strip_prefix pattern too long (max 256 characters)")
	}

	wildcardCount := strings.Count(pattern, "*") + strings.Count(pattern, "?")
	if wildcardCount > 20 {
		return fmt.Errorf("strip_prefix pattern contains too many wildcards (max 20)")
	}

	if strings.Contains(pattern, "//") {
		return fmt.Errorf("strip_prefix pattern contains empty segment '//'")
	}

	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if strings.Contains(part, "**") && part != "**" {
			return fmt.Errorf("'**' must be a standalone directory segment")
		}
	}

	return nil
}

func patternToRegex(pattern string) (*regexp.Regexp, error) {
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, `\*\*`, "(.+)")
	escaped = strings.ReplaceAll(escaped, `\*`, "([^/]+)")
	escaped = strings.ReplaceAll(escaped, `\?`, "([^/])")
	escaped = "^" + escaped
	return regexp.Compile(escaped)
}

func stripWildcardPrefixWithRegex(path, pattern string, re *regexp.Regexp) (string, bool, error) {
	if pattern == "" {
		return path, false, nil
	}

	path = strings.ReplaceAll(path, "\\", "/")
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	if !strings.ContainsAny(pattern, "*?") {
		if !strings.HasPrefix(path, pattern) {
			return path, false, nil
		}
		stripped := strings.TrimPrefix(path, pattern)
		if stripped == "" || stripped == "/" || strings.TrimPrefix(stripped, "/") == "" {
			return path, true, fmt.Errorf("strip_prefix removes entire path for '%s'", filepath.Base(path))
		}
		return stripped, true, nil
	}

	var err error
	if re == nil {
		re, err = patternToRegex(pattern)
		if err != nil {
			return path, false, fmt.Errorf("invalid pattern: %v", err)
		}
	}

	m := re.FindStringSubmatch(path)
	if len(m) == 0 {
		return path, false, nil
	}
	full := m[0]
	stripped := strings.TrimPrefix(path, full)
	if stripped == "" || stripped == "/" || strings.TrimPrefix(stripped, "/") == "" {
		return path, true, fmt.Errorf("strip_prefix removes entire path for '%s'", filepath.Base(path))
	}
	return stripped, true, nil
}

func stripWildcardPrefix(path, pattern string) (string, error) {
	stripped, _, err := stripWildcardPrefixWithRegex(path, pattern, nil)
	return stripped, err
}
