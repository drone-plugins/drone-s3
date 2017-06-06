Use this plugin to upload files and build artifacts to an S3 bucket or a Minio
bucket.

## Config

The following parameters are used to configure the plugin:

* **endpoint** - custom endpoint URL (optional, to use a S3 compatible non-Amazon service)
* **access_key** - amazon key (optional)
* **secret_key** - amazon secret key (optional)
* **bucket** - bucket name
* **region** - bucket region (`us-east-1`, `eu-west-1`, etc)
* **acl** - access to files that are uploaded (`private`, `public-read`, etc)
* **source** - source location of the files, using a glob matching pattern
* **target** - target location of files in the bucket
* **encryption** - if provided, use server-side encryption (`AES256`, `aws:kms`, etc)
* **strip_prefix** - strip the prefix from source path
* **exclude** - glob exclusion patterns
* **path_style** - whether path style URLs should be used (true for minio, false for aws)
* **env_file** - environment file from which to load environment variables

The following secret values can be set to configure the plugin.

* **AWS_ACCESS_KEY_ID** - corresponds to **webhook**
* **AWS_SECRET_ACCESS_KEY** - corresponds to **webhook**
* **S3_BUCKET** - corresponds to **webhook**
* **S3_REGION** - corresponds to **webhook**
* **S3_ENDPOINT** - corresponds to **webhook**

It is highly recommended to put the **AWS_ACCESS_KEY_ID** and
**AWS_SECRET_ACCESS_KEY** into a secret so it is not exposed to users. This can
be done using the drone-cli.

```bash
drone secret add --image=plugins/s3 \
    octocat/hello-world AWS_ACCESS_KEY_ID <YOUR_ACCESS_KEY_ID>

drone secret add --image=plugins/s3 \
    octocat/hello-world AWS_SECRET_ACCESS_KEY <YOUR_SECRET_ACCESS_KEY>
```

Then sign the YAML file after all secrets are added.

```bash
drone sign octocat/hello-world
```

See [secrets](http://readme.drone.io/0.5/usage/secrets/) for additional
information on secrets

## Example

Common example to upload to S3:

```yaml
pipeline:
  s3:
    image: plugins/s3
    acl: public-read
    region: "us-east-1"
    bucket: "my-bucket-name"
    access_key: "970d28f4dd477bc184fbd10b376de753"
    secret_key: "9c5785d3ece6a9cdefa42eb99b58986f9095ff1c"
    source: public/**/*
    strip_prefix: public/
    target: /target/location
    encryption: AES256
    exclude:
      - **/*.xml
```
