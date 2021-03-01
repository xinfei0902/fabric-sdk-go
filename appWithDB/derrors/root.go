package derrors

import (
	"github.com/pkg/errors"
)

//
var (
	ErrorSameKeyExist = errors.New("Same key is already exist")
	ErrorEmptyValue   = errors.New("Input value is empty")
	ErrorWrongObject  = errors.New("wrong object")

	ErrorEmptyPath = errors.New("Empty path")

	ErrorTwiceCall = errors.New("Twice call or call before init")

	ErrorNeedInit = errors.New("Need Init before call this")

	ErrorNotSupport = errors.New("Not support yet")

	ErrorNoDBConnected = errors.New("No DB")

	ErrorNeverGetHere = errors.New("never get here")
	ErrorPrivKey      = errors.New("not found priv login state")
)

func ErrorSameKeyExistf(key string) error {
	return errors.Errorf("Same key is already exist: %s", key)
}

func ErrorKeyNotContainValuef(key string) error {
	return errors.Errorf("Value is not available. Key: %s", key)
}
