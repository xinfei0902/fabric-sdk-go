package dstore

import (
	"github.com/pkg/errors"

	"../convert"
	"../dlog"
)

// PushOne block into DB
func PushOne(input *convert.EventBlockTable, opt ...PushOption) (err error) {
	pusher := NewDBPusher()
	err = pusher.Begin()
	if err != nil {
		return err
	}

	err = pusher.PushOne(input)
	if err != nil {
		pusher.Abort()
		for _, one := range opt {
			one.Abort()
		}
		return err
	}

	err = pusher.Commit()
	if err != nil {

		for _, one := range opt {
			one.Abort()
		}
		return err
	}

	for _, one := range opt {
		err = one.Begin()
		if err != nil {
			dlog.Error(errors.WithMessage(err, "pushOne option err"))
		}
	}

	for _, one := range opt {
		err = one.PushOne(input)
		if err != nil {
			one.Abort()
			continue
		}

		err = one.Commit()
	}
	return nil
}

// PushMore blocks into DB
func PushMore(ones []*convert.EventBlockTable, opts ...PushOption) (err error) {
	pusher := NewDBPusher()
	err = pusher.Begin()
	if err != nil {
		return err
	}

	for _, opt := range opts {
		err = opt.Begin()
		if err != nil {
			dlog.Error(errors.WithMessage(err, "pushMore option err"))
		}
	}

	for _, one := range ones {
		err = pusher.PushOne(one)
		if err != nil {
			pusher.Abort()
			for _, opt := range opts {
				opt.Abort()
			}
			return err
		}

		for _, opt := range opts {
			err = opt.PushOne(one)
			if err != nil {
				dlog.Error(errors.WithMessage(err, "pushMore option err"))
				opt.Abort()
			}
		}
	}

	err = pusher.Commit()
	if err != nil {
		for _, opt := range opts {
			opt.Abort()
		}
		return
	}
	for _, opt := range opts {
		err = opt.Commit()
		if err != nil {
			dlog.Error(errors.WithMessage(err, "pushMore option err"))
			opt.Abort()
		}
	}
	return nil
}
