package vpcutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

type vpcDetails struct {
	VpcName string
	VpcId   string
}

type securityGroupRule struct {
	GroupRuleId string
	FromPort    int32
	ToPort      int32
	IpProtocol  string
	Direction   string // ingress or egress
}

type SecurityGroupRuleDetails struct {
	SecurityGroup string
	VpcDetails    vpcDetails
	Rule          securityGroupRule
}

func getSecurityGroupRules(ctx context.Context, ec2Client *ec2.Client, groupId string) ([]securityGroupRule, error) {
	fieldName := "group-id"
	rules, err := ec2Client.DescribeSecurityGroupRules(ctx, &ec2.DescribeSecurityGroupRulesInput{
		//No pagination needed. If MaxResults is not specified, then all items are returned
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

	res := []securityGroupRule{}
	for _, rule := range rules.SecurityGroupRules {
		var direction string

		if *rule.IsEgress {
			direction = "egress"
		} else {
			direction = "ingress"
		}
		res = append(res, securityGroupRule{
			GroupRuleId: *rule.SecurityGroupRuleId,
			FromPort:    *rule.FromPort,
			ToPort:      *rule.ToPort,
			IpProtocol:  *rule.IpProtocol,
			Direction:   direction,
		})
	}
	return res, nil
}

func getVpcDetails(ctx context.Context, ec2Client *ec2.Client, groupId string) (vpcDetails, error) {
	groupDescriptions, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupId},
	})
	if err != nil {
		return vpcDetails{}, err
	}

	res := []vpcDetails{}
	for _, group := range groupDescriptions.SecurityGroups {
		vpcs, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
			VpcIds: []string{*group.VpcId},
		})
		if err != nil {
			return vpcDetails{}, err
		}
		for _, vpc := range vpcs.Vpcs {
			name := FindTag(vpc.Tags, "Name", "unknown")
			res = append(res, vpcDetails{
				VpcName: name,
				VpcId:   *group.VpcId,
			})
		}
	}
	return res[0], nil // A security group cannot be associated with multiple VPCs.
}

func getSecurityGroupRuleDetails(ctx context.Context, ec2Client *ec2.Client, groupId string) ([]SecurityGroupRuleDetails, error) {
	rules, err := getSecurityGroupRules(ctx, ec2Client, groupId)
	if err != nil {
		return nil, err
	}
	vpcDetails, err := getVpcDetails(ctx, ec2Client, groupId)
	if err != nil {
		return nil, err
	}

	res := []SecurityGroupRuleDetails{}
	for _, rule := range rules {
		res = append(res, SecurityGroupRuleDetails{
			SecurityGroup: groupId,
			VpcDetails:    vpcDetails,
			Rule:          rule,
		})
	}
	return res, nil
}

func findUnusedSecurityGroups(ctx context.Context, ec2Client *ec2.Client, sgIds []string) ([]string, error) {

	securityGroupsInNetworkInterfaces := []string{}
	maxInterfaceResults := int32(10)

	allNetworkInterfaces, err := common.Paginate(func(nextToken *string) ([]types.NetworkInterface, *string, error) {
		input := &ec2.DescribeNetworkInterfacesInput{
			MaxResults: &maxInterfaceResults,
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}
		resp, err := ec2Client.DescribeNetworkInterfaces(ctx, input)
		if err != nil {
			return nil, nil, err
		}
		return resp.NetworkInterfaces, resp.NextToken, nil
	})
	if err != nil {
		return nil, err
	}

	for _, networkInterface := range allNetworkInterfaces {
		for _, group := range networkInterface.Groups {
			securityGroupsInNetworkInterfaces = append(securityGroupsInNetworkInterfaces, *group.GroupId)
		}
	}

	return common.Complement(sgIds, securityGroupsInNetworkInterfaces), nil
}

func FindUnusedSecurityGroupRules(ctx context.Context, ec2Client *ec2.Client, securityHubClient *securityhub.Client, accountId string, region string) ([]SecurityGroupRuleDetails, error) {

	findings, err := common.ReturnFindings(ctx, securityHubClient, "EC2.2", 100, accountId, region)
	if err != nil {
		return nil, err
	}

	securityGroups := []string{}

	for _, finding := range findings {
		for _, resource := range finding.Resources {
			sgId := IdFromArn(*resource.Id)
			securityGroups = append(securityGroups, sgId)
		}
	}

	unusedSecurityGroups, err := findUnusedSecurityGroups(ctx, ec2Client, securityGroups)
	if err != nil {
		return nil, err
	}
	securityGroupRuleDetails := []SecurityGroupRuleDetails{}

	for _, sg := range unusedSecurityGroups {
		rules, err := getSecurityGroupRuleDetails(ctx, ec2Client, sg)
		if err != nil {
			return nil, err
		}
		securityGroupRuleDetails = append(securityGroupRuleDetails, rules...)
	}

	if len(securityGroupRuleDetails) > 0 {

		fmt.Println("Ingress/egress rules on unused default security groups:")

		// Print out results as a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID\tRule Id\tFrom Port\tTo Port\tIP Protocol\tDirection")
		for _, sg := range securityGroupRuleDetails {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n", sg.SecurityGroup, sg.VpcDetails.VpcName, sg.VpcDetails.VpcId, sg.Rule.GroupRuleId, sg.Rule.FromPort, sg.Rule.ToPort, sg.Rule.IpProtocol, sg.Rule.Direction)
		}

		err = w.Flush()
	}

	if err != nil {
		return nil, err
	}

	return securityGroupRuleDetails, nil
}

func deleteSecurityGroupRule(ctx context.Context, ec2Client *ec2.Client, rule SecurityGroupRuleDetails) error {
	if rule.Rule.Direction == "egress" {
		_, err := ec2Client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:              &rule.SecurityGroup,
			SecurityGroupRuleIds: []string{rule.Rule.GroupRuleId},
		})
		if err != nil {
			return err
		}
	} else {
		_, err := ec2Client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:              &rule.SecurityGroup,
			SecurityGroupRuleIds: []string{rule.Rule.GroupRuleId},
		})
		if err != nil {
			return err
		}
	}
	fmt.Printf("Deleted rule %s from security group %s\n", rule.Rule.GroupRuleId, rule.SecurityGroup)
	return nil

}

func DeleteSecurityGroupRules(ctx context.Context, ec2Client *ec2.Client, securityGroupRuleDetails []SecurityGroupRuleDetails) {
	var failures int = 0
	log.Println("Starting to delete rules...")
	for _, sgr := range securityGroupRuleDetails {
		err := deleteSecurityGroupRule(ctx, ec2Client, sgr)
		if err != nil {
			log.Printf("Error deleting rule: %v\n", sgr.Rule.GroupRuleId)
			log.Printf("Error: %v\n", err)
			failures++
		}
	}
	log.Printf("Finished deleting rules.")
	if failures > 0 {
		log.Fatalf("Failed to delete %d rules", failures)
	}
}
