package ssmutils

import (
	"context"
	"fmt"
    
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// GetAllRegions retrieves all regions that are ENABLED for the current AWS account.
// It uses a bootstrap region (e.g., us-east-1) to initialize the EC2 client.
func GetAllRegions(ctx context.Context, profile string) ([]string, error) {
    // You must use a region to initialize the client, we'll use a widely available one.
    bootstrapRegion := "us-east-1" 

    // Load configuration for the bootstrap region
	cfg, err := config.LoadDefaultConfig(ctx, 
		config.WithSharedConfigProfile(profile),
		config.WithRegion(bootstrapRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for region discovery: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	
    // DescribeRegions input. We only want regions ENABLED for this account.
	input := &ec2.DescribeRegionsInput{} 

	resp, err := ec2Client.DescribeRegions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	var regions []string
	for _, region := range resp.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}

	return regions, nil
}