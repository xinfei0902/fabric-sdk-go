package dservice

import (
	"time"

	"../dloop"
	"../web"
	"./chaincode"
	"github.com/pkg/errors"
)

// RegisterCCAPI for chaincode Web API
func RegisterCCAPI(debug bool, db bool) (err error) {
	if debug {
		for _, one := range chaincode.WebAPIForDebug() {
			err = web.ResiterMaker(one)
			if err != nil {
				err = errors.WithMessage(err, "Resiter API failed: "+one.Path())
				return
			}
		}
	}
	for _, one := range chaincode.WebAPI() {
		err = web.ResiterMaker(one)
		if err != nil {
			err = errors.WithMessage(err, "Resiter API failed: "+one.Path())
			return
		}
	}
	return nil
}

// RegisterWebAPI for regsiter Web Work APIs
func RegisterWebAPI(debug bool, db bool) (err error) {
	// start / stop debug

	// with / without db
	list := []httpHandlePair{
		{"/block/height", makeBlockHeight},
		{"/block/info", makeBlockInfo},
		{"/block/range", makeBlockRange},
		{"/tran/detail", makeTranscationDetail},
		{"/peers", makePeers},
		{"/channel", makeChannel},
		{"/chaincode", makeChainCode},
		{"/test", makeTest},

		{"/report/peer", makePeerReport},

		{"/trace/user", makeUserTrace},
	}

	if db {
		list = append(list, httpHandlePair{"/block/sync", makeBlockSync})
	}

	if debug {
		err = web.AddDebugServer("")
		if err != nil {
			return
		}
	}

	// API
	for _, one := range list {
		f := one.f(debug, db)
		if f == nil {
			continue
		}
		err = web.PushHandleFunc(one.p, f)
		if err != nil {
			return
		}
	}

	return nil
}

// RegisterBackend for regsiter Timeservice or backend loop work jobs
func RegisterBackend(debug bool, db bool) (err error) {
	// start / stop debug

	// with / without db

	// Running
	for _, pair := range []timeHandlePair{
		{1 * time.Second, makeEventPumpHandle},
		{30 * time.Second, makeSyncHandle},
		{10 * time.Second, makePeerHandle},
		{30 * time.Second, makeMailHandle},
	} {
		f := pair.f(debug, db)
		if f == nil {
			continue
		}

		err = dloop.RegisterHandle(pair.d, f)
		if err != nil {
			return
		}
	}

	return nil
}
