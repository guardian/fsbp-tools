package bucketutils

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

func FixS3_8(ctx context.Context, cfg aws.Config, bucketCount int, exclusions []string, execute bool) {

	accountId, err := common.GetAccountId(ctx, cfg)
	if err != nil {
		log.Fatalf("Error getting account ID: %v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	cfnClient := cloudformation.NewFromConfig(cfg)
	bucketsToBlock, err := FindBucketsToBlock(ctx, securityHubClient, s3Client, cfnClient, int32(bucketCount), exclusions, accountId, cfg.Region)
	if err != nil {
		log.Fatalf("Error working out which buckets need blocking: %v", err)
	}

	BlockBuckets(ctx, s3Client, bucketsToBlock, execute)
}
