package main

import (
	"io"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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

	// Strip the prefix from the target path
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
}

// Exec runs the plugin
func (p *Plugin) Exec() error {
	if p.Download {
		p.Source = normalizePath(p.Source)
		p.Target = normalizePath(p.Target)
	} else {
		p.Target = strings.TrimPrefix(p.Target, "/")
	}

	// create the client
	client := p.createS3Client()

	// If in download mode, call the downloadS3Objects method
	if p.Download {
		sourceDir := normalizePath(p.Source)

		return p.downloadS3Objects(client, sourceDir)
	}

	// find the bucket
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

	for _, match := range matches {
		// skip directories
		if isDir(match, matches) {
			continue
		}

		target := resolveKey(p.Target, match, p.StripPrefix)

		contentType := matchExtension(match, p.ContentType)
		contentEncoding := matchExtension(match, p.ContentEncoding)
		cacheControl := matchExtension(match, p.CacheControl)

		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(match))

			if contentType == "" {
				contentType = "application/octet-stream"
			}
		}

		// log file for debug purposes.
		log.WithFields(log.Fields{
			"name":   match,
			"bucket": p.Bucket,
			"target": target,
		}).Info("Uploading file")

		// when executing a dry-run we exit because we don't actually want to
		// upload the file to S3.
		if p.DryRun {
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
			putObjectInput.ServerSideEncryption = aws.String(p.Encryption)
		}

		if p.StorageClass != "" {
			putObjectInput.StorageClass = &(p.StorageClass)
		}

		if p.Access != "" {
			putObjectInput.ACL = &(p.Access)
		}

		_, err = client.PutObject(putObjectInput)

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

	return nil
}

// matches is a helper function that returns a list of all files matching the
// included Glob pattern, while excluding all files that matche the exclusion
// Glob pattners.
func matches(include string, exclude []string) ([]string, error) {
	matches, err := zglob.Glob(include)
	if err != nil {
		return nil, err
	}
	if len(exclude) == 0 {
		return matches, nil
	}

	// find all files that are excluded and load into a map. we can verify
	// each file in the list is not a member of the exclusion list.
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

func assumeRole(roleArn, roleSessionName, externalID string) *credentials.Credentials {
	sess, _ := session.NewSession()
	client := sts.New(sess)
	duration := time.Hour * 1
	stsProvider := &stscreds.AssumeRoleProvider{
		Client:          client,
		Duration:        duration,
		RoleARN:         roleArn,
		RoleSessionName: roleSessionName,
	}

	if externalID != "" {
		stsProvider.ExternalID = &externalID
	}

	return credentials.NewCredentials(stsProvider)
}

// resolveKey is a helper function that returns s3 object key where file present at srcPath is uploaded to.
// srcPath is assumed to be in forward slash format
func resolveKey(target, srcPath, stripPrefix string) string {
	key := filepath.Join(target, strings.TrimPrefix(srcPath, filepath.ToSlash(stripPrefix)))
	key = filepath.ToSlash(key)
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	return key
}

func resolveSource(sourceDir, source, stripPrefix string) string {
	// Remove the leading sourceDir from the source path
	path := strings.TrimPrefix(strings.TrimPrefix(source, sourceDir), "/")

	// Add the specified stripPrefix to the resulting path
	return stripPrefix + path
}

// checks if the source path is a dir
func isDir(source string, matches []string) bool {
	stat, err := os.Stat(source)
	if err != nil {
		return true // should never happen
	}
	if stat.IsDir() {
		count := 0
		for _, match := range matches {
			if strings.HasPrefix(match, source) {
				count++
			}
		}
		if count <= 1 {
			log.Warnf("Skipping '%s' since it is a directory. Please use correct glob expression if this is unexpected.", source)
		}
		return true
	}
	return false
}

// normalizePath converts the path to a forward slash format and trims the prefix.
func normalizePath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(path), "/")
}

// downloadS3Object downloads a single object from S3
func (p *Plugin) downloadS3Object(client *s3.S3, sourceDir, key, target string) error {
	log.WithFields(log.Fields{
		"bucket": p.Bucket,
		"key":    key,
	}).Info("Getting S3 object")

	obj, err := client.GetObject(&s3.GetObjectInput{
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

	// Create the destination file path
	destination := filepath.Join(p.Target, target)
	log.Println("Destination: ", destination)

	// Extract the directory from the destination path
	dir := filepath.Dir(destination)

	// Create the directory and any necessary parent directories
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "error creating directories")
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

// downloadS3Objects downloads all objects in the specified S3 bucket path
func (p *Plugin) downloadS3Objects(client *s3.S3, sourceDir string) error {
	log.WithFields(log.Fields{
		"bucket": p.Bucket,
		"dir":    sourceDir,
	}).Info("Listing S3 directory")

	list, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
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
		// resolveSource takes a source directory, a source path, and a prefix to strip,
		// and returns a resolved target path by removing the sourceDir from the source
		// and appending the stripPrefix.
		target := resolveSource(sourceDir, *item.Key, p.StripPrefix)

		if err := p.downloadS3Object(client, sourceDir, *item.Key, target); err != nil {
			return err
		}
	}

	return nil
}

// createS3Client creates and returns an S3 client based on the plugin configuration
func (p *Plugin) createS3Client() *s3.S3 {
    conf := &aws.Config{
        Region:           aws.String(p.Region),
        Endpoint:         &p.Endpoint,
        DisableSSL:       aws.Bool(strings.HasPrefix(p.Endpoint, "http://")),
        S3ForcePathStyle: aws.Bool(p.PathStyle),
    }

    sess, err := session.NewSession(conf)
    if err != nil {
        log.Fatalf("failed to create AWS session: %v", err)
    }

    if p.Key != "" && p.Secret != "" {
        conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
    } else if p.IdToken != "" && p.AssumeRole != "" {
        creds, err := assumeRoleWithWebIdentity(sess, p.AssumeRole, p.AssumeRoleSessionName, p.IdToken)
        if err != nil {
            log.Fatalf("failed to assume role with web identity: %v", err)
        }
        conf.Credentials = creds
    } else if p.AssumeRole != "" {
        conf.Credentials = assumeRole(p.AssumeRole, p.AssumeRoleSessionName, p.ExternalID)
    } else {
        log.Warn("AWS Key and/or Secret not provided (falling back to ec2 instance profile)")
    }

	sess, err = session.NewSession(conf)
    if err != nil {
        log.Fatalf("failed to create AWS session: %v", err)
    }

    client := s3.New(sess, conf)

    if len(p.UserRoleArn) > 0 {
        log.WithFields(log.Fields{
            "UserRoleArn": p.UserRoleArn,
        }).Info("Assuming user role ARN")

        // Create new credentials by assuming the UserRoleArn with ExternalID
        creds := stscreds.NewCredentials(sess, p.UserRoleArn, func(provider *stscreds.AssumeRoleProvider) {
            if p.UserRoleExternalID != "" {
                provider.ExternalID = aws.String(p.UserRoleExternalID)
            }
        })

        // Create a new configuration with the new credentials
        confRoleArn := aws.Config{
            Region:           aws.String(p.Region),
            Endpoint:         &p.Endpoint,
            DisableSSL:       aws.Bool(strings.HasPrefix(p.Endpoint, "http://")),
            S3ForcePathStyle: aws.Bool(p.PathStyle),
            Credentials:      creds,
        }

        // Update the S3 client with the new configuration
        client = s3.New(sess, &confRoleArn)
    }

    return client
}

func assumeRoleWithWebIdentity(sess *session.Session, roleArn, roleSessionName, idToken string) (*credentials.Credentials, error) {
    svc := sts.New(sess)
    input := &sts.AssumeRoleWithWebIdentityInput{
        RoleArn:          aws.String(roleArn),
        RoleSessionName:  aws.String(roleSessionName),
        WebIdentityToken: aws.String(idToken),
    }
    result, err := svc.AssumeRoleWithWebIdentity(input)
    if err != nil {
        log.Fatalf("failed to assume role with web identity: %v", err)
    }
    return credentials.NewStaticCredentials(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken), nil
}
