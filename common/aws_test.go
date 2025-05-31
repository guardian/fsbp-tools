package common

import (
	"fmt"
	"testing"
)

type Input struct {
	NextToken *string
	PagesLeft *uint32
}

func PaginationStub(input *Input) (uint32, error) {
	*input.PagesLeft--
	if *input.PagesLeft == 0 {
		input.NextToken = nil
	}
	return *input.PagesLeft, nil

}

func TestMultiPagePagination(t *testing.T) {
	nextToken := "next"
	pagesLeft := uint32(5)
	input := &Input{
		NextToken: &nextToken,
		PagesLeft: &pagesLeft,
	}

	allItems, _ := Paginate(func(nextToken *string) ([]uint32, *string, error) {
		results, _ := PaginationStub(input)
		return []uint32{results}, input.NextToken, nil
	})

	fmt.Println("All items:", allItems)
	expected := 5
	actual := len(allItems)
	if actual != expected {
		t.Errorf("Paginated list should have %d items, but got %d.", expected, actual)
	}

}

func TestSinglePagePagination(t *testing.T) {
	pagesLeft := uint32(1)
	input := &Input{
		NextToken: nil,
		PagesLeft: &pagesLeft,
	}

	allItems, _ := Paginate(func(nextToken *string) ([]uint32, *string, error) {
		results, _ := PaginationStub(input)
		return []uint32{results}, input.NextToken, nil
	})

	fmt.Println("All items:", allItems)
	expected := 1
	actual := len(allItems)
	if actual != expected {
		t.Errorf("Paginated list should have %d items, but got %d.", expected, actual)
	}
}
