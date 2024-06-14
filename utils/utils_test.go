package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRemoveElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice = RemoveIndexFromSlice(slice, 2)

	if !reflect.DeepEqual(slice, []string{"a", "b", "d", "e"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveLastElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice = RemoveIndexFromSlice(slice, 4)
	if !reflect.DeepEqual(slice, []string{"a", "b", "c", "d"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveFirstElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice = RemoveIndexFromSlice(slice, 0)
	if !reflect.DeepEqual(slice, []string{"b", "c", "d", "e"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveNonExistingElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(RemoveIndexFromSlice(slice, 5), slice) {
		t.Errorf("Did not return original slice when index out of range")
	}
}

func TestRemoveElementsWithForbiddenSubstrings(t *testing.T) {
	slice := []string{"aa", "bb", "cc", "dd", "ee"}
	forbiddenSubstrings := []string{"c", "e"}
	filteredSlice := RemoveElementsWithForbiddenSubstrings(slice, forbiddenSubstrings)
	if !reflect.DeepEqual(filteredSlice, []string{"aa", "bb", "dd"}) {
		fmt.Println(slice)
		fmt.Println(filteredSlice)
		t.Errorf("Error removing elements from slice")
	}
}
