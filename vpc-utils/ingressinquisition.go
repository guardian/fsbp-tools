package vpcutils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

func findEc2_2(ch chan<- SecurityGroupRuleDetails, ctx context.Context, cfg aws.Config, ec2Client *ec2.Client, accountId string) {
	securityHubClient := securityhub.NewFromConfig(cfg)
	res, err := FindUnusedSecurityGroupRules(ctx, ec2Client, securityHubClient, accountId, cfg.Region)
	common.ExitOnError(err, "Failed to find unused security group rules in region "+cfg.Region)
	if len(res.Groups) > 0 {
		ch <- res
	} else {
		fmt.Printf("No unused security group rules found in %s\n", cfg.Region)
	}
}

func deleteRulesForRegion(ctx context.Context, unusedSecurityGroups SecurityGroupRuleDetails, execute bool, profile string, failures *[]string) {
	if execute && common.UserConfirmation() {
		cfg, err := common.Auth(ctx, profile, unusedSecurityGroups.Region)
		if err != nil {
			fmt.Printf("Failed to authenticate with AWS for region %s: %v\n", unusedSecurityGroups.Region, err)
			for _, sg := range unusedSecurityGroups.Groups {
				*failures = append(*failures, sg.Rule.GroupRuleId)
			}
			return
		}

		ec2Client := ec2.NewFromConfig(cfg)
		DeleteSecurityGroupRules(ctx, ec2Client, unusedSecurityGroups, failures)
	} else {
		fmt.Println("Skipping deletion.")
	}
}

func FindUnusedSgRules(ctx context.Context, accountDetails common.AccountDetails, ch chan<- SecurityGroupRuleDetails, wg *sync.WaitGroup, profile string) {
	fmt.Println("Finding unused security group rules. This will take several seconds.")
	for _, r := range accountDetails.Regions {
		cfg, err := common.Auth(ctx, profile, r)
		if err != nil {
			fmt.Printf("Failed to authenticate with AWS for region %s: %v\n", r, err)
			continue
		}
		ec2Client := ec2.NewFromConfig(cfg)
		wg.Add(1)
		go func() {
			defer wg.Done()
			findEc2_2(ch, ctx, cfg, ec2Client, accountDetails.AccountId)
		}()
	}

}

func FixEc2_2(ctx context.Context, ch <-chan SecurityGroupRuleDetails, execute *bool, profile *string) {
	failures := []string{}

	for result := range ch {
		// Print out results as a table
		fmt.Printf("%s - Unused security group rules\n\n", result.Region)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID\tRule Id\tFrom Port\tTo Port\tIP Protocol\tDirection")
		for _, sg := range result.Groups {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n", sg.SecurityGroup, sg.VpcDetails.VpcName, sg.VpcDetails.VpcId, sg.Rule.GroupRuleId, sg.Rule.FromPort, sg.Rule.ToPort, sg.Rule.IpProtocol, sg.Rule.Direction)
		}

		err := w.Flush()
		common.ExitOnError(err, "")

		deleteRulesForRegion(ctx, result, *execute, *profile, &failures) // Set execute to false for dry run
		fmt.Println("----------------------------------------------------")
	}
	if len(failures) > 0 {
		fmt.Println("Failed to delete the following rules:")
		for _, failure := range failures {
			fmt.Println(failure)
		}
	}
}
