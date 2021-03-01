package dloop

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"../derrors"
)

type workcore func(time.Time)

type serviceHold struct {
	quit      chan int
	waitGroup sync.WaitGroup
	lock      sync.Mutex
}

func newServiceHold() *serviceHold {
	return &serviceHold{
		quit: make(chan int),
	}
}

func (serv *serviceHold) decorateWithRecoverAndLog(input TimeHandle) func(time.Time) {
	return func(n time.Time) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithField("error", err).Warning("Panic occur in workShell")
			}
		}()

		err := input(n, serv.quit)
		if err != nil {
			logrus.WithField("error", err).Warning("work shell")
		}
	}
}

func (serv *serviceHold) run(input workPair) {
	defer serv.waitGroup.Done()

	one := serv.decorateWithRecoverAndLog(input.core)
	ticker := time.NewTicker(input.d)

Loop:
	for {
		select {
		case _, ok := <-serv.quit:
			if !ok {
				break Loop
			}
		case n, ok := <-ticker.C:
			if !ok {
				break Loop
			}
			one(n)
		}
	}
}

func (serv *serviceHold) Start(work []workPair) error {
	if len(work) == 0 {
		return derrors.ErrorEmptyValue
	}
	count := 0
	for _, one := range work {
		if one.d == 0 || one.core == nil {
			continue
		}

		serv.waitGroup.Add(1)
		go serv.run(one)

		count++
	}
	if count == 0 {
		return derrors.ErrorEmptyValue
	}
	return nil
}

func (serv *serviceHold) Quit() error {
	if serv.quit == nil {
		return derrors.ErrorTwiceCall
	}
	close(serv.quit)
	return nil
}

func (serv *serviceHold) Wait() error {
	serv.waitGroup.Wait()
	return nil
}

var globalService *serviceHold

func initService() {
	globalService = newServiceHold()
}

func getService() *serviceHold {
	return globalService
}
