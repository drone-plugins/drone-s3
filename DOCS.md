Use the S3 plugin to upload files and build artifacts to an S3 bucket. The following parameters are used to configure this plugin:

* **access_key** - amazon key (optional)
* **secret_key** - amazon secret key (optional)
* **use_iam** - use the host's AWS IAM role (optional)
* **bucket** - bucket name
* **region** - bucket region (`us-east-1`, `eu-west-1`, etc)
* **acl** - access to files that are uploaded (`private`, `public-read`, etc)
* **source** - location of files to upload
* **target** - target location of files in your S3 bucket
* **recursive** - if true, recursively upload files

The following is a sample S3 configuration that will use the host's IAM role in your .drone.yml file:

```yaml
publish:
  s3:
    acl: public-read
    region: "us-east-1"
    bucket: "my-bucket-name"
    use_iam: true
    source: files/to/archive
    target: /target/location
    recursive: true
```

Or you can specify an access_key and secret_key:

```yaml
publish:
  s3:
    acl: public-read
    region: "us-east-1"
    bucket: "my-bucket-name"
    access_key: "970d28f4dd477bc184fbd10b376de753"
    secret_key: "9c5785d3ece6a9cdefa42eb99b58986f9095ff1c"
    source: files/to/archive
    target: /target/location
    recursive: true
```
