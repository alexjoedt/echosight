package osutils

import (
	"errors"
	"os"
)

func IsFileExists(fp string) bool {
	_, err := os.Stat(fp)
	if err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}

	return true
}
