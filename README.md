# FSBP tools

This repository contains a collection of tools that help us maintain compliance with AWS'
[Foundational Security Best Practices](https://docs.aws.amazon.com/securityhub/latest/userguide/fsbp-standard.html)
(FSBP).

Not all of the FSBP controls are covered by these tools - we aim to provide tools for common
control failures that are easy to remediate automatically.

## Tools

- [bucket-blocker](./bucket-blocker/README.md) - Adds policies to S3 buckets blocking public access.
- [ingress-inquisition](./ingress-inquisition/README.md) - Removes ingress and egress rules from default security groups in VPCs.

## Local development

When committing your changes, please use the
[conventional commit](https://www.conventionalcommits.org/en/v1.0.0/#summary)
format. This will allow us to automatically generate a changelog and correctly
version the application when it is released.
