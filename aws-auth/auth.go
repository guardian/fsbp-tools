package awsauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	shTypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func validateCredentials(ctx context.Context, stsClient *sts.Client, profile string) (*sts.GetCallerIdentityOutput, error) {
	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return resp, errors.New("Could not find valid credentials for profile: " + profile)
	}
	return resp, nil
}

func LoadDefaultConfig(ctx context.Context, profile string, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile), config.WithDefaultRegion(region))
	if err != nil {
		fmt.Println("Error loading configuration")
		return cfg, err
	}

	stsClient := sts.NewFromConfig(cfg)
	_, err = validateCredentials(ctx, stsClient, profile)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func ReturnFindings(ctx context.Context, securityHubClient *securityhub.Client, controlId string, maxResults int32) (*securityhub.GetFindingsOutput, error) {

	complianceStatus := "PASSED"
	recordState := "ACTIVE"

	fmt.Printf("Retrieving Security Hub control failures for %s\n", controlId)
	findings, err := securityHubClient.GetFindings(ctx, &securityhub.GetFindingsInput{
		MaxResults: &maxResults,
		Filters: &shTypes.AwsSecurityFindingFilters{
			ComplianceSecurityControlId: []shTypes.StringFilter{{
				Value:      &controlId,
				Comparison: shTypes.StringFilterComparisonEquals,
			}},
			ComplianceStatus: []shTypes.StringFilter{{
				Value:      &complianceStatus,
				Comparison: shTypes.StringFilterComparisonNotEquals,
			}},
			RecordState: []shTypes.StringFilter{{
				Value:      &recordState,
				Comparison: shTypes.StringFilterComparisonEquals,
			}},
		},
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}
