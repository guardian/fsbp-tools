# bucket-blocker

## How to run:

Bucket blocker takes up to 3 flags:

- **bucket**: _Required argument_.The name of the bucket to block

- **profile**: _Optional argument_. The profile to use when connecting to AWS. Defaults to `default`

- **region**: _Optional argument_. The region where the bucket is located. Defaults to `eu-west-1`

Currently, there isn't a process to build the binary. You can run the application using the following command from the root of the repository

```bash
go run src/bucketblocker/main.go \
-bucket <bucket_name> \
-profile <profile_name> \
-region <region>
```
