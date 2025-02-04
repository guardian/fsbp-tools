# FSBP tools

## What is this thing?

fsbp-fix is a tool that searches for and automatically remediates auto-fixable violations of the [AWS FSBP standard](https://docs.aws.amazon.com/securityhub/latest/userguide/fsbp-standard.html). Currently, just [S3.8](https://docs.aws.amazon.com/securityhub/latest/userguide/s3-controls.html#s3-8), which states that all buckets should have individual configurations blocking public access, and [EC2.2](https://docs.aws.amazon.com/securityhub/latest/userguide/ec2-controls.html#ec2-2), which states that default security groups in VPCs should not allow any inbound or outbound traffic, are supported.

## Installation

```bash
brew tap guardian/homebrew-devtools && \
brew install fsbp-fix
```

## S3.8 - S3 general purpose buckets should block public access

### Usage

The minimal flags required to resolve S3.8 are as follows. This will execute in dry run mode.

```bash
fsbp-fix S3.8 -profile <PROFILE> -region <REGION> [OPTIONAL_FLAGS]
```

<details>
  <summary>Details</summary>
### Function

First, we find all the buckets that are breaking this rule. It skips over any that are in CloudFormation stacks (to avoid introducing stack drift), and then blocks public access to the remaining buckets.

```mermaid
flowchart TB
    stack[Is it part of a cloudformation stack]
    excl[Is it in a list of excluded \n buckets provided by the user?]
    block[Block public access to the bucket]
    ruleBreak[Does the bucket break S3.8?]
    break[Do nothing.]
    noAccess[No. Access already \n blocked]

    ruleBreak --> Yes --> stack --> No --> excl --> Nope --> block
    ruleBreak --> noAccess --> break
    stack --> Yeah --> break
    excl --> Yep --> break
```

There are a few extra features, controlled by flags, enumerated below.
</details>

<details>
    <summary>CLI options</summary>
fsbp-fix takes a subcommand and up to 3 flags:

- **profile**: _Required._ The profile to use when connecting to AWS.

- **region**: _Required._ The region you want to search in.

- **execute**: _Optional._ Takes no value. If present, it will ask the user to confirm, then block the buckets. If not, it will only print
  the buckets that would have been blocked.

- **exclusions**: _Optional._ Comma-delimited list of buckets to exclude from blocking.

- **max**: _Optional._ The maximum number of buckets to block. Between 1
  and 100. Defaults to 100, which is the maximum number of buckets that can
  exist in an AWS account.

You will also need credentials for the relevant AWS account from Janus.
</details>

<details>
    <summary>Local development</summary>
While developing locally, you can test the application using the following
command from the bucket-blocker subdirectory, without needing to build the binary:

```bash
go run main.go s3.8 -profile <PROFILE> -region <REGION> [OPTIONAL_FLAGS]
```

</details>

## EC2.2 - VPC default security groups should not allow inbound or outbound traffic

### Usage

The minimal flags required to resolve EC2.2 are as follows. This will execute in dry run mode.

```bash
fsbp-fix ec2.2 -profile <PROFILE> -region <REGION> [OPTIONAL_FLAGS]
```

<details>
  <summary>Details</summary>
AWS Security Hub Control [EC2.2](https://docs.aws.amazon.com/securityhub/latest/userguide/ec2-controls.html#ec2-2) states that default security groups in VPCs should not allow any inbound or outbound traffic. VPCs set up recently are compliant by default, but older VPCs are not.

The tool will search for relevant security groups that are not compliant with this control, and check to see if the security group is being used. If the group is not in use, it will remove the offending ingress/egress rules.

```mermaid
flowchart TB
    usage[Is the group in use?]
    block[Delete all security group rules]
    ruleBreak[Does the group break EC2.2?]
    break[Do nothing.]
    inUse[Yeah]

    ruleBreak --> Yes --> usage --> No --> block
    usage --> inUse --> break
    ruleBreak --> Nope --> break

```

</details>

<details>
    <summary>CLI options</summary>
ingress-inquisition takes the following flags:

- **profile**: _Required._ The profile to use when connecting to AWS.

- **region**: _Required._ The region you want to search in.

- **execute**: _Optional._ Takes no value. If present, it will ask the user to confirm, then delete the rules. Otherwise, it will just list the rules that would have been deleted.

</details>

<details>
    <summary>Local development</summary>
While developing locally, you can test the application using the following
command from the ingress-inquisitor subdirectory, without needing to build the binary:

```bash
go run main.go ec2.2 -profile <PROFILE> -region <REGION> [OPTIONAL_FLAGS]
```

</details>

<details>
    <summary>FAQ</summary>
### FAQ

#### How do we know if a security group is being used?

Security groups are associated with resources such as EC2 instances, databases, etc via an Elastic Network Interface (ENI). Ingress inquisition queries the AWS API to check all ENIs in the region, and if a security group is associated with an ENI, it is considered in use, and the rules will not be deleted.
</details>

## Local development

When committing your changes, please use the
[conventional commit](https://www.conventionalcommits.org/en/v1.0.0/#summary)
format. This will allow us to automatically generate a changelog and correctly
version the application when it is released.

## Releasing to brew

Creating a new release of the application on brew, is currently a manual
process. You will need to update the version, urls, and SHAs in
[homebrew-devtools](https://github.com/guardian/homebrew-devtools). The SHAs are generated by running `shasum -a 256 <filename>` on the binary, or by checking the annotations on the release step in the GitHub Actions workflow.
