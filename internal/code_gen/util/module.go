package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func SearchGoMod(cwd string) (moduleName, path string, found bool) {
	for {
		path = filepath.Join(cwd, "go.mod")
		data, err := ioutil.ReadFile(path)
		if err == nil {
			re := regexp.MustCompile(`^\s*module\s+(\S+)\s*`)
			for _, line := range strings.Split(string(data), "\n") {
				m := re.FindStringSubmatch(line)
				if m != nil {
					return m[1], cwd, true
				}
			}
			return fmt.Sprintf("<module name not found in '%s'>", path), path, true
		}

		if !os.IsNotExist(err) {
			return
		}
		parentCwd := filepath.Dir(cwd)
		if parentCwd == cwd {
			break
		}
		cwd = parentCwd
	}
	return
}
