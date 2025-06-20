package bucketutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

func findFailingBuckets(ctx context.Context, securityHubClient *securityhub.Client, bucketCount int32, accountId string, region string) ([]string, error) {
	controlId := "S3.8"
	fmt.Printf("Retrieving Security Hub control failures for %s, in %s\n", controlId, region)
	findings, err := common.ReturnFindings(ctx, securityHubClient, controlId, bucketCount, accountId, region)
	if err != nil {
		return nil, err
	}

	var bucketsToBlock []string
	for _, finding := range findings {
		for _, resource := range finding.Resources {
			bucketsToBlock = append(bucketsToBlock, strings.TrimPrefix(*resource.Id, "arn:aws:s3:::"))
		}
	}

	return bucketsToBlock, nil
}

func getAllStackSummaries(ctx context.Context, cfnClient *cloudformation.Client) ([]cfnTypes.StackSummary, error) {
	var allStackSummaries []cfnTypes.StackSummary

	input := &cloudformation.ListStacksInput{}
	stackPaginator := cloudformation.NewListStacksPaginator(cfnClient, input)
	for stackPaginator.HasMorePages() {
		page, err := stackPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list stacks: %w", err)
		}
		allStackSummaries = append(allStackSummaries, page.StackSummaries...)
	}

	fmt.Println("Found " + fmt.Sprint(len(allStackSummaries)) + " stacks.")
	return allStackSummaries, nil
}

func FindBucketsInStack(summaries []cfnTypes.StackResourceSummary, stackName string) []string {

	var buckets []string
	for _, resource := range summaries {
		if *resource.ResourceType == "AWS::S3::Bucket" {
			buckets = append(buckets, *resource.PhysicalResourceId)
		}
	}
	if len(buckets) > 0 {
		fmt.Printf("\nStack: %s - Buckets: %v", stackName, buckets)
	}
	return buckets
}

func getAllStackResources(ctx context.Context, cfnClient *cloudformation.Client, stackName string) ([]cfnTypes.StackResourceSummary, error) {

	allStackResources := []cfnTypes.StackResourceSummary{}
	stackResourcePaginator := cloudformation.NewListStackResourcesPaginator(cfnClient, &cloudformation.ListStackResourcesInput{StackName: &stackName})

	for stackResourcePaginator.HasMorePages() { // We fetch 1MB at a time, so more than one page is unlikely
		page, err := stackResourcePaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get stack resources for stack %s: %w", stackName, err)
		}
		allStackResources = append(allStackResources, page.StackResourceSummaries...)
	}

	return allStackResources, nil
}

func listBucketsInStacks(ctx context.Context, cfnClient *cloudformation.Client) []string {
	allStackSummaries, _ := getAllStackSummaries(ctx, cfnClient)
	var bucketsInAStack []string

	for _, stack := range allStackSummaries {
		if stack.StackStatus != cfnTypes.StackStatusDeleteComplete {
			stackResourceSummaries, _ := getAllStackResources(ctx, cfnClient, *stack.StackName)
			buckets := FindBucketsInStack(stackResourceSummaries, *stack.StackName)
			bucketsInAStack = append(bucketsInAStack, buckets...)
		}
	}
	fmt.Println("") //Tidy up the log output
	return bucketsInAStack
}

func FindBucketsToBlock(ctx context.Context, securityHubClient *securityhub.Client, s3Client *s3.Client, cfnClient *cloudformation.Client, bucketCount int32, exclusions []string, accountId string, region string) ([]string, error) {
	failingBuckets, err := findFailingBuckets(ctx, securityHubClient, bucketCount, accountId, region)
	if err != nil {
		return nil, err
	}

	failingBucketCount := len(failingBuckets)
	excludedBuckets := append(listBucketsInStacks(ctx, cfnClient), exclusions...)

	if len(excludedBuckets) > 0 {
		fmt.Println("\nBuckets to exclude:")
	}
	bucketsToBlock := common.Complement(failingBuckets, excludedBuckets)

	bucketsToBlockCount := len(bucketsToBlock)
	bucketsToSkipCount := failingBucketCount - bucketsToBlockCount

	if len(bucketsToBlock) > 0 {
		fmt.Println("\nBlocking the following buckets:")
		for idx, bucket := range bucketsToBlock {
			fmt.Println(idx+1, bucket)
		}
		fmt.Print("\n")
	}

	fmt.Println(failingBucketCount, "failing buckets found.")
	fmt.Println(bucketsToBlockCount, "to block, and", bucketsToSkipCount, "to skip.")
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

func BlockBuckets(ctx context.Context, s3Client *s3.Client, bucketsToBlock []string, execute bool) {
	if execute {
		userConfirmed := common.UserConfirmation()
		if userConfirmed {
			for _, name := range bucketsToBlock {
				_, err := blockPublicAccess(ctx, s3Client, name)
				if err != nil {
					fmt.Println("Error blocking public access: " + err.Error())
				}
			}
			fmt.Println("Public access blocked for all buckets. Please note it may take 24 hours for SecurityHub to update.")
		} else {
			fmt.Println("Exiting without blocking public access.")
		}
	} else {
		fmt.Println("\nSkipping execution.")
		fmt.Println("Re-run with flag -execute to block access.")
	}
}
