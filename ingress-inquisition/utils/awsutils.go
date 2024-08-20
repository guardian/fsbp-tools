package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/guardian/fsbp-tools/common"
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
			name := FindTag(vpc.Tags, "Name", "unknown")
			res = append(res, VpcDetails{
				VpcName: name,
				VpcId:   *group.VpcId,
			})
		}
	}
	return res[0], nil // A security group cannot be associated with multiple VPCs.
}

func GetSecurityGroupRuleDetails(ctx context.Context, ec2Client *ec2.Client, groupId string) ([]SecurityGroupRuleDetails, error) {
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

func FindUnusedSecurityGroups(ctx context.Context, ec2Client *ec2.Client, sgIds []string) ([]string, error) {

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
