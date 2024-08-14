package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var exampleNonBucket = cfnTypes.StackResourceSummary{
	LogicalResourceId:    aws.String("LogicalResourceId"),
	PhysicalResourceId:   aws.String("PhysicalResourceId"),
	ResourceType:         aws.String("Not::ABucket"),
	LastUpdatedTimestamp: aws.Time(time.Now()),
}

var exampleBucket = cfnTypes.StackResourceSummary{
	LogicalResourceId:    aws.String("LogicalResourceId"),
	PhysicalResourceId:   aws.String("PhysicalResourceId"),
	ResourceType:         aws.String("AWS::S3::Bucket"),
	LastUpdatedTimestamp: aws.Time(time.Now()),
}

func FindBucketsInStacksIfTheyExist(t *testing.T) {
	stackResources := []cfnTypes.StackResourceSummary{
		exampleBucket, exampleNonBucket,
	}
	buckets := FindBucketsInStack(stackResources, "myStack")
	if len(buckets) != 1 {
		fmt.Println("Found buckets: ", buckets)
		t.Errorf("Error finding buckets in stack")
	}
}

func DoNotFindNonBucketResources(t *testing.T) {
	stackResources := []cfnTypes.StackResourceSummary{
		exampleNonBucket, exampleNonBucket,
	}
	buckets := FindBucketsInStack(stackResources, "myStack")
	if len(buckets) != 0 {
		fmt.Println("Found buckets: ", buckets)
		t.Errorf("Found a bucket where there shouldn't be one")
	}
}
