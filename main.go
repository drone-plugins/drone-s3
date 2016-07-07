package main

import (
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	_ "github.com/joho/godotenv/autoload"
)

var version string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "s3 artifact plugin"
	app.Usage = "s3 artifact plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "endpoint for the s3 connection",
			EnvVar: "PLUGIN_ENDPOINT",
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
			EnvVar: "PLUGIN_BUCKET",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_REGION",
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
		cli.BoolFlag{
			Name:   "recursive",
			Usage:  "upload files recursively",
			EnvVar: "PLUGIN_RECURSIVE",
		},
		cli.StringSliceFlag{
			Name:   "exclude",
			Usage:  "ignore files matching exclude pattern",
			EnvVar: "PLUGIN_EXCLUDE",
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
		cli.BoolTFlag{
			Name:   "yaml-verified",
			Usage:  "Ensure the yaml was signed",
			EnvVar: "DRONE_YAML_VERIFIED",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
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
		Recursive:    c.Bool("recursive"),
		Exclude:      c.StringSlice("exclude"),
		PathStyle:    c.Bool("path-style"),
		DryRun:       c.Bool("dry-run"),
		YamlVerified: c.BoolT("yaml-verified"),
	}

	// normalize the target URL
	if strings.HasPrefix(plugin.Target, "/") {
		plugin.Target = plugin.Target[1:]
	}

	return plugin.Exec()
}
