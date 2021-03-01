package web

import (
	"fmt"
	"testing"
)

func Test_AddPath(t *testing.T) {

	AddDebugServer("/a")

	for k, v := range globalHandles {
		fmt.Println(k, "-|=", v)
	}

	for k, v := range globalHandleFuncs {
		fmt.Println(k, "-|-", v)
	}
}
