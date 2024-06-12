package main

import (
	"reflect"
	"testing"
)

func TestRemoveElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice, _ = removeIndexFromSlice(slice, 2)

	if !reflect.DeepEqual(slice, []string{"a", "b", "d", "e"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveLastElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice, _ = removeIndexFromSlice(slice, 4)
	if !reflect.DeepEqual(slice, []string{"a", "b", "c", "d"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveFirstElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice, _ = removeIndexFromSlice(slice, 0)
	if !reflect.DeepEqual(slice, []string{"b", "c", "d", "e"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveNonExistingElementFromSlice(t *testing.T) {
	//create a slice
	slice := []string{"a", "b", "c", "d", "e"}
	//remove element "f" from slice
	_, err := removeIndexFromSlice(slice, 8)
	//check if the error is returned
	if err == nil {
		t.Errorf("Did not return error for non existing element in slice")
	}
}
