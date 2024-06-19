package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func findFailingBuckets(ctx context.Context, securityHubClient *securityhub.Client, bucketCount int32) ([]string, error) {
	controlId := "S3.8"
	complianceStatus := "PASSED"
	recordState := "ACTIVE"

	fmt.Println("Retrieving Security Hub control failures for S3.8")
	findings, err := securityHubClient.GetFindings(ctx, &securityhub.GetFindingsInput{
		MaxResults: &bucketCount,
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

	findingsArr := findings.Findings

	var bucketsToBlock []string
	for _, finding := range findingsArr {
		for _, resource := range finding.Resources {
			bucketsToBlock = append(bucketsToBlock, strings.TrimPrefix(*resource.Id, "arn:aws:s3:::"))
		}
	}

	return bucketsToBlock, nil
}

func listBucketsInStacks(ctx context.Context, cfnClient *cloudformation.Client) []string {

	var bucketsInAStack []string
	stacks, _ := cfnClient.ListStacks(ctx, &cloudformation.ListStacksInput{})
	fmt.Println("Found " + fmt.Sprint(len(stacks.StackSummaries)) + " stacks in account. Enumerating stacks with buckets:")

	for _, stack := range stacks.StackSummaries {
		if stack.StackStatus != cfnTypes.StackStatusDeleteComplete {
			stackResources, _ := cfnClient.ListStackResources(ctx, &cloudformation.ListStackResourcesInput{StackName: stack.StackName})
			for _, resource := range stackResources.StackResourceSummaries {
				var buckets []string
				if *resource.ResourceType == "AWS::S3::Bucket" {
					buckets = append(buckets, *resource.PhysicalResourceId)
					bucketsInAStack = append(bucketsInAStack, *resource.PhysicalResourceId)
				}
				if len(buckets) > 0 {
					fmt.Printf("\nStack: %s - Buckets: %v", *stack.StackName, buckets)
				}
			}
		}
	}
	fmt.Println("") //Tidy up the log output

	return bucketsInAStack
}

func FindBucketsToBlock(ctx context.Context, securityHubClient *securityhub.Client, s3Client *s3.Client, cfnClient *cloudformation.Client, bucketCount int32, exclusions []string) ([]string, error) {
	failingBuckets, err := findFailingBuckets(ctx, securityHubClient, bucketCount)
	if err != nil {
		return nil, err
	}

	failingBucketCount := len(failingBuckets)
	excludedBuckets := append(listBucketsInStacks(ctx, cfnClient), exclusions...)

	fmt.Println("\nBuckets to exclude:")
	bucketsToBlock := Complement(failingBuckets, excludedBuckets)

	bucketsToBlockCount := len(bucketsToBlock)
	bucketsToSkipCount := failingBucketCount - bucketsToBlockCount

	fmt.Println("\nBlocking the following buckets:")
	for idx, bucket := range bucketsToBlock {
		fmt.Println(idx+1, bucket)
	}

	fmt.Println(failingBucketCount, "failing buckets found. ", bucketsToSkipCount, "will be skipped, to avoid stack drift.")
	return bucketsToBlock, nil

}

func blockPublicAccess(ctx context.Context, s3Client *s3.Client, name string) (*s3.PutPublicAccessBlockOutput, error) {
	resp, err := s3Client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(name),
		PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return resp, err
	}
	fmt.Println("Public access blocked for bucket: " + name)
	return resp, nil
}

func BlockBuckets(ctx context.Context, s3Client *s3.Client, bucketsToBlock []string, dryRun bool) {
	if dryRun {
		fmt.Println("Dry run mode enabled. Skipping blocking public access for buckets")
	} else {
		buf := bufio.NewReader(os.Stdin)
		fmt.Println("\nPress 'y', to confirm, and enter to continue. Otherwise, hit enter to exit.")
		fmt.Print("> ")
		input, err := buf.ReadBytes('\n')
		if err != nil {
			fmt.Println("Error reading input: " + err.Error())
		}
		if strings.ToLower(strings.TrimSpace(string(input))) == "y" {
			for _, name := range bucketsToBlock {
				_, err := blockPublicAccess(ctx, s3Client, name)
				if err != nil {
					fmt.Println("Error blocking public access: " + err.Error())
				}
			}
			fmt.Println("Public access blocked for all buckets. Please note it may take 24 hours for SecurityHub to update.")
		} else {
			fmt.Println("Exiting without blocking public access")
		}
	}
}
