package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "0.0.0"
	build   = "0"
)

func main() {
	app := cli.NewApp()
	app.Name = "s3 plugin"
	app.Usage = "s3 plugin"
	app.Version = fmt.Sprintf("%s+%s", version, build)
	app.Action = run
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
		cli.StringFlag{
			Name:   "cache-control",
			Usage:  "set cache-control header for uploaded objects",
			EnvVar: "PLUGIN_CACHE_CONTROL",
		},
		cli.StringFlag{
			Name:  "env-file",
			Usage: "source env file",
		},
		cli.StringFlag{
			Name:   "target-remove",
			Usage:  "regex for which files to remove from target",
			EnvVar: "PLUGIN_TARGET_REMOVE",
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
		Endpoint:     c.String("endpoint"),
		Key:          c.String("access-key"),
		Secret:       c.String("secret-key"),
		Bucket:       c.String("bucket"),
		Region:       c.String("region"),
		Access:       c.String("acl"),
		Source:       c.String("source"),
		Target:       c.String("target"),
		StripPrefix:  c.String("strip-prefix"),
		Exclude:      c.StringSlice("exclude"),
		Encryption:   c.String("encryption"),
		CacheControl: c.String("cache-control"),
		PathStyle:    c.Bool("path-style"),
		DryRun:       c.Bool("dry-run"),
		TargetRemove: c.String("target-remove"),
	}

	return plugin.Exec()
}
