package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var build = "0" // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "s3 plugin"
	app.Usage = "s3 plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "endpoint for the s3 connection",
			EnvVar: "PLUGIN_ENDPOINT,S3_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "bucket",
			Usage:  "aws bucket",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_BUCKET,S3_BUCKET",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_REGION,S3_REGION",
		},
		cli.StringFlag{
			Name:   "acl",
			Usage:  "upload files with acl",
			Value:  "private",
			EnvVar: "PLUGIN_ACL",
		},
		cli.StringFlag{
			Name:   "source",
			Usage:  "upload files from source folder",
			EnvVar: "PLUGIN_SOURCE",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "upload files to target folder",
			EnvVar: "PLUGIN_TARGET",
		},
		cli.StringFlag{
			Name:   "strip-prefix",
			Usage:  "strip the prefix from the target",
			EnvVar: "PLUGIN_STRIP_PREFIX",
		},
		cli.StringSliceFlag{
			Name:   "exclude",
			Usage:  "ignore files matching exclude pattern",
			EnvVar: "PLUGIN_EXCLUDE",
		},
		cli.StringFlag{
			Name:   "encryption",
			Usage:  "server-side encryption algorithm, defaults to none",
			EnvVar: "PLUGIN_ENCRYPTION",
		},
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run for debug purposes",
			EnvVar: "PLUGIN_DRY_RUN",
		},
		cli.BoolFlag{
			Name:   "path-style",
			Usage:  "use path style for bucket paths",
			EnvVar: "PLUGIN_PATH_STYLE",
		},
		cli.GenericFlag{
			Name:   "content-type",
			Usage:  "content-type settings for uploads",
			EnvVar: "PLUGIN_CONTENT_TYPE",
			Value:  &StringMapFlag{},
		},
		cli.GenericFlag{
			Name:   "content-encoding",
			Usage:  "content-encoding settings for uploads",
			EnvVar: "PLUGIN_CONTENT_ENCODING",
			Value:  &StringMapFlag{},
		},
		cli.GenericFlag{
			Name:   "cache-control",
			Usage:  "cache-control settings for uploads",
			EnvVar: "PLUGIN_CACHE_CONTROL",
			Value:  &StringMapFlag{},
		},
		cli.GenericFlag{
			Name:   "metadata",
			Usage:  "additional metadata for uploads",
			EnvVar: "PLUGIN_METADATA",
			Value:  &DeepStringMapFlag{},
		},
		cli.StringFlag{
			Name:  "env-file",
			Usage: "source env file",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := Plugin{
		Endpoint:        c.String("endpoint"),
		Key:             c.String("access-key"),
		Secret:          c.String("secret-key"),
		Bucket:          c.String("bucket"),
		Region:          c.String("region"),
		Access:          c.String("acl"),
		Source:          c.String("source"),
		Target:          c.String("target"),
		StripPrefix:     c.String("strip-prefix"),
		Exclude:         c.StringSlice("exclude"),
		Encryption:      c.String("encryption"),
		PathStyle:       c.Bool("path-style"),
		CacheControl:    c.Generic("cache-control").(*StringMapFlag).Get(),
		ContentType:     c.Generic("content-type").(*StringMapFlag).Get(),
		ContentEncoding: c.Generic("content-encoding").(*StringMapFlag).Get(),
		Metadata:        c.Generic("metadata").(*DeepStringMapFlag).Get(),
		DryRun:          c.Bool("dry-run"),
	}

	return plugin.Exec()
}
