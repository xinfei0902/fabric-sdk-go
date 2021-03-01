package tools

import (
	"crypto/md5"
	"encoding/hex"
)

// EncodeSumHashMD5 sum one string hash
func EncodeSumHashMD5(input string) (string, error) {
	op := md5.New()
	buff := []byte(input)

	for offset := 0; offset < len(buff); {
		step, err := op.Write(buff[offset:])
		if err != nil {
			return "", err
		}
		if step == 0 {
			break
		}
		offset += step
	}

	return hex.EncodeToString(op.Sum(nil)), nil
}

// SumHashMD5 sum one string hash
func SumHashMD5(input string) ([]byte, error) {
	op := md5.New()
	buff := []byte(input)

	for offset := 0; offset < len(buff); {
		step, err := op.Write(buff[offset:])
		if err != nil {
			return nil, err
		}
		if step == 0 {
			break
		}
		offset += step
	}

	return op.Sum(nil), nil
}
