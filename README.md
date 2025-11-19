# go-cfr2
Tool to operate Cloudflare R2

## Build
```bash
go mod tidy && go build -o build/go-cfr2
```

## Setup
go-cfr2 read config file from $HOME/.local/cfg/cfr2.toml，example:
```cfr2.toml
AccountID = 'Your cloudflare r2 AccountID'
AccessKeyID = 'Your cloudflare r2 AccessKeyID'
SecretAccessKey = 'Your cloudflare r2 SecretAccessKey'
DefaultBucket = 'Your default bucket'
```

## Usage
```bash
Usage: go-cfr2 <command> [flags]

Commands:
  list      List all objects in the default R2 bucket
            Flags:
              -b, --bucket <name> Specify the R2 bucket name (optional)
                                   (Defaults to DefaultBucket in config)

 download  Download an object from the default R2 bucket
            Flags:
              -b, --bucket <name> Specify the R2 bucket name (optional)
                                   (Defaults to DefaultBucket in config)
              -k, --key <key>      Specify the object key to download (required)
              -o, --output <path> Specify the output file path or directory (optional)
                                   (Defaults to current directory, filename from key)

  upload    Upload a file to the default R2 bucket
            Flags:
              -b, --bucket <name> Specify the R2 bucket name (optional)
                                   (Defaults to DefaultBucket in config)
              -f, --file <path>    Specify the local file to upload (required)
              -k, --key <key>      Specify the object key for the uploaded file (required)

  delete    Delete an object from the default R2 bucket
            Flags:
              -b, --bucket <name> Specify the R2 bucket name (optional)
                                   (Defaults to DefaultBucket in config)
              -k, --key <key>      Specify the object key to delete (required)

 rename    Rename an object in the default R2 bucket
            Flags:
              -b, --bucket <name> Specify the R2 bucket name (optional)
                                   (Defaults to DefaultBucket in config)
              -o, --old-key <key>   Specify the old object key to rename (required)
              -n, --new-key <key>   Specify the new object key (required)

 presign   Generate a presigned URL for an object with default 24-hour expiration
            Flags:
              -b, --bucket <name> Specify the R2 bucket name (optional)
                                   (Defaults to DefaultBucket in config)
              -k, --key <key>      Specify the object key (required)
              -e, --expiry <hours> Specify the URL expiry time in hours (optional)
                                   (Defaults to 24 hours)
```