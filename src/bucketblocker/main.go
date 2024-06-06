package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

func blockPublicAccess(s3Client *s3.S3, name string) (*s3.PutPublicAccessBlockOutput, error) {
	resp, err := s3Client.PutPublicAccessBlock(&s3.PutPublicAccessBlockInput{
		Bucket: aws.String(name),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
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

func validateCredentials(stsClient *sts.STS, profile string) (*sts.GetCallerIdentityOutput, error) {
	resp, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return resp, errors.New("Could not find valid credentials for profile: " + profile)
	}
	return resp, nil
}

func main() {
	name := flag.String("bucket", "", "The name of the bucket to block")
	profile := flag.String("profile", "default", "The name of the profile to use")
	region := flag.String("region", "eu-west-1", "The region of the bucket")
	flag.Parse()

	if *name == "" {
		fmt.Println("Please provide a bucket name")
		return
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           *profile,
		Config: aws.Config{
			Region: aws.String(*region),
		},
	}))

	stsClient := sts.New(sess)
	_, err := validateCredentials(stsClient, *profile)
	if err != nil {
		fmt.Println(err)
		return
	}

	s3Client := s3.New(sess)

	//check bucket exists
	_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(*name),
	})
	if err != nil {
		fmt.Println("Unable to find bucket. Please make the bucket exists and you have the correct region set.")
		return
	}
	fmt.Println("Found bucket: " + *name + " in region: " + *region)

	_, err = blockPublicAccess(s3Client, *name)
	if err != nil {
		fmt.Println("Error blocking public access: " + err.Error())
		return
	}
}
