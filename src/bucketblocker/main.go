package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go/aws"
)

func blockPublicAccess(s3Client *s3.Client, ctx context.Context, name string) (*s3.PutPublicAccessBlockOutput, error) {
	resp, err := s3Client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(name),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
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
	name := flag.String("bucket", "", "The name of the bucket to block")
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The region of the bucket")
	flag.Parse()

	if *name == "" {
		fmt.Println("Please provide a bucket name")
		return
	}

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

	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: name,
	})
	if err != nil {
		fmt.Println("Unable to find bucket. Please make the bucket exists and you have the correct region set.")
		return
	}
	fmt.Println("Found bucket: " + *name + " in region: " + *region)

	_, err = blockPublicAccess(s3Client, ctx, *name)
	if err != nil {
		fmt.Println("Error blocking public access: " + err.Error())
		return
	}

}
