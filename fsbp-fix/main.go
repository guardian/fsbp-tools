package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	bucketUtils "github.com/guardian/fsbp-tools/fsbp-fix/bucket-utils"
)

func main() {

	ctx := context.Background()

	fixS3_8 := flag.NewFlagSet("S3.8", flag.ExitOnError)

	// fixEc2_2 := flag.NewFlagSet("EC2.2", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println(len(os.Args))
		fmt.Println(os.Args)
		fmt.Println("expected 'S3.8' or 'EC2.2' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "S3.8":

		execute := fixS3_8.Bool("execute", false, "Execute the block operation")
		profile := fixS3_8.String("profile", "", "AWS profile to use")
		region := fixS3_8.String("region", "", "The region of the bucket")
		bucketCount := fixS3_8.Int("max", 100, "The maximum number of buckets to attempt to process")
		exclusions := fixS3_8.String("exclusions", "", "Comma-separated list of buckets to skip")

		fixS3_8.Parse(os.Args[2:])

		if *profile == "" {
			log.Fatal("Please provide a named AWS profile")
		}

		if *region == "" {
			log.Fatal("Please provide a region")
		}

		if *bucketCount < 1 || *bucketCount > 100 {
			log.Fatal("Please provide a max between 1 and 100")
		}

		var exclusionsSlice []string

		if *exclusions == "" {
			exclusionsSlice = []string{}
		} else {
			fmt.Printf("Parsing exclusions")
			exclusionsSlice = bucketUtils.SplitAndTrim(*exclusions)
		}
		bucketUtils.BucketBlocker(ctx, *profile, *region, *bucketCount, exclusionsSlice, *execute)

	// case "EC2.2":
	// 	fixEc2_2.Parse(os.Args[2:])
	default:
		fmt.Println("expected 'S3.8' or 'EC2.2' subcommands")
		os.Exit(1)
	}

}
