package utils

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

type cliArgs struct {
	Profile string
	Region  string
	DryRun  bool
}

func ParseArgs() cliArgs {
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The region of the bucket")
	dryRun := flag.Bool("dry-run", true, "Dry run mode")
	flag.Parse()

	if *profile == "" {
		log.Fatal("Please provide a named AWS profile")
	}

	if *region == "" {
		log.Fatal("Please provide a region")
	}

	return cliArgs{
		Profile: *profile,
		Region:  *region,
		DryRun:  *dryRun,
	}
}

func RemoveIndexFromSlice(slice []string, idx int) []string {

	if idx < 0 || idx >= len(slice) {
		fmt.Println("Index out of range, returning original slice")
		return slice
	}

	return append(slice[:idx], slice[idx+1:]...)
}

func RemoveElementsWithForbiddenSubstrings(slice []string, forbiddenSubstrings []string) []string {
	for idx, element := range slice {
		for _, forbiddenSubstring := range forbiddenSubstrings {
			containsSubstring := strings.Contains(element, forbiddenSubstring)
			if containsSubstring {
				fmt.Println("Removing " + element + " as it contains forbidden string: " + forbiddenSubstring)
				slice = RemoveIndexFromSlice(slice, idx)
				break
			}
		}

	}
	return slice
}
