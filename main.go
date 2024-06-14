package main

import (
	"context"
	"fmt"
	"log"

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

	fmt.Println("Retrieving Security Hub control failures for S3.8")
	securityHubClient := securityhub.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)

	bucketsToBlock, err := utils.FindBucketsToBlock(ctx, securityHubClient, s3Client)

	if err != nil {
		log.Fatalf("Error removing GuCDK provisioned buckets: %v", err)
	}

	utils.BlockBuckets(s3Client, ctx, bucketsToBlock, args.DryRun)
}
