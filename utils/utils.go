package utils

import "errors"

func RemoveIndexFromSlice(slice []string, idx int) ([]string, error) {

	if idx < 0 || idx >= len(slice) {
		return slice, errors.New("index out of range")
	}

	return append(slice[:idx], slice[idx+1:]...), nil
}
