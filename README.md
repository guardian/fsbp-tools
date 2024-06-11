# bucket-blocker

## How to run:

Bucket blocker takes 3 flags:

- **bucket**: The name of the bucket to block

- **profile**: The profile to use when connecting to AWS.

- **region**: The region where the bucket is located.

Currently, there isn't a process to build the binary. You can run the application using the following command from the root of the repository

```bash
go run src/bucketblocker/main.go \
-bucket <bucket_name> \
-profile <profile_name> \
-region <region>
```
