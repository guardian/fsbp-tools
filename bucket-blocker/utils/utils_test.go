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

func TestSplittingTrimmingEmptyString(t *testing.T) {
	input := ""
	result := SplitAndTrim(input)
	if len(result) != 0 {
		t.Errorf("Error splitting/trimming empty string")
	}
}

func TestStringSplit(t *testing.T) {
	input := "a,b,c"
	result := SplitAndTrim(input)
	expected := []string{"a", "b", "c"}
	evaluateResult(t, result, expected, "Error splitting string")
}

func TestStringTrim(t *testing.T) {
	input := " a     "
	result := SplitAndTrim(input)
	expected := []string{"a"}
	evaluateResult(t, result, expected, "Error trimming string")
}

func TestStringSplitAndTrim(t *testing.T) {
	input := " a, b    , c"
	result := SplitAndTrim(input)
	expected := []string{"a", "b", "c"}
	evaluateResult(t, result, expected, "Error splitting and trimming string")
}

func TestStringSplitAndTrimSpecialChars(t *testing.T) {
	input := "--a,.@/b,c123"
	result := SplitAndTrim(input)
	expected := []string{"--a", ".@/b", "c123"}
	evaluateResult(t, result, expected, "Error splitting and trimming string")
}

func TestSplittingStringWithLeadingAndTrailingComma(t *testing.T) {
	input := ",a,b,c,"
	result := SplitAndTrim(input)
	expected := []string{"a", "b", "c"}
	evaluateResult(t, result, expected, "Error splitting string with leading and trailing comma")
}
