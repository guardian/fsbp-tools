package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/account"
	acc "github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	shTypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AccountDetails struct {
	AccountId string
	Profile   string
	Regions   []string
}

func validateCredentials(ctx context.Context, stsClient *sts.Client, profile string) (*sts.GetCallerIdentityOutput, error) {
	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return resp, errors.New("Could not find valid credentials for profile: " + profile)
	}
	return resp, nil
}

func Auth(ctx context.Context, profile string, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithSharedConfigProfile(profile),
		config.WithDefaultRegion(region),
	)
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

func getAccountId(ctx context.Context, cfg aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(cfg)
	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("error getting caller identity: %w", err)
	}

	return *resp.Account, nil
}

func listEnabledRegions(ctx context.Context, cfg aws.Config) ([]string, error) {
	fmt.Println("No region provided, running globally in all enabled regions")
	accountClient := account.NewFromConfig(cfg)
	resp, err := accountClient.ListRegions(ctx, &account.ListRegionsInput{
		RegionOptStatusContains: []acc.RegionOptStatus{acc.RegionOptStatusEnabled, acc.RegionOptStatusEnabledByDefault},
	})

	if err != nil {
		return nil, err
	}
	fmt.Printf("%d regions enabled.\n", len(resp.Regions))
	enabledRegions := []string{}
	for _, region := range resp.Regions {
		if region.RegionName != nil {
			enabledRegions = append(enabledRegions, *region.RegionName)
		}
	}
	return enabledRegions, nil
}

func GetAccountDetails(ctx context.Context, profile string, region string) (AccountDetails, error) {
	cfg, err := Auth(ctx, profile, "eu-west-1") // used to get accountId and enabled regions
	if err != nil {
		return AccountDetails{}, fmt.Errorf("failed to authenticate with AWS: %w", err)
	}

	accountId, err := getAccountId(ctx, cfg)

	var regions []string

	if region == "" {
		regions, err = listEnabledRegions(ctx, cfg)
	} else {
		regions = []string{region}
	}

	if err != nil {
		return AccountDetails{}, fmt.Errorf("failed to get account details: %w", err)
	}

	return AccountDetails{
		AccountId: accountId,
		Profile:   profile,
		Regions:   regions,
	}, nil
}

func findingsInput(controlId string, maxResults int32, accountId string, region string) *securityhub.GetFindingsInput {
	return &securityhub.GetFindingsInput{
		MaxResults: &maxResults,
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

	fmt.Printf("Retrieving Security Hub control failures for %s, in %s\n", controlId, region)
	allFindings := []shTypes.AwsSecurityFinding{}
	input := findingsInput(controlId, maxResults, accountId, region)

	paginator := securityhub.NewGetFindingsPaginator(securityHubClient, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get findings: %w", err)
		}
		allFindings = append(allFindings, page.Findings...)
	}

	return allFindings, nil
}
