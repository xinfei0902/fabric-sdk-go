package dstore

import (
	"../convert"
)

// PushPeerInfomation into DB
func PushPeerInfomation(input *convert.PeerSystemInformation) (err error) {
	if input == nil {
		return
	}

	one := NewDBPusher()
	err = one.Begin()
	if err != nil {
		return
	}
	err = one.PushOne(input)
	if err != nil {
		return
	}
	err = one.Commit()
	return
}
