package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	bucketutils "github.com/guardian/fsbp-tools/fsbp-fix/bucket-utils"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
	vpcutils "github.com/guardian/fsbp-tools/fsbp-fix/vpc-utils"
)

func main() {

	ctx := context.Background()
	fixS3_8 := flag.NewFlagSet("s3.8", flag.ExitOnError)
	fixEc2_2 := flag.NewFlagSet("ec2.2", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println(len(os.Args))
		fmt.Println(os.Args)
		fmt.Println("expected 's3.8' or 'ec2.2' subcommands")
		os.Exit(1)
	}

	switch strings.ToLower(os.Args[1]) {
	case "s3.8":

		execute := fixS3_8.Bool("execute", false, "Execute the block operation")
		profile := fixS3_8.String("profile", "", "AWS profile to use")
		region := fixS3_8.String("region", "", "The region of the bucket")
		bucketCount := fixS3_8.Int("max", 100, "The maximum number of buckets to attempt to process")
		exclusions := fixS3_8.String("exclusions", "", "Comma-separated list of buckets to skip")

		fixS3_8.Parse(os.Args[2:])

		if *profile == "" {
			log.Fatal("Please provide a named AWS profile")
		}

		if *bucketCount < 1 || *bucketCount > 100 {
			log.Fatal("Please provide a max between 1 and 100")
		}

		var exclusionsSlice []string

		if *exclusions == "" {
			exclusionsSlice = []string{}
		} else {
			fmt.Printf("Parsing exclusions")
			exclusionsSlice = bucketutils.SplitAndTrim(*exclusions)
		}

		accountDetails, err := common.GetAccountDetails(ctx, *profile, *region)
		if err != nil {
			log.Fatalf("Error getting account details: %v", err)
		}

		for i, r := range accountDetails.Regions {
			fmt.Printf("Region %d: %s\n", i+1, r)
			bucketutils.FixS3_8(ctx, *profile, r, *bucketCount, exclusionsSlice, *execute)
			fmt.Printf("----------------------------------------------------\n\n")
		}

	case "ec2.2":
		execute := fixEc2_2.Bool("execute", false, "Execute the block operation")
		profile := fixEc2_2.String("profile", "", "AWS profile to use")
		region := fixEc2_2.String("region", "", "The region of the bucket")

		fixEc2_2.Parse(os.Args[2:])

		if *profile == "" {
			log.Fatal("Please provide a named AWS profile")
		}

		accountDetails, err := common.GetAccountDetails(ctx, *profile, *region)
		common.ExitOnError(err, "Failed to get account details")

		ch := make(chan vpcutils.SecurityGroupRuleDetails)
		wg := sync.WaitGroup{}

		vpcutils.FindUnusedSgRules(ctx, accountDetails, ch, &wg, *profile)
		time.Sleep(100 * time.Millisecond) // Give some time for the goroutines to stop
		vpcutils.FixEc2_2(ctx, ch, execute, profile)

	default:
		fmt.Println("expected 's3.8' or 'ec2.2' subcommands")
		os.Exit(1)
	}
}
