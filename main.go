package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	bucketutils "github.com/guardian/fsbp-tools/fsbp-fix/bucket-utils"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
	ssmutils "github.com/guardian/fsbp-tools/fsbp-fix/ssm-utils"
	vpcutils "github.com/guardian/fsbp-tools/fsbp-fix/vpc-utils"
)

func main() {

	ctx := context.Background()
	fixS3_8 := flag.NewFlagSet("s3.8", flag.ExitOnError)
	fixEc2_2 := flag.NewFlagSet("ec2.2", flag.ExitOnError)
	fixSSM_7 := flag.NewFlagSet("ssm.7", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println(len(os.Args))
		fmt.Println(os.Args)
		fmt.Println("expected 's3.8' or 'ec2.2' or 'ssm.7' subcommands")
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
		go func() {
			wg.Wait()
			close(ch)
		}()
		vpcutils.FixEc2_2(ctx, ch, execute, profile)

	case "ssm.7":
		execute := fixSSM_7.Bool("execute", false, "Execute the fix operation")
		profile := fixSSM_7.String("profile", "", "AWS profile to use")
		region := fixSSM_7.String("region", "", "The region to check")

		fixSSM_7.Parse(os.Args[2:])

		if *profile == "" {
			log.Fatal("Please provide a named AWS profile")
		}

		accountDetails, err := common.GetAccountDetails(ctx, *profile, *region)
		common.ExitOnError(err, "Failed to get account details")

		ssmutils.RunSSM7FixerForAllRegions(ctx, *profile, accountDetails, *execute)

	default:
		fmt.Println("expected 's3.8' or 'ec2.2' or 'ssm.7' subcommands")
		os.Exit(1)
	}
}
