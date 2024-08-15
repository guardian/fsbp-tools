module github.com/guardian/fsbp-tools/ingress-inquisition

go 1.22.1

replace github.com/guardian/fsbp-tools/aws-auth => ../aws-auth

require (
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.49.2
	github.com/guardian/fsbp-tools/aws-auth v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.175.0
)

require (
	github.com/aws/aws-sdk-go-v2 v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.18 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect

	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.12 // indirect
	github.com/aws/smithy-go v1.20.3 // indirect
)
