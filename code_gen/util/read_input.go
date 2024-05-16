package util

import (
	"io/ioutil"
	"os"
)

func ReadInput() ([]byte, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ReadFile(path string) (string, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	return string(bs), nil
}
