package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func evaluateResult(t *testing.T, result []string, expected []string, message string) {
	if !reflect.DeepEqual(result, expected) {
		fmt.Println("Result: ")
		fmt.Println(result)
		fmt.Println("Expected: ")
		fmt.Println(expected)
		t.Errorf(message)
	}
}

func TestComplementOfEmptySlices(t *testing.T) {
	slice := []string{}
	toRemove := []string{}
	result := Complement(slice, toRemove)
	if len(result) != 0 {
		t.Errorf("Error computing complement where both slices are empty")
	}
}

func TestComplementOfEmptySlice(t *testing.T) {
	slice := []string{}
	toRemove := []string{"a", "b", "c"}
	result := Complement(slice, toRemove)
	if len(result) != 0 {
		t.Errorf("Error computing complement of empty slice")
	}
}

func TestComplementOfEmptyToRemove(t *testing.T) {
	slice := []string{"a", "b", "c"}
	toRemove := []string{}
	result := Complement(slice, toRemove)
	expected := []string{"a", "b", "c"}
	evaluateResult(t, result, expected, "Error computing complement of slice with empty toRemove")
}

func TestComplementOfNonEmptySlices(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	toRemove := []string{"b", "d"}
	result := Complement(slice, toRemove)
	expected := []string{"a", "c", "e"}
	evaluateResult(t, result, expected, "Error computing complement two non-empty slices")
}

func TestComplementOfNonEmptySlicesWithNoIntersection(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	toRemove := []string{"f", "g"}
	result := Complement(slice, toRemove)
	expected := []string{"a", "b", "c", "d", "e"}
	evaluateResult(t, result, expected, "Error computing complement of slice with no intersection")
}

func TestComplementOfSlicesWithPartialIntersection(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	toRemove := []string{"c", "d", "f"}
	result := Complement(slice, toRemove)
	expected := []string{"a", "b", "e"}
	evaluateResult(t, result, expected, "Error computing complement of slice with partial intersection")
}
