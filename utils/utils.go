package utils

import (
	"errors"
	"flag"
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

func RemoveIndexFromSlice(slice []string, idx uint) ([]string, error) {

	if int(idx) >= len(slice) {
		return slice, errors.New("index out of range")
	}

	return append(slice[:idx], slice[idx+1:]...), nil
}
