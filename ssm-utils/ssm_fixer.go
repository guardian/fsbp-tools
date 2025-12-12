package ssmutils

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

// RegionDocuments holds public documents found in a region
type RegionDocuments struct {
	Region    string
	Documents []string
}

// FixSSM_7 runs the SSM.7 remediation for a single region and returns public documents found
func FixSSM_7(ctx context.Context, profile string, region string, accountId string, execute bool) []string {
	cfg, err := common.Auth(ctx, profile, region)
	common.ExitOnError(err, "Failed to authenticate with AWS for region "+region)

	ssmClient := ssm.NewFromConfig(cfg)

	publicDocs, err := FixSSMDocuments(ctx, ssmClient, region, execute)
	if err != nil {
		fmt.Printf("Error in region %s: %v\n", region, err)
		return []string{}
	}

	return publicDocs
}

// RunSSM7FixerForAllRegions processes SSM.7 remediation across all regions and prints summary
func RunSSM7FixerForAllRegions(ctx context.Context, profile string, accountDetails common.AccountDetails, execute bool) {
	var allPublicDocuments []RegionDocuments

	for i, r := range accountDetails.Regions {
		fmt.Printf("Region %d: %s\n", i+1, r)
		publicDocs := FixSSM_7(ctx, profile, r, accountDetails.AccountId, execute)
		if len(publicDocs) > 0 {
			allPublicDocuments = append(allPublicDocuments, RegionDocuments{
				Region:    r,
				Documents: publicDocs,
			})
		}
		fmt.Printf("----------------------------------------------------\n\n")
	}

	PrintSummary(allPublicDocuments, execute)
}

// PrintSummary displays a summary of all public documents found across regions
func PrintSummary(allPublicDocuments []RegionDocuments, execute bool) {
	if len(allPublicDocuments) > 0 {
		fmt.Println("\n========================================")
		fmt.Println("SUMMARY: Public SSM Documents Found")
		fmt.Println("========================================")
		totalDocs := 0
		for _, rd := range allPublicDocuments {
			fmt.Printf("\n%s (%d document(s)):\n", rd.Region, len(rd.Documents))
			for _, doc := range rd.Documents {
				fmt.Printf("  - %s\n", doc)
			}
			totalDocs += len(rd.Documents)
		}
		fmt.Printf("\nTotal: %d public document(s) across %d region(s)\n", totalDocs, len(allPublicDocuments))
		if !execute {
			fmt.Println("\n⚠️  IMPORTANT: Please investigate these documents before executing fixes.")
			fmt.Println("Fixing will remove public sharing and disable region-level public sharing.")
			fmt.Println("\nTo proceed with fixes, re-run with the -execute flag.")
		}
		fmt.Println("========================================")
	} else {
		fmt.Println("\n✅ No public SSM documents found across all regions.")
	}
}
