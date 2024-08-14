package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awsauth "github.com/guardian/fsbp-tools/aws-auth"
	"github.com/guardian/fsbp-tools/bucket-blocker/utils"
)

func main() {

	ctx := context.Background()
	args := utils.ParseArgs()

	cfg, err := awsauth.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	cfnClient := cloudformation.NewFromConfig(cfg)
	bucketsToBlock, err := utils.FindBucketsToBlock(ctx, securityHubClient, s3Client, cfnClient, args.BucketCount, args.Exclusions)
	if err != nil {
		log.Fatalf("Error working out which buckets need blocking: %v", err)
	}

	utils.BlockBuckets(ctx, s3Client, bucketsToBlock, args.Execute)
}
