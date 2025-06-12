package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	bucketutils "github.com/guardian/fsbp-tools/fsbp-fix/bucket-utils"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
	vpcutils "github.com/guardian/fsbp-tools/fsbp-fix/vpc-utils"
)

type AccountDetails struct {
	AccountId string
	Profile   string
	Regions   []string
}

func getAccountDetails(ctx context.Context, profile string, region string) (AccountDetails, error) {
	cfg, err := common.Auth(ctx, profile, "eu-west-1") // used to get accountId and enabled regions
	if err != nil {
		return AccountDetails{}, fmt.Errorf("failed to authenticate with AWS: %w", err)
	}

	accountId, err := common.GetAccountId(ctx, cfg)

	var regions []string

	if region == "" {
		regions, err = common.ListEnabledRegions(ctx, cfg)
	} else {
		regions = []string{region}
	}

	if err != nil {
		return AccountDetails{}, fmt.Errorf("failed to get account details: %w", err)
	}

	return AccountDetails{
		AccountId: accountId,
		Profile:   profile,
		Regions:   regions,
	}, nil
}

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

		accountDetails, err := getAccountDetails(ctx, *profile, *region)
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

		accountDetails, err := getAccountDetails(ctx, *profile, *region)
		common.ExitOnError(err, "Failed to get account details")

		ch := make(chan vpcutils.SecurityGroupRuleDetails)
		wg := sync.WaitGroup{}

		for _, r := range accountDetails.Regions {
			wg.Add(1)
			cfg, err := common.Auth(ctx, *profile, r)
			common.ExitOnError(err, "Failed to authenticate with AWS for region "+r)
			ec2Client := ec2.NewFromConfig(cfg)
			go vpcutils.FindEc2_2(ch, &wg, ctx, cfg, ec2Client, accountDetails.AccountId)
		}

		go func() {
			wg.Wait()
			time.Sleep(100 * time.Millisecond) // Give goroutines time to finish before closing the channel
			close(ch)
		}()

		for result := range ch {
			if len(result.Groups) > 0 {
				// Print out results as a table
				fmt.Printf("%s - Unused security group rules\n\n", result.Region)
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
				fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID\tRule Id\tFrom Port\tTo Port\tIP Protocol\tDirection")
				for _, sg := range result.Groups {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n", sg.SecurityGroup, sg.VpcDetails.VpcName, sg.VpcDetails.VpcId, sg.Rule.GroupRuleId, sg.Rule.FromPort, sg.Rule.ToPort, sg.Rule.IpProtocol, sg.Rule.Direction)
				}

				err = w.Flush()
				common.ExitOnError(err, "Failed to flush tabwriter")

				vpcutils.FixEc2_2(ctx, result, *execute, *profile) // Set execute to false for dry run
				fmt.Println("----------------------------------------------------")

			} else {
				fmt.Println("No unused security group rules found")
			}
		}

	default:
		fmt.Println("expected 's3.8' or 'ec2.2' subcommands")
		os.Exit(1)
	}

}
