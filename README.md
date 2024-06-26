# bucket-blocker

## Command line options:

Bucket blocker takes up to 3 flags:

- **profile**: _Required._ The profile to use when connecting to AWS.

- **region**: _Required._ The region where the bucket is located.

- **dry-run**: _Optional._ Defaults to true, meaning no operation will be performed.

- **exclusions**: _Optional._ Comma-delimited list of buckets to exclude from blocking.

- **max**: _Optional._ The maximum number of buckets to block. Between 1 and 100. Defaults to 100, which is the maximum number of buckets that can exist in an AWS account.

You will also need credentials for the relevant AWS account from Janus.

## Running the binary

This app runs as a binary executable on both Intel and Apple Silicon architectures. To run the binary, you can either build it from source or download it from the releases page. Apple Silicon users should download the `darwin-arm64` binary, while Intel users should download the `darwin-amd64` binary.

1. **Recommended**: `wget https://github.com/guardian/bucket-blocker/releases/download/v<VERSION>/bucket-blocker-<ARCHITECTURE>`
   into a directory of your choice. You can see a list of releases
   [here](https://github.com/guardian/bucket-blocker/releases). `chmod +x` the binary.
2. Clone the repository and build the binary using `./build.sh`
3. Download the binary from the releases page and `chmod +x` the binary

Once you have the binary, you can run it, passing in the desired flags, for example:
`./bucket-blocker-darwin-arm64 --profile deployTools --region eu-west-1`

## Local development

<!-- TODO enforce conventional commits via GHA-->

When committing your changes, please use the [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/#summary) format. This will allow us to automatically generate a changelog and correctly version the application when it is released.

While developing locally, you can test the application using the following command from the root of the repository,
without needing to build the binary:

```bash
go run main.go -profile=<PROFILE> -region=<REGION> [OPTIONAL_FLAGS]
```
