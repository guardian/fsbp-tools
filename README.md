# bucket-blocker

## How to run:

Bucket blocker takes up to 3 flags:

- **profile**: _Required._ The profile to use when connecting to AWS.

- **region**: _Required._ The region where the bucket is located.

- **dry-run**: _Optional._ Defaults to true, meaning no operation will be performed.

Currently, there isn't a process to build the binary. You can run the application using the following command from the root of the repository

```bash
go run main.go \
-profile <profile_name> \
-region <region>
```
