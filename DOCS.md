Use the S3 plugin to upload files and build artifacts to an S3 bucket. The following parameters are used to configure this plugin:

* **endpoint** - custom endpoint URL (optional, to use a S3 compatible non-Amazon service)
* **access_key** - amazon key (optional)
* **secret_key** - amazon secret key (optional)
* **bucket** - bucket name
* **region** - bucket region (`us-east-1`, `eu-west-1`, etc)
* **acl** - access to files that are uploaded (`private`, `public-read`, etc)
* **source** - source location of the files, using a glob matching pattern
* **target** - target location of files in the bucket
* **exclude** - glob exclusion patterns
* **path_style** - whether path style URLs should be used (true for minio, false for aws)
* **compress** - prior to upload, compress files and use gzip content-encoding


The following is a sample S3 configuration in your .drone.yml file:

```yaml
publish:
  s3:
    acl: public-read
    region: "us-east-1"
    bucket: "my-bucket-name"
    access_key: "970d28f4dd477bc184fbd10b376de753"
    secret_key: "9c5785d3ece6a9cdefa42eb99b58986f9095ff1c"
    source: public/**/*
    target: /target/location
    exclude:
      - **/*.xml
    compress: false
```
