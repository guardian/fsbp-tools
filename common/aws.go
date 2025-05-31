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

// Generic paginator for AWS SDK v2
type pageFetcherFunc[T any] func(nextToken *string) (items []T, next *string, err error)

func Paginate[T any](fetch pageFetcherFunc[T]) ([]T, error) { //TODO test this
	var allItems []T
	var nextToken *string
	var page int32 = 0 //For debugging. Delete if not needed
	for {
		page++
		fmt.Printf("Fetching page %d...\n", page)
		items, next, err := fetch(nextToken)
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, items...)
		if next == nil {
			break
		}
		nextToken = next
	}
	return allItems, nil
}

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

func GetAccountId(ctx context.Context, profile string, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return "", fmt.Errorf("error loading config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("error getting caller identity: %w", err)
	}

	return *resp.Account, nil
}

func findingsInput(controlId string, maxResults int32, nextToken *string, accountId string, region string) *securityhub.GetFindingsInput {
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
			AwsAccountId: []shTypes.StringFilter{{
				Value:      &accountId,
				Comparison: shTypes.StringFilterComparisonEquals,
			}},
			Region: []shTypes.StringFilter{{
				Value:      &region,
				Comparison: shTypes.StringFilterComparisonEquals,
			}},
		},
	}
}
func ReturnFindings(ctx context.Context, securityHubClient *securityhub.Client, controlId string, maxResults int32, accountId string, region string) ([]shTypes.AwsSecurityFinding, error) {

	fmt.Printf("Retrieving Security Hub control failures for %s\n", controlId)

	allFindings, err := Paginate(func(nextToken *string) ([]shTypes.AwsSecurityFinding, *string, error) {
		input := findingsInput(controlId, maxResults, nextToken, accountId, region)
		resp, err := securityHubClient.GetFindings(ctx, input)
		if err != nil {
			return nil, nil, err
		}

		return resp.Findings, resp.NextToken, nil
	})
	if err != nil {
		return nil, err
	}
	return allFindings, nil
}
