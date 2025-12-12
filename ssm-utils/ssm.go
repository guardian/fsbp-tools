package ssmutils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/guardian/fsbp-tools/fsbp-fix/common"
)

// CheckAllDocumentsForPublicSharing checks all customer-owned SSM documents for public sharing
func CheckAllDocumentsForPublicSharing(ctx context.Context, ssmClient *ssm.Client) ([]string, error) {
	var publicDocuments []string

	fmt.Println("Scanning all customer-owned SSM documents...")

	// List all documents owned by the account
	paginator := ssm.NewListDocumentsPaginator(ssmClient, &ssm.ListDocumentsInput{
		Filters: []ssmTypes.DocumentKeyValuesFilter{
			{
				Key:    aws.String("Owner"),
				Values: []string{"Self"},
			},
		},
	})

	documentCount := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list SSM documents: %w", err)
		}

		for _, document := range page.DocumentIdentifiers {
			if document.Name == nil {
				continue
			}
			docName := *document.Name
			documentCount++

			// Check if document is publicly shared
			isPublic, err := CheckDocumentPublicAccess(ctx, ssmClient, docName)
			if err != nil {
				fmt.Printf("Warning: Could not check public access for document %s: %v\n", docName, err)
				continue
			}

			if isPublic {
				publicDocuments = append(publicDocuments, docName)
			}

			// Rate limiting to avoid throttling
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Printf("Scanned %d customer-owned document(s)\n", documentCount)
	return publicDocuments, nil
}

// CheckDocumentPublicAccess checks if an SSM document is publicly shared
func CheckDocumentPublicAccess(ctx context.Context, ssmClient *ssm.Client, documentName string) (bool, error) {
	input := &ssm.DescribeDocumentPermissionInput{
		Name:           aws.String(documentName),
		PermissionType: ssmTypes.DocumentPermissionTypeShare,
	}

	result, err := ssmClient.DescribeDocumentPermission(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not supported") {
			return false, nil
		}
		return false, fmt.Errorf("failed to get document permission: %w", err)
	}

	// Check if "all" is in the AccountIds list
	for _, accountID := range result.AccountIds {
		if accountID == "all" {
			return true, nil
		}
	}

	return false, nil
}

// DisablePublicSharingForRegion disables public sharing at the region level
func DisablePublicSharingForRegion(ctx context.Context, ssmClient *ssm.Client, region string) error {
	settingId := "/ssm/documents/console/public-sharing-permission"

	input := &ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(settingId),
		SettingValue: aws.String("Disable"),
	}

	_, err := ssmClient.UpdateServiceSetting(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to disable public sharing for region %s: %w", region, err)
	}

	fmt.Printf("✅ Disabled public sharing at region level for %s\n", region)
	return nil
}

// MakeDocumentPrivate removes public sharing from an SSM document
func MakeDocumentPrivate(ctx context.Context, ssmClient *ssm.Client, documentName string) error {
	input := &ssm.ModifyDocumentPermissionInput{
		Name:               aws.String(documentName),
		PermissionType:     ssmTypes.DocumentPermissionTypeShare,
		AccountIdsToRemove: []string{"all"},
	}

	_, err := ssmClient.ModifyDocumentPermission(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to remove public access from %s: %w", documentName, err)
	}

	fmt.Printf("✅ Removed public access from document: %s\n", documentName)
	return nil
}

// FixSSMDocuments checks all documents, makes public ones private, then disables region-level sharing
func FixSSMDocuments(ctx context.Context, ssmClient *ssm.Client, region string, execute bool) ([]string, error) {
	publicDocuments, err := CheckAllDocumentsForPublicSharing(ctx, ssmClient)
	if err != nil {
		return nil, err
	}

	if len(publicDocuments) > 0 {
		fmt.Printf("\n⚠️  Found %d public SSM document(s) in %s:\n", len(publicDocuments), region)
		for idx, doc := range publicDocuments {
			fmt.Printf("%d. %s\n", idx+1, doc)
		}

		if execute {
			fmt.Printf("\nFix this region (%s)? ", region)
			if common.UserConfirmation() {
				fmt.Println("\nRemoving public access from documents...")
				failures := []string{}
				for _, doc := range publicDocuments {
					err := MakeDocumentPrivate(ctx, ssmClient, doc)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						failures = append(failures, doc)
					}
				}

				if len(failures) > 0 {
					fmt.Printf("\n⚠️  Failed to fix %d document(s):\n", len(failures))
					for _, doc := range failures {
						fmt.Printf("- %s\n", doc)
					}
					return publicDocuments, fmt.Errorf("failed to make all documents private")
				}
				// Continue to disable region-level sharing below
			} else {
				fmt.Printf("⏭️  Skipping %s - continuing with remaining regions...\n", region)
				return publicDocuments, nil
			}
		} else {
			fmt.Println("\nDry run mode - no changes made.")
			fmt.Println("Re-run with -execute flag to fix these documents and disable region-level public sharing.")
			return publicDocuments, nil
		}
	} else {
		fmt.Println("\n No public SSM documents found.")
	}

	// If we get here and execute is true, disable public sharing at region level
	if execute {
		fmt.Printf("\nDisabling public sharing at region level for %s...\n", region)
		err := DisablePublicSharingForRegion(ctx, ssmClient, region)
		if err != nil {
			return publicDocuments, err
		}
	} else {
		fmt.Printf("\nDry run: Would disable public sharing at region level for %s\n", region)
	}

	return publicDocuments, nil
}
