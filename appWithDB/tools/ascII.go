package tools

func IsAscII(c byte) bool {
	return c >= 32 && c <= 126
}
