package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/guardian/bucketblocker/utils"
)

func main() {

	ctx := context.Background()
	args := utils.ParseArgs()

	cfg, err := utils.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	fmt.Println("Retrieving Security Hub control failures for S3.8")
	securityHubClient := securityhub.NewFromConfig(cfg)
	maxResults := int32(100)
	controlId := "S3.8"
	complianceStatus := "PASSED"
	recordState := "ACTIVE"

	findings, err := securityHubClient.GetFindings(ctx, &securityhub.GetFindingsInput{
		MaxResults: &maxResults,
		Filters: &types.AwsSecurityFindingFilters{
			ComplianceSecurityControlId: []types.StringFilter{{
				Value:      &controlId,
				Comparison: types.StringFilterComparisonEquals,
			}},
			ComplianceStatus: []types.StringFilter{{
				Value:      &complianceStatus,
				Comparison: types.StringFilterComparisonNotEquals,
			}},
			RecordState: []types.StringFilter{{
				Value:      &recordState,
				Comparison: types.StringFilterComparisonEquals,
			}},
		},
	})

	if err != nil {
		fmt.Println("Unable to retrieve Security Hub findings")
		return
	}

	findingsArr := findings.Findings

	var bucketsToBlock []string
	for _, finding := range findingsArr {
		for _, resource := range finding.Resources {
			bucketsToBlock = append(bucketsToBlock, strings.TrimPrefix(*resource.Id, "arn:aws:s3:::"))
		}
	}

	fmt.Println("Found " + fmt.Sprint(len(bucketsToBlock)) + " buckets failing control " + controlId)

	s3Client := s3.NewFromConfig(cfg)

	fmt.Println("Finding buckets provisioned with GuCDK, which will be skipped, to avoid drift")
	for idx, bucket := range bucketsToBlock {
		tagging, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
			Bucket: aws.String(bucket),
		})
		if err == nil {
			for _, tag := range tagging.TagSet {
				if *tag.Key == "gu:cdk:version" {
					fmt.Println("Skipping bucket: " + bucket + " provisioned with GuCDK")
					bucketsToBlock, err = utils.RemoveIndexFromSlice(bucketsToBlock, idx)
					if err != nil {
						log.Fatal("Error removing bucket from list")
					}

				}

			}
		}
	}

	fmt.Println("Found " + fmt.Sprint(len(bucketsToBlock)) + " buckets not provisioned with GuCDK")

	utils.BlockBuckets(s3Client, ctx, bucketsToBlock, args.DryRun)
}
