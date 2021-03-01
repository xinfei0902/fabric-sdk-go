package chaincode

import (
	"../../convert"
	"../../dstore"
)

type UserTrace struct {
	Core *dstore.DBPusher

	Count int64
}

func (opt *UserTrace) Begin() error {
	return opt.Core.Begin()
}

func (opt *UserTrace) PushOne(one *convert.EventBlockTable) error {
	two, ok := convert.MiddleToUserTrace(one)
	if !ok {
		return nil
	}

	for _, three := range two {
		err := opt.Core.PushOne(three)
		if err != nil {
			opt.Count = 0
			return err
		}
		opt.Count++
	}

	return nil
}

func (opt *UserTrace) Commit() error {
	if opt.Count == 0 {
		opt.Core.Abort()
		return nil
	}

	return opt.Core.Commit()
}

func (opt *UserTrace) Abort() error {
	return opt.Core.Abort()
}

func NewUserTrace() dstore.PushOption {
	return &UserTrace{
		Core: dstore.NewDBPusher(),
	}
}
