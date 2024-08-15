package utils

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestEmptyArn(t *testing.T) {
	str := ""
	result := IdFromArn(str)
	if result != "" {
		t.Errorf("Error processing ARN. Expected empty string, got %s", result)
	}
}

func TestArnWithoutSlash(t *testing.T) {
	str := "abcd"
	result := IdFromArn(str)
	if result != str {
		t.Errorf("Error processing ARN. Expected %s, got %s", str, result)
	}
}

func TestArnWithSingleSlash(t *testing.T) {
	str := "arn:aws:ec2:us-west-2:123456789012:instance/i-1234567890abcdef0"
	result := IdFromArn(str)
	if result != "i-1234567890abcdef0" {
		t.Errorf("Error processing ARN. Expected i-1234567890abcdef0, got %s", result)
	}
}

func TestFindTagNoTags(t *testing.T) {
	tags := []types.Tag{}
	key := "Name"
	defaultValue := "none"
	result := FindTag(tags, key, defaultValue)
	if result != defaultValue {
		t.Errorf("Error finding tag. Expected %s, got %s", defaultValue, result)
	}
}

func TestFindTagNoMatchingTag(t *testing.T) {

	key := "MyKey"
	value := "MyValue"

	tags := []types.Tag{
		{Key: &key, Value: &value},
	}
	defaultValue := "none"
	result := FindTag(tags, "Name", defaultValue)
	if result != defaultValue {
		t.Errorf("Error finding tag. Expected %s, got %s", defaultValue, result)
	}
}

func TestFindTagMatchingTag(t *testing.T) {
	key := "Name"
	value := "MyValue"
	tags := []types.Tag{
		{Key: &key, Value: &value},
	}
	defaultValue := "none"
	result := FindTag(tags, key, defaultValue)
	if result != value {
		t.Errorf("Error finding tag. Expected %s, got %s", value, result)
	}
}
