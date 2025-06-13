package vpcutils

import (
	"context"
	"fmt"

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

type RuleDetails struct { //does this need to be exported?
	SecurityGroup string
	VpcDetails    vpcDetails
	Rule          securityGroupRule
}

type SecurityGroupRuleDetails struct {
	Region string
	Groups []RuleDetails
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

func getSecurityGroupRuleDetails(ctx context.Context, ec2Client *ec2.Client, groupId string, region string) (SecurityGroupRuleDetails, error) {
	rules, err := getSecurityGroupRules(ctx, ec2Client, groupId)
	if err != nil {
		return SecurityGroupRuleDetails{}, err
	}
	vpcDetails, err := getVpcDetails(ctx, ec2Client, groupId)
	if err != nil {
		return SecurityGroupRuleDetails{}, err
	}

	res := SecurityGroupRuleDetails{
		Region: region,
		Groups: []RuleDetails{},
	}
	for _, rule := range rules {
		res.Groups = append(res.Groups, RuleDetails{
			SecurityGroup: groupId,
			VpcDetails:    vpcDetails,
			Rule:          rule,
		})
	}
	return res, nil
}

func findUnusedSecurityGroups(ctx context.Context, ec2Client *ec2.Client, sgIds []string) ([]string, error) {

	allNetworkInterfaces := []types.NetworkInterface{}
	securityGroupsInNetworkInterfaces := []string{}
	maxInterfaceResults := int32(1000) // Unlikely we will ever have more than 1000 network interfaces in one region

	paginator := ec2.NewDescribeNetworkInterfacesPaginator(ec2Client, &ec2.DescribeNetworkInterfacesInput{
		MaxResults: &maxInterfaceResults,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe network interfaces: %w", err)
		}
		allNetworkInterfaces = append(allNetworkInterfaces, page.NetworkInterfaces...)
	}

	for _, networkInterface := range allNetworkInterfaces {
		for _, group := range networkInterface.Groups {
			securityGroupsInNetworkInterfaces = append(securityGroupsInNetworkInterfaces, *group.GroupId)
		}
	}

	return common.Complement(sgIds, securityGroupsInNetworkInterfaces), nil
}

func FindUnusedSecurityGroupRules(ctx context.Context, ec2Client *ec2.Client, securityHubClient *securityhub.Client, accountId string, region string) (SecurityGroupRuleDetails, error) {

	findings, err := common.ReturnFindings(ctx, securityHubClient, "EC2.2", 100, accountId, region)
	if err != nil {
		return SecurityGroupRuleDetails{}, err
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
		return SecurityGroupRuleDetails{}, err
	}
	securityGroupRuleDetails := SecurityGroupRuleDetails{}

	for _, sg := range unusedSecurityGroups {
		rules, err := getSecurityGroupRuleDetails(ctx, ec2Client, sg, region)
		if err != nil {
			return SecurityGroupRuleDetails{}, err
		}
		securityGroupRuleDetails.Groups = append(securityGroupRuleDetails.Groups, rules.Groups...)
	}

	securityGroupRuleDetails.Region = region //Only set the region once we've collected all the rules
	return securityGroupRuleDetails, nil
}

func deleteSecurityGroupRule(ctx context.Context, ec2Client *ec2.Client, rule RuleDetails) error {

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

func DeleteSecurityGroupRules(ctx context.Context, ec2Client *ec2.Client, securityGroupRuleDetails SecurityGroupRuleDetails, failures *[]string) {

	fmt.Println("Deleting...")
	for _, group := range securityGroupRuleDetails.Groups {
		err := deleteSecurityGroupRule(ctx, ec2Client, group)
		if err != nil {
			fmt.Printf("Error deleting %v\n", group.Rule.GroupRuleId)
			fmt.Println(err)
			*failures = append(*failures, group.Rule.GroupRuleId)
		}
	}
}
