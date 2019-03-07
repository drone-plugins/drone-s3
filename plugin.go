package main

import (
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bmatcuk/doublestar"
	"github.com/mattn/go-zglob"
	log "github.com/sirupsen/logrus"
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
	// Glob for which files to remove from target
	TargetRemove string
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

	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	} else {
		log.Warn("AWS Key and/or Secret not provided (falling back to ec2 instance profile)")
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

	if len(p.TargetRemove) != 0 {

		log.WithFields(log.Fields{
			"glob": p.TargetRemove,
		}).Info("Deleting files according to glob")

		log.Info("Listing files in bucket")
		listInput := &s3.ListObjectsInput{
			Bucket: &p.Bucket,
		}

		s3Objects, err := client.ListObjects(listInput)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error listing objects from bucket")
			return err
		}

		var toRemove []string
		for _, object := range s3Objects.Contents {
			filename := object.Key

			globmatch, err := doublestar.PathMatch(p.TargetRemove, *filename)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"glob":  p.TargetRemove,
				}).Error("Error with provided glob")
				return err
			}

			if globmatch {
				toRemove = append(toRemove, *filename)
			}
		}

		if len(toRemove) > 0 {
			log.WithFields(log.Fields{
				"num_files": len(toRemove),
			}).Info("Deleting files from bucket")

			var removeIdentifiers []*s3.ObjectIdentifier
			for _, key := range toRemove {
				id := s3.ObjectIdentifier{
					Key: aws.String(key),
				}
				removeIdentifiers = append(removeIdentifiers, &id)
			}

			deleteInput := &s3.DeleteObjectsInput{
				Bucket: &p.Bucket,
				Delete: &s3.Delete{
					Objects: removeIdentifiers,
					Quiet:   aws.Bool(false),
				},
			}

			// when executing a dry-run we skip this step because we don't actually
			// want to remove files from S3.
			if !p.DryRun {
				log.WithFields(log.Fields{
					"num_files": len(removeIdentifiers),
				}).Info("Attempting to delete files")
				_, err := client.DeleteObjects(deleteInput)

				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Error deleting objects from S3")
					return err
				}
			}
		}
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

		// amazon S3 has pretty crappy default content-type headers so this pluign
		// attempts to provide a proper content-type.
		content := contentType(match)

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

		if p.CacheControl != "" {
			putObjectInput.CacheControl = &(p.CacheControl)
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
func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)
	if typ == "" {
		typ = "application/octet-stream"
	}
	return typ
}
