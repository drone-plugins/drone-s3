package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-zglob"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// Plugin defines the S3 plugin parameters.
type Plugin struct {
	Key    string
	Secret string
	Bucket string

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

	// Recursive uploads
	Recursive bool

	// Exclude files matching this pattern.
	Exclude []string

	// Dry run without uploading/
	DryRun bool
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	auth, err := aws.GetAuth(p.Key, p.Secret)
	if err != nil {
		return err
	}
	region := s3.New(auth, aws.Regions[p.Region])
	bucket := region.Bucket(p.Bucket)

	matches, err := matches(p.Source, p.Exclude)
	if err != nil {
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

		access := s3.ACL(p.Access)
		target := filepath.Join(p.Target, match)
		if !strings.HasPrefix(target, "/") {
			target = "/" + target
		}

		// amazon S3 has pretty crappy default content-type headers so this pluign
		// attempts to provide a proper content-type.
		headers := map[string][]string{
			"Content-Type": contentType(match),
		}

		// log file for debug purposes.
		log.Printf("upload %q to %q at %q\n", match, p.Bucket, target)

		// when executing a dry-run we exit because we don't actually want to
		// upload the file to S3.
		if p.DryRun {
			continue
		}

		f, err := os.Open(match)
		if err != nil {
			return err
		}
		defer f.Close()

		err = bucket.PutReaderHeader(target, f, stat.Size(), headers, access)
		if err != nil {
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
func contentType(path string) []string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)
	if typ == "" {
		typ = "application/octet-stream"
	}
	return []string{typ}
}

// func (p *Plugin) execOld() error {
//
// 	cmd := p.toCommand()
// 	cmd.Env = os.Environ()
// 	if len(p.Key) > 0 {
// 		cmd.Env = append(cmd.Env, "AWS_ACCESS_KEY_ID="+p.Key)
// 	}
// 	if len(p.Secret) > 0 {
// 		cmd.Env = append(cmd.Env, "AWS_SECRET_ACCESS_KEY="+p.Secret)
// 	}
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	trace(cmd)
//
// 	// run the command and exit if failed.
// 	return cmd.Run()
// }

// toCommand is a helper function that returns the command and arguments to
// upload to aws from the command line.
// func (p *Plugin) toCommand() *exec.Cmd {
//
// 	// remote path S3 uri
// 	path := fmt.Sprintf("s3://%s/%s", p.Bucket, p.Target)
//
// 	// command line args
// 	args := []string{
// 		"s3",
// 		"cp",
// 		p.Source,
// 		path,
// 		"--recursive",
// 		"--acl",
// 		p.Access,
// 		"--region",
// 		p.Region,
// 	}
//
// 	// if not recursive, remove from the above arguments.
// 	if !p.Recursive {
// 		args = append(args[:4], args[4+1:]...)
// 	}
//
// 	for i := 0; i < len(p.Include); i++ {
// 		args = append(args, "--include", p.Include[i])
// 	}
//
// 	for i := 0; i < len(p.Exclude); i++ {
// 		args = append(args, "--exclude", p.Exclude[i])
// 	}
//
// 	return exec.Command("aws", args...)
// }

// trace writes each command to standard error (preceded by a ‘+ ’) before it
// is executed. Used for debugging your build.
// func trace(cmd *exec.Cmd) {
// 	fmt.Println("+", strings.Join(cmd.Args, " "))
// }
