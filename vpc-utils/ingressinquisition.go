package vpcutils

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

func FindEc2_2(ch chan<- SecurityGroupRuleDetails, wg *sync.WaitGroup, ctx context.Context, cfg aws.Config, ec2Client *ec2.Client, accountId string) {
	defer wg.Done()
	securityHubClient := securityhub.NewFromConfig(cfg)
	res, _ := FindUnusedSecurityGroupRules(ctx, ec2Client, securityHubClient, accountId, cfg.Region)
	ch <- res
}

func FixEc2_2(ctx context.Context, unusedSecurityGroups SecurityGroupRuleDetails, execute bool, profile string) {
	if execute && common.UserConfirmation() {
		cfg, _ := common.Auth(ctx, profile, unusedSecurityGroups.Region) //TODO handle error
		ec2Client := ec2.NewFromConfig(cfg)
		DeleteSecurityGroupRules(ctx, ec2Client, unusedSecurityGroups)
	} else {
		fmt.Println("Skipping deletion.")
	}
}
