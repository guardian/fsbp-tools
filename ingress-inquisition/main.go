package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/guardian/fsbp-tools/bucket-blocker/utils"
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

func main() {

	ctx := context.Background()

	args := ParseArgs()

	_, err := utils.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	} else {
		fmt.Println("Config loaded successfully")
	}
}
