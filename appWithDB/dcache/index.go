package dcache

import (
	"fmt"

	"../derrors"
	"../fabclient"
)

// BlockDataHashIndex builds an index from a field on an object that is a
// string slice ([]string). Each value within the string slice can be used for
// lookup.
type BlockDataHashIndex struct {
}

func (s *BlockDataHashIndex) FromObject(obj interface{}) (bool, []byte, error) {
	v, ok := obj.(*fabclient.MiddleCommonBlock)
	if !ok {
		return false, nil, derrors.ErrorWrongObject
	}

	return true, v.DataHash, nil
}

func (s *BlockDataHashIndex) FromArgs(args ...interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}
	arg, ok := args[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("argument must be a string: %#v", args[0])
	}

	return arg, nil
}

func (s *BlockDataHashIndex) PrefixFromArgs(args ...interface{}) ([]byte, error) {
	val, err := s.FromArgs(args...)
	if err != nil {
		return nil, err
	}

	// Strip the null terminator, the rest is a prefix
	n := len(val)
	if n > 0 {
		return val[:n-1], nil
	}
	return val, nil
}
