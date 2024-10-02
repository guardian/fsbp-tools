package vpcutils

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

func FixEc2_2(ctx context.Context, profile *string, region *string, execute *bool) {

	cfg, err := common.LoadDefaultConfig(ctx, *profile, *region)
	if err != nil {
		log.Fatalf("%v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	securityHubClient := securityhub.NewFromConfig(cfg)

	securityGroupRuleDetails, err := FindUnusedSecurityGroupRules(ctx, ec2Client, securityHubClient)

	if err != nil {
		log.Fatalf("Error finding unused security group rules: %v", err)
	} else if len(securityGroupRuleDetails) == 0 {
		fmt.Println("No unused security groups found")
	} else if *execute && common.UserConfirmation() {
		fmt.Println("\n ")
		DeleteSecurityGroupRules(ctx, ec2Client, securityGroupRuleDetails)
	}
}
