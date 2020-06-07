# s3uploader

Tool for easy uploads to s3 without the need of installing AWS CLI

## Why? And how?
Uploading to s3 is probably easier done than explained, however in many cases you have to install the aws cli, even when it's a python script that is doing the uploading you still have to pip install the necessary modules. What this tool represents is the exact opposite of that. **Copy the binary and run it**. If you don't want to build it (maybe installing go is too much for you?) there's a simple Dockerfile in the repo which you can also use to trigger the upload from a docker container.

## Limitations

You only get to upload to a certain s3 bucket (and folder within) one file at a time, that's it. No multiparts (although it's performance optimized to handle large files, such as DB backups and encrypted dumps)

## Download

- **Install from source**
```shell
go get github.com/aws/aws-sdk-go/...
go build -a -o s3uploader
```

- **Run it from docker**
```shell
docker build . -t local/s3uploader:0.0.1
docker run --rm -it \
  -e BUCKET="my-s3-bucket" \
  -e ACCESSKEY="my-access_key" \
  -e SECRET="my-secret-key" \
  local/s3uploader:latest /tmp/file
```

## Usage

```
Please define the arguments as environment variables! Ex:
ACCESSKEY=AWSACCESSKEY SECRETKEY=AWSSECRETKEY BUCKET=BUCKETNAME REGION=REGION.... go run main.go /path/to/file
```


## Examples

### Part of a simple cront script for daily backup and uploading
```bash
#!/bin/bash
...
... create backup ...
...
BACKUP_NAME=my_backup_$(date -u +%Y-%m-%dT%H-%M-%S)
export $(grep -v '^#' .env | xargs) && /opt/sbin/s3uploader BACKUP_NAME
```
