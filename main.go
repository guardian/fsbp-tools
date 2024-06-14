package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/bucketblocker/utils"
)

func main() {

	ctx := context.Background()
	args := utils.ParseArgs()

	cfg, err := utils.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	cfnClient := cloudformation.NewFromConfig(cfg)
	bucketsToBlock, err := utils.FindBucketsToBlock(ctx, securityHubClient, s3Client, cfnClient)
	if err != nil {
		log.Fatalf("Error removing GuCDK provisioned buckets: %v", err)
	}

	utils.BlockBuckets(ctx, s3Client, bucketsToBlock, args.DryRun)
}
