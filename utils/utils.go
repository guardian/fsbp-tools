package utils

import (
	"flag"
	"fmt"
	"log"
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

func Complement[T comparable](slice []T, toRemove []T) []T {
	var complement []T

	for _, element := range slice {
		found := false
		for _, remove := range toRemove {
			if element == remove {
				fmt.Printf("\nExcluding: %v", element)
				found = true
				break
			}
		}
		if !found {
			complement = append(complement, element)
		}
	}
	fmt.Println("") //Tidy up the log output

	return complement
}
