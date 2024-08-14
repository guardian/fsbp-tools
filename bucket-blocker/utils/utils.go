package utils

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

type cliArgs struct {
	Profile     string
	Region      string
	Execute     bool
	BucketCount int32
	Exclusions  []string
}

func ParseArgs() cliArgs {
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The region of the bucket")
	execute := flag.Bool("execute", false, "Execute the block operation")
	bucketCount := flag.Int("max", 100, "The maximum number of buckets to attempt to process")
	exclusions := flag.String("exclusions", "", "Comma-separated list of buckets to skip")

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
		Execute:     *execute,
		BucketCount: int32(*bucketCount),
		Exclusions:  SplitAndTrim(*exclusions),
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

func SplitAndTrim(str string) []string {
	split := strings.Split(str, ",")
	var trimmed []string
	for _, s := range split {
		s := strings.Trim(s, " ")
		trimmed = append(trimmed, s)
	}
	return Complement(trimmed, []string{""})
}
