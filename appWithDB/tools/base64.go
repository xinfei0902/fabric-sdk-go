package tools

import "encoding/base64"

func Base64Decode(input []byte) (ret []byte, err error) {
	if len(input) == 0 {
		return
	}
	ret = make([]byte, base64.StdEncoding.DecodedLen(len(input)))

	n, err := base64.StdEncoding.Decode(ret, input)
	if err != nil {
		return
	}
	ret = ret[0:n]
	return
}

func Base64Encode(input []byte) (ret string) {
	if len(input) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(input)
}
