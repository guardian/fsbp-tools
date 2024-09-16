package main

import (
	"context"
	"fmt"
	"log"

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

	ec2Client := ec2.NewFromConfig(cfg)
	securityHubClient := securityhub.NewFromConfig(cfg)

	securityGroupRuleDetails, err := utils.FindUnusedSecurityGroupRules(ctx, ec2Client, securityHubClient)

	if err != nil {
		log.Fatalf("Error finding unused security group rules: %v", err)
	} else if len(securityGroupRuleDetails) == 0 {
		fmt.Println("No unused security groups found")
	} else if args.Execute && common.UserConfirmation() {
		fmt.Println("\n ")
		utils.DeleteSecurityGroupRules(ctx, ec2Client, securityGroupRuleDetails)
	}

}
