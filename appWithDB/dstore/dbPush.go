package dstore

import (
	"../derrors"
	"github.com/jinzhu/gorm"
)

// DBPusher for DB Push
type DBPusher struct {
	opt *gorm.DB
}

// NewDBPusher for work
func NewDBPusher() *DBPusher {
	return &DBPusher{}
}

// Begin DB push job
func (p *DBPusher) Begin() error {
	if globalDBOpt == nil {
		return derrors.ErrorNeedInit
	}
	p.opt = globalDBOpt.Begin()
	err := p.opt.Error
	if err != nil {
		p.opt = nil
	}
	return err
}

// PushOne one line into DB
func (p *DBPusher) PushOne(one interface{}) error {
	if p.opt == nil {
		return nil
	}
	err := p.opt.Create(one).Error
	if err != nil {
		p.opt.Rollback()
		p.opt = nil
	}
	return err
}

// Commit job
func (p *DBPusher) Commit() error {
	if p.opt == nil {
		return nil
	}
	return p.opt.Commit().Error
}

// Abort job
func (p *DBPusher) Abort() error {
	if p.opt == nil {
		return nil
	}
	err := p.opt.Rollback().Error
	p.opt = nil
	return err
}
