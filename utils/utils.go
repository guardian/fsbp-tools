package utils

import (
	"flag"
	"fmt"
	"log"
)

type cliArgs struct {
	Profile     string
	Region      string
	DryRun      bool
	BucketCount int32
}

func ParseArgs() cliArgs {
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The region of the bucket")
	dryRun := flag.Bool("dry-run", true, "Dry run mode")
	bucketCount := flag.Int("max", 100, "The maximum number of buckets to attempt to process")
	flag.Parse()

	if *profile == "" {
		log.Fatal("Please provide a named AWS profile")
	}

	if *region == "" {
		log.Fatal("Please provide a region")
	}

	if *bucketCount < 1 || *bucketCount > 100 {
		log.Fatal("Please provide a max between 1 and 100")
	}

	return cliArgs{
		Profile:     *profile,
		Region:      *region,
		DryRun:      *dryRun,
		BucketCount: int32(*bucketCount),
	}
}

func Complement[T comparable](slice []T, toRemove []T) []T {
	var complement []T

	//put toRemove into a slice in a map for faster lookup
	removeMap := make(map[T]bool)
	for _, remove := range toRemove {
		removeMap[remove] = true
	}

	for _, element := range slice {
		_, found := removeMap[element]
		if !found {
			complement = append(complement, element)
		} else {
			fmt.Printf("\nExcluding: %v", element)
		}
	}
	fmt.Println("") //Tidy up the log output

	return complement
}
