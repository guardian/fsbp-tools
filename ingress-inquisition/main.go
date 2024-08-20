package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/common"
	"github.com/guardian/fsbp-tools/ingress-inquisition/utils"
)

type VpcDetails struct {
	VpcName string
	VpcId   string
}

type SecurityGroupRule struct {
	GroupRuleId string
	FromPort    int32
	ToPort      int32
	IpProtocol  string
	Direction   string // ingress or egress
}

type SecurityGroupRuleDetails struct {
	SecurityGroup string
	VpcDetails    VpcDetails
	Rule          SecurityGroupRule
}

func getSecurityGroupRules(ctx context.Context, ec2Client *ec2.Client, groupId string) ([]SecurityGroupRule, error) {
	fieldName := "group-id"
	rules, err := ec2Client.DescribeSecurityGroupRules(ctx, &ec2.DescribeSecurityGroupRulesInput{
		Filters: []types.Filter{
			{
				Name:   &fieldName,
				Values: []string{groupId},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	res := []SecurityGroupRule{}
	for _, rule := range rules.SecurityGroupRules {
		var direction string

		if *rule.IsEgress {
			direction = "egress"
		} else {
			direction = "ingress"
		}
		res = append(res, SecurityGroupRule{
			GroupRuleId: *rule.SecurityGroupRuleId,
			FromPort:    *rule.FromPort,
			ToPort:      *rule.ToPort,
			IpProtocol:  *rule.IpProtocol,
			Direction:   direction,
		})
	}
	return res, nil
}

func getVpcDetails(ctx context.Context, ec2Client *ec2.Client, groupId string) (VpcDetails, error) {
	groupDescriptions, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupId},
	})
	if err != nil {
		return VpcDetails{}, err
	}

	res := []VpcDetails{}
	for _, group := range groupDescriptions.SecurityGroups {
		vpcs, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
			VpcIds: []string{*group.VpcId},
		})
		if err != nil {
			return VpcDetails{}, err
		}
		for _, vpc := range vpcs.Vpcs {
			name := utils.FindTag(vpc.Tags, "Name", "unknown")
			res = append(res, VpcDetails{
				VpcName: name,
				VpcId:   *group.VpcId,
			})
		}
	}
	return res[0], nil // A security group cannot be associated with multiple VPCs.
}

func findUnusedSecurityGroups(ctx context.Context, ec2Client *ec2.Client, sgIds []string) ([]string, error) {

	securityGroupsInNetworkInterfaces := []string{}

	maxInterfaceResults := int32(1000)
	res, err := ec2Client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		MaxResults: &maxInterfaceResults,
	})
	if err != nil {
		return nil, err
	}
	for _, networkInterface := range res.NetworkInterfaces {
		for _, group := range networkInterface.Groups {
			securityGroupsInNetworkInterfaces = append(securityGroupsInNetworkInterfaces, *group.GroupId)
		}
	}

	return common.Complement(sgIds, securityGroupsInNetworkInterfaces), nil
}

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

	unusedSecurityGroups, err := findUnusedSecurityGroups(ctx, ec2Client, securityGroups)
	if err != nil {
		log.Fatalf("Error finding unused security groups: %v", err)
	}

	var securityGroupRuleDetails []SecurityGroupRuleDetails

	for _, sg := range unusedSecurityGroups {
		vpcDetails, err := getVpcDetails(ctx, ec2Client, sg)
		if err != nil {
			log.Fatalf("Error getting VPC details: %v", err)
		}
		rules, err := getSecurityGroupRules(ctx, ec2Client, sg)
		if err != nil {
			log.Fatalf("Error getting security group rules: %v", err)
		}
		for _, rule := range rules {
			securityGroupRuleDetails = append(securityGroupRuleDetails, SecurityGroupRuleDetails{
				SecurityGroup: sg,
				VpcDetails:    vpcDetails,
				Rule:          rule,
			})
		}
	}

	fmt.Println("\nIngress/egress rules on unused default security groups:")

	// Print out results as a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID\tFrom Port\tTo Port\tIP Protocol\tDirection")
	for _, sg := range securityGroupRuleDetails {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n", sg.SecurityGroup, sg.VpcDetails.VpcName, sg.VpcDetails.VpcId, sg.Rule.FromPort, sg.Rule.ToPort, sg.Rule.IpProtocol, sg.Rule.Direction)
	}

	w.Flush()

	if err != nil {
		log.Fatalf("Error describing security group rules: %v", err)
	}

}
