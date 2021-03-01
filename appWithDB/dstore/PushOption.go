package dstore

import "../convert"

// PushOption in output protocol
type PushOption interface {
	Begin() error
	PushOne(one *convert.EventBlockTable) error
	Commit() error
	Abort() error
}

// MakePusher function to create push
type MakePusher func() PushOption
