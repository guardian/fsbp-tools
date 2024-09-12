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
	}

	fmt.Println("\n ")

	var failures int = 0
	if args.Execute && len(securityGroupRuleDetails) > 0 {
		userConfirmed := common.UserConfirmation()
		if userConfirmed {
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
	}

	if failures > 0 {
		log.Fatalf("Failed to delete %d rules", failures)
	}

}
