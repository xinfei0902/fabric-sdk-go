package tools

import (
	"fmt"
	"strconv"
)

func StringToUInt64(input string) (uint64, error) {
	return strconv.ParseUint(input, 0, 0)
}

func UInt64ToString(input uint64) string {
	return fmt.Sprintf("%v", input)
}
