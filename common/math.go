package common

import (
	"fmt"
)

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
			fmt.Printf("\nExcluding: '%v'", element)
		}
	}
	fmt.Println("") //This ensures sure the log output is tidy

	return complement
}
