package main

import (
	"mime"
	"os"
	"path/filepath"
	"strings"

	"errors"

	log "github.com/Sirupsen/logrus"
	glob "github.com/ryanuber/go-glob"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mattn/go-zglob"
)

// Plugin defines the S3 plugin parameters.
type Plugin struct {
	Endpoint string
	Key      string
	Secret   string
	Bucket   string

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

	// Indicates the files ACL, which should be one
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	Access string

	// Sets the Cache-Control header on each uploaded object
	CacheControl string

	// A standard MIME type describing the format of the object data.
	ContentType map[string]string

	// Specifies what content encodings have been applied to the object and thus
	// what decoding mechanisms must be applied to obtain the media-type referenced
	// by the Content-Type header field.
	ContentEncoding map[string]string

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

	YamlVerified bool

	// Exclude files matching this pattern.
	Exclude []string

	// Use path style instead of domain style.
	//
	// Should be true for minio and false for AWS.
	PathStyle bool
	// Dry run without uploading/
	DryRun bool
}

// Exec runs the plugin
func (p *Plugin) Exec() error {
	// normalize the target URL
	if strings.HasPrefix(p.Target, "/") {
		p.Target = p.Target[1:]
	}

	// create the client
	conf := &aws.Config{
		Region:           aws.String(p.Region),
		Endpoint:         &p.Endpoint,
		DisableSSL:       aws.Bool(strings.HasPrefix(p.Endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(p.PathStyle),
	}

	//Allowing to use the instance role or provide a key and secret
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	} else if p.YamlVerified != true {
		return errors.New("Security issue: When using instance role you must have the yaml verified")
	}
	client := s3.New(session.New(), conf)

	log.WithFields(log.Fields{
		"region":           p.Region,
		"endpoint":         p.Endpoint,
		"bucket":           p.Bucket,
		"access":           p.Access,
		"source":           p.Source,
		"target":           p.Target,
		"strip-prefix":     p.StripPrefix,
		"exclude":          p.Exclude,
		"path-style":       p.PathStyle,
		"dry-run":          p.DryRun,
		"content-type":     p.ContentType,
		"content-encoding": p.ContentEncoding,
		"cache-control":    p.CacheControl,
	}).Info("Attempting to upload")

	matches, err := matches(p.Source, p.Exclude)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Could not match files")
		return err
	}

	for _, match := range matches {

		stat, err := os.Stat(match)
		if err != nil {
			continue // should never happen
		}

		// skip directories
		if stat.IsDir() {
			continue
		}

		target := filepath.Join(p.Target, strings.TrimPrefix(match, p.StripPrefix))
		if !strings.HasPrefix(target, "/") {
			target = "/" + target
		}

		contentType := getContentType(match, p.ContentType)
		contentEncoding := globMatch(match, p.ContentEncoding)

		// log file for debug purposes.
		log.WithFields(log.Fields{
			"name":             match,
			"bucket":           p.Bucket,
			"target":           target,
			"cache-control":    p.CacheControl,
			"content-type":     contentType,
			"content-encoding": contentEncoding,
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
			Body:        f,
			Bucket:      &(p.Bucket),
			Key:         &target,
			ACL:         &(p.Access),
			ContentType: &contentType,
		}

		if p.Encryption != "" {
			putObjectInput.ServerSideEncryption = &(p.Encryption)
		}

		if p.CacheControl != "" {
			putObjectInput.CacheControl = &(p.CacheControl)
		}

		if contentType != "" {
			putObjectInput.ContentType = aws.String(contentType)
		}

		if contentEncoding != "" {
			putObjectInput.ContentEncoding = aws.String(contentEncoding)
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

// getContentType is a helper function that returns the content type for the file
// based on a list of glob patterns. If no match is found, it returns the type
// based on the extension. If the file extension is unknown "application/octet-stream"
// is returned.
func getContentType(path string, patterns map[string]string) string {
	typ := globMatch(path, patterns)

	// amazon S3 has pretty crappy default content-type headers so this pluign
	// attempts to provide a proper content-type in case it is not set by the user.
	if typ == "" {
		typ = mime.TypeByExtension(filepath.Ext(path))
	}

	if typ == "" {
		typ = "application/octet-stream"
	}

	return typ
}

// globMatch is a helper function that iterates map of glob patterns
// and returns the value of that map once it finds a pattern that matches
// the given string.
func globMatch(path string, patterns map[string]string) string {
	for pattern := range patterns {
		if glob.Glob(pattern, path) {
			return patterns[pattern]
		}
	}
	return ""
}
