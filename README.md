# nas-backuper

CLI that compresses a folder and sends it to a AmazonS3 bucket.

Buffers the entire compressed file to memory before sending to S3.

## Usage

1. Download the binary
2. Make sure to set your `~/.aws/credentials` file
3. Run:
```console
./backuper ./relative-path-to-folder s3-bucket-name
```

## Optional flags:

### --name 
The name to be used to the file being uploaded to the bucket. Defaults to a generic name containing the current date of the program execution.

### --region
The region of the S3 Bucket. Defaults to us-east-1.

## TODO
[] Allow for temporarily writing a file in disk and then streaming it to S3. Will require the usage of a flag passing the location of where to write the file. This is useful to handle large uploads.

[] Add support to compress files as well

[] Create binary download on github

[] Dropbox provider

[] DigitalOcean provider

[] GoogleCloud provider
