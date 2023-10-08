package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "0.0.0"
	build   = "0"
)

func loadEnvVar() {
	fileName := os.Getenv("ENV_FILE")
	if fileName == "" {
		fileName = os.Getenv("PLUGIN_ENV-FILE")
	}

	if fileName != "" {
		err := godotenv.Overload(fileName)
		log.Info(fmt.Sprintf("Successfully loaded/overloaded environment variables from %s", fileName))
		if err != nil {
			log.Error(fmt.Sprintf("Error reading env file %s - %v", fileName, err))
		}
	}
}

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
			EnvVar: "PLUGIN_ACCESS-KEY,AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "PLUGIN_SECRET-KEY,AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "assume-role",
			Usage:  "aws iam role to assume",
			EnvVar: "PLUGIN_ASSUME-ROLE,ASSUME_ROLE",
		},
		cli.StringFlag{
			Name:   "assume-role-session-name",
			Usage:  "aws iam role session name to assume",
			Value:  "drone-s3",
			EnvVar: "PLUGIN_ASSUME-ROLE-SESSION-NAME,ASSUME_ROLE_SESSION_NAME",
		},
		cli.StringFlag{
			Name:   "user-role-arn",
			Usage:  "AWS user role",
			EnvVar: "PLUGIN_USER-ROLE-ARN,AWS_USER_ROLE_ARN",
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
			EnvVar: "PLUGIN_STRIP-PREFIX",
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
			EnvVar: "PLUGIN_DRY-RUN",
		},
		cli.BoolFlag{
			Name:   "path-style",
			Usage:  "use path style for bucket paths",
			EnvVar: "PLUGIN_PATH-STYLE",
		},
		cli.GenericFlag{
			Name:   "content-type",
			Usage:  "set content type header for uploaded objects",
			EnvVar: "PLUGIN_CONTENT-TYPE",
			Value:  &StringMapFlag{},
		},
		cli.GenericFlag{
			Name:   "content-encoding",
			Usage:  "set content encoding header for uploaded objects",
			EnvVar: "PLUGIN_CONTENT-ENCODING",
			Value:  &StringMapFlag{},
		},
		cli.GenericFlag{
			Name:   "cache-control",
			Usage:  "set cache-control header for uploaded objects",
			EnvVar: "PLUGIN_CACHE-CONTROL",
			Value:  &StringMapFlag{},
		},
		cli.StringFlag{
			Name:   "storage-class",
			Usage:  "set storage class to choose the best backend",
			EnvVar: "PLUGIN_STORAGE-CLASS",
		},
		cli.StringFlag{
			Name:   "env-file",
			Usage:  "source env file",
			EnvVar: "PLUGIN_ENV-FILE,ENV_FILE",
		},
		cli.StringFlag{
			Name:   "external-id",
			Usage:  "external ID to use when assuming role",
			EnvVar: "PLUGIN_EXTERNAL-ID",
		},
	}

	loadEnvVar()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{
		Endpoint:              c.String("endpoint"),
		Key:                   c.String("access-key"),
		Secret:                c.String("secret-key"),
		AssumeRole:            c.String("assume-role"),
		AssumeRoleSessionName: c.String("assume-role-session-name"),
		Bucket:                c.String("bucket"),
		UserRoleArn:           c.String("user-role-arn"),
		Region:                c.String("region"),
		Access:                c.String("acl"),
		Source:                c.String("source"),
		Target:                c.String("target"),
		StripPrefix:           c.String("strip-prefix"),
		Exclude:               c.StringSlice("exclude"),
		Encryption:            c.String("encryption"),
		ContentType:           c.Generic("content-type").(*StringMapFlag).Get(),
		ContentEncoding:       c.Generic("content-encoding").(*StringMapFlag).Get(),
		CacheControl:          c.Generic("cache-control").(*StringMapFlag).Get(),
		StorageClass:          c.String("storage-class"),
		PathStyle:             c.Bool("path-style"),
		DryRun:                c.Bool("dry-run"),
		ExternalID:            c.String("external-id"),
	}

	return plugin.Exec()
}
