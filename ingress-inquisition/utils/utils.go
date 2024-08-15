package utils

import (
	"flag"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type cliArgs struct {
	Profile string
	Region  string
}

func ParseArgs() cliArgs {
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The AWS region to use")

	flag.Parse()

	if *profile == "" {
		log.Fatal("Please provide a named AWS profile")
	}

	if *region == "" {
		log.Fatal("Please provide a region")
	}

	return cliArgs{
		Profile: *profile,
		Region:  *region,
	}
}

func IdFromArn(arn string) string {
	splitArr := strings.Split(arn, "/")
	return splitArr[len(splitArr)-1]
}

func FindTag(tags []types.Tag, key string, defaultValue string) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}
	return defaultValue
}
