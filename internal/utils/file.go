package utils

import (
	"os"
	"time"
)

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0770)
}

func GetMtime(filePath string) (time.Time, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}

	return stat.ModTime(), nil
}

func MustGetMtime(filePath string) time.Time {
	t, err := GetMtime(filePath)
	if err != nil {
		panic(err)
	}
	return t
}
