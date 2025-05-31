package common

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

func findingsInput(controlId string, maxResults int32, nextToken *string) *securityhub.GetFindingsInput {
	return &securityhub.GetFindingsInput{
		MaxResults: &maxResults,
		NextToken:  nextToken,
		Filters: &shTypes.AwsSecurityFindingFilters{
			ComplianceSecurityControlId: []shTypes.StringFilter{{
				Value:      &controlId,
				Comparison: shTypes.StringFilterComparisonEquals,
			}},
			ComplianceStatus: []shTypes.StringFilter{{
				Value:      aws.String("PASSED"),
				Comparison: shTypes.StringFilterComparisonNotEquals,
			}},
			RecordState: []shTypes.StringFilter{{
				Value:      aws.String("ACTIVE"),
				Comparison: shTypes.StringFilterComparisonEquals,
			}},
		},
	}
}

func ReturnFindings(ctx context.Context, securityHubClient *securityhub.Client, controlId string, maxResults int32) (*[]shTypes.AwsSecurityFinding, error) {

	fmt.Printf("Retrieving Security Hub control failures for %s\n", controlId)
	var page = 1 //TODO for debugging. Delete later
	findingsPage, err := securityHubClient.GetFindings(ctx, findingsInput(controlId, maxResults, nil))
	if err != nil {
		return nil, err
	}

	allFindings := []shTypes.AwsSecurityFinding{}
	allFindings = append(allFindings, findingsPage.Findings...)
	var nextToken = findingsPage.NextToken
	for nextToken != nil {
		page++
		fmt.Printf("Retrieving page %d of Security Hub control failures for %s\n", page, controlId)
		findingsPage, err := securityHubClient.GetFindings(ctx, findingsInput(controlId, maxResults, nextToken))
		if err != nil {
			return nil, err
		}
		allFindings = append(allFindings, findingsPage.Findings...)
		nextToken = findingsPage.NextToken
	}
	return &(allFindings), nil
}
