package utils

import "runtime"

func GetNThreads() int {
	return runtime.GOMAXPROCS(0)
}
