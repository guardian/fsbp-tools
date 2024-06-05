package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

func blockPublicAccess(name string, s3Client s3.S3) {
	publicAccessBlock, err := s3Client.GetPublicAccessBlock(&s3.GetPublicAccessBlockInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(publicAccessBlock)

	_, err = s3Client.PutPublicAccessBlock(&s3.PutPublicAccessBlockInput{
		Bucket: aws.String(name),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Public access blocked for bucket: " + name)
}

func validateCredentials(stsClient sts.STS, profile string) {
	_, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Println("Could not find valid credentials for profile: " + profile)
		return
	}
	fmt.Println("Credentials validated for profile: " + profile)
}

func main() {
	name := flag.String("bucket", "", "The name of the bucket to block")
	profile := flag.String("profile", "default", "The name of the profile to use")
	region := flag.String("region", "eu-west-1", "The region of the bucket")
	flag.Parse()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           *profile,
		Config: aws.Config{
			Region: aws.String(*region),
		},
	}))

	stsClient := sts.New(sess)
	validateCredentials(*stsClient, *profile)

	s3Client := s3.New(sess)

	//check bucket exists
	bucketInfo, err := s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(*name),
	})
	fmt.Println(bucketInfo)
	if err != nil {
		fmt.Println("Unable to find bucket. Please make the bucket exists and you have the correct region set.")
		return
	}

	blockPublicAccess(*name, *s3Client)
}
