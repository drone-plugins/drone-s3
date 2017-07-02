package main

import (
	"mime"
	"os"
	"path/filepath"
	"strings"

	"errors"
	log "github.com/Sirupsen/logrus"
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

	// Recursive uploads
	Recursive bool

	YamlVerified bool

	// Exclude files matching this pattern.
	Exclude []string

	// Use path style instead of domain style.
	//
	// Should be true for minio and false for AWS.
	PathStyle bool
	// Dry run without uploading/
	DryRun bool

	// Content-Type
	ContentTypes map[string]string
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

		// Attempt to find the correct content-type
		content := contentType(p, match)

		// log file for debug purposes.
		log.WithFields(log.Fields{
			"name":         match,
			"bucket":       p.Bucket,
			"target":       target,
			"content-type": content,
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
			ContentType: &content,
		}

		if p.Encryption != "" {
			putObjectInput.ServerSideEncryption = &(p.Encryption)
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

// contentType is a helper function that returns the content type for the file
// based on extension. If the file extension is unknown application/octet-stream
// is returned.
func contentType(p *Plugin, path string) string {
	ext := filepath.Ext(path)
	var typ string

	if ext != "" {
		typ = p.ContentTypes[ext]
		if typ != "" {
			return typ
		}

		typ = mime.TypeByExtension(ext)
		if typ != "" {
			return typ
		}
	}

	typ = p.ContentTypes["default"]
	if typ != "" {
		return typ
	}

	typ = "application/octet-stream"
	return typ
}
