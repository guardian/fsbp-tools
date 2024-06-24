# bucket-blocker

## How to run:

Bucket blocker takes up to 3 flags:

- **profile**: _Required._ The profile to use when connecting to AWS.

- **region**: _Required._ The region where the bucket is located.

- **dry-run**: _Optional._ Defaults to true, meaning no operation will be performed.

- **exclusions**: _Optional._ Comma-delimited list of buckets to exclude from blocking.

- **max**: _Optional._ The maximum number of buckets to block. Between 1 and 100. Defaults to 100, which is the maximum number of buckets that can exist in an AWS account.

Currently, there isn't a process to build the binary. You can run the application using the following command from the root of the repository

```bash
go run main.go \
-profile=<profile_name> \
-region=<region>
```
