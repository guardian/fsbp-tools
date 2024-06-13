package utils

import (
	"reflect"
	"testing"
)

func TestRemoveElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice, _ = RemoveIndexFromSlice(slice, 2)

	if !reflect.DeepEqual(slice, []string{"a", "b", "d", "e"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveLastElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice, _ = RemoveIndexFromSlice(slice, 4)
	if !reflect.DeepEqual(slice, []string{"a", "b", "c", "d"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveFirstElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	slice, _ = RemoveIndexFromSlice(slice, 0)
	if !reflect.DeepEqual(slice, []string{"b", "c", "d", "e"}) {
		t.Errorf("Error removing element from slice")
	}
}

func TestRemoveNonExistingElementFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	_, err := RemoveIndexFromSlice(slice, 8)
	if err == nil {
		t.Errorf("Did not return error for non existing element in slice")
	}
}
