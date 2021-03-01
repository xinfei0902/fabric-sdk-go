package dservice

import (
	"time"

	"../dloop"
	"../fabclient"
)

type makeTimeHandle func(debug, db bool) dloop.TimeHandle

type timeHandlePair struct {
	d time.Duration
	f makeTimeHandle
}

func makeEventPumpHandle(debug, db bool) dloop.TimeHandle {
	// if db
	// pump event into db

	// all
	// pump event into peers memdb
	// pump blocks into blocks memdb
	peers := fabclient.QueryPeersTargets(nil)

	var core func(input *fabclient.MiddleCommonBlock)

	if db {
		core = func(input *fabclient.MiddleCommonBlock) {
			pumpBlockIntoDB(input)

			// all
			pumpPeersIntoCache(peers)
			pumpBlockIntoMemdb(input)
		}
	} else {
		core = func(input *fabclient.MiddleCommonBlock) {

			// all
			pumpPeersIntoCache(peers)
			pumpBlockIntoMemdb(input)
		}
	}

	return func(n time.Time, quit <-chan int) (err error) {
		return fabclient.LoopBlockEvent(core, quit)
	}
}

func makeSyncHandle(debug, db bool) dloop.TimeHandle {
	// all
	// pump blocks into blocks memdb

	// if db
	// pump blocks into db
	if db {
		return func(n time.Time, quit <-chan int) (err error) {
			syncBlockIntoMemdb()
			syncBlockIntoDB(n)
			return nil
		}
	}
	return func(n time.Time, quit <-chan int) (err error) {
		syncBlockIntoMemdb()
		return nil
	}

}

func makePeerHandle(debug, db bool) dloop.TimeHandle {
	// all
	// pump peers into peers memdb

	peers := fabclient.QueryPeersTargets(nil)

	return func(n time.Time, quit <-chan int) (err error) {
		return pumpPeersIntoCache(peers)
	}
}
