package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/common"
	"github.com/guardian/fsbp-tools/ingress-inquisition/utils"
)

func main() {

	ctx := context.Background()

	args := utils.ParseArgs()

	cfg, err := common.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("%v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)

	findings, err := common.ReturnFindings(ctx, securityHubClient, "EC2.2", 100)
	if err != nil {
		log.Fatalf("Error getting findings: %v", err)
	}

	securityGroups := []string{}
	for _, finding := range findings.Findings {
		for _, resource := range finding.Resources {
			sgId := utils.IdFromArn(*resource.Id)
			securityGroups = append(securityGroups, sgId)
		}
	}

	ec2Client := ec2.NewFromConfig(cfg)

	unusedSecurityGroups, err := utils.FindUnusedSecurityGroups(ctx, ec2Client, securityGroups)
	if err != nil {
		log.Fatalf("Error finding unused security groups: %v", err)
	}

	securityGroupRuleDetails := []utils.SecurityGroupRuleDetails{}

	for _, sg := range unusedSecurityGroups {
		rules, err := utils.GetSecurityGroupRuleDetails(ctx, ec2Client, sg)
		if err != nil {
			log.Fatalf("Error getting security group rules: %v", err)
		}
		securityGroupRuleDetails = append(securityGroupRuleDetails, rules...)
	}

	fmt.Println("\nIngress/egress rules on unused default security groups:")

	// Print out results as a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID\tRule Id\tFrom Port\tTo Port\tIP Protocol\tDirection")
	for _, sg := range securityGroupRuleDetails {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n", sg.SecurityGroup, sg.VpcDetails.VpcName, sg.VpcDetails.VpcId, sg.Rule.GroupRuleId, sg.Rule.FromPort, sg.Rule.ToPort, sg.Rule.IpProtocol, sg.Rule.Direction)
	}

	err = w.Flush()

	if err != nil {
		log.Fatalf("Error describing security group rules: %v", err)
	}

	fmt.Println("\n ")

	var failures int = 0
	if args.Execute {
		log.Println("Starting to delete rules...")
		for _, sgr := range securityGroupRuleDetails {
			err := utils.DeleteSecurityGroupRule(ctx, ec2Client, sgr)
			if err != nil {
				log.Printf("Error deleting rule: %v\n", sgr.Rule.GroupRuleId)
				log.Printf("Error: %v\n", err)
				failures++
			}

		}
	}

	if failures > 0 {
		log.Fatalf("Failed to delete %d rules", failures)
	}

}
