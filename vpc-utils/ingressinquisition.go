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

func findEc2_2(ch chan<- SecurityGroupRuleDetails, wg *sync.WaitGroup, ctx context.Context, cfg aws.Config, ec2Client *ec2.Client, accountId string) {
	defer wg.Done()
	securityHubClient := securityhub.NewFromConfig(cfg)
	res, err := FindUnusedSecurityGroupRules(ctx, ec2Client, securityHubClient, accountId, cfg.Region)
	common.ExitOnError(err, "Failed to find unused security group rules in region "+cfg.Region)
	if len(res.Groups) > 0 {
		ch <- res
	} else {
		fmt.Printf("No unused security group rules found in %s\n", cfg.Region)
	}
}

func deleteRulesForRegion(ctx context.Context, unusedSecurityGroups SecurityGroupRuleDetails, execute bool, profile string) {
	if execute && common.UserConfirmation() {
		cfg, _ := common.Auth(ctx, profile, unusedSecurityGroups.Region) //TODO handle error
		ec2Client := ec2.NewFromConfig(cfg)
		DeleteSecurityGroupRules(ctx, ec2Client, unusedSecurityGroups)
	} else {
		fmt.Println("Skipping deletion.")
	}
}

func FindUnusedSgRules(ctx context.Context, accountDetails common.AccountDetails, ch chan<- SecurityGroupRuleDetails, wg *sync.WaitGroup, profile string) {
	for _, r := range accountDetails.Regions {
		wg.Add(1)
		cfg, err := common.Auth(ctx, profile, r)
		common.ExitOnError(err, "Failed to authenticate with AWS for region "+r)
		ec2Client := ec2.NewFromConfig(cfg)
		go findEc2_2(ch, wg, ctx, cfg, ec2Client, accountDetails.AccountId)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()
}

func FixEc2_2(ctx context.Context, ch <-chan SecurityGroupRuleDetails, execute *bool, profile *string) {
	for result := range ch {
		if len(result.Groups) > 0 {
			// Print out results as a table
			fmt.Printf("%s - Unused security group rules\n\n", result.Region)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
			fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID\tRule Id\tFrom Port\tTo Port\tIP Protocol\tDirection")
			for _, sg := range result.Groups {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n", sg.SecurityGroup, sg.VpcDetails.VpcName, sg.VpcDetails.VpcId, sg.Rule.GroupRuleId, sg.Rule.FromPort, sg.Rule.ToPort, sg.Rule.IpProtocol, sg.Rule.Direction)
			}

			err := w.Flush()
			common.ExitOnError(err, "")

			deleteRulesForRegion(ctx, result, *execute, *profile) // Set execute to false for dry run
			fmt.Println("----------------------------------------------------")

		} else {
			fmt.Println("No unused security group rules found")
		}
	}
}
