package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	shTypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go/aws"
)

func blockPublicAccess(s3Client *s3.Client, ctx context.Context, name string) (*s3.PutPublicAccessBlockOutput, error) {
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

func validateCredentials(stsClient *sts.Client, ctx context.Context, profile string) (*sts.GetCallerIdentityOutput, error) {
	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return resp, errors.New("Could not find valid credentials for profile: " + profile)
	}
	return resp, nil
}

func main() {
	ctx := context.Background()
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The region of the bucket")
	dryRun := flag.Bool("dry-run", true, "Dry run mode")
	flag.Parse()

	if *profile == "" {
		fmt.Println("Please provide a profile name")
		return
	}

	if *region == "" {
		fmt.Println("Please provide a region")
		return
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(*profile), config.WithDefaultRegion(*region))
	if err != nil {
		fmt.Println("Error loading configuration")
		return
	}

	stsClient := sts.NewFromConfig(cfg)

	_, err = validateCredentials(stsClient, ctx, *profile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Retrieving Security Hub control failures for S3.8")
	securityHubClient := securityhub.NewFromConfig(cfg)
	maxResults := int32(100)
	controlId := "S3.8"
	complianceStatus := "PASSED"
	recordState := "ACTIVE"

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
					bucketsToBlock, err = removeIndexFromSlice(bucketsToBlock, idx)
					if err != nil {
						fmt.Println("Error removing bucket from list")
						return
					}

				}

			}
		}
	}

	fmt.Println("Found " + fmt.Sprint(len(bucketsToBlock)) + " buckets not provisioned with GuCDK")

	if *dryRun {
		fmt.Println("Dry run mode enabled. Skipping blocking public access for buckets")
	} else {
		fmt.Println("Blocking public access for buckets in 5 seconds. Press CTRL+C to cancel.")
		time.Sleep(5 * time.Second)
		for _, name := range bucketsToBlock {
			_, err = blockPublicAccess(s3Client, ctx, name)
			if err != nil {
				fmt.Println("Error blocking public access: " + err.Error())
				return
			}

		}
	}

}
