package dservice

import (
	"encoding/json"
	"fmt"
	"static/mail"
	"sync"
	"time"

	"../convert"
	"../dconfig"
	"../dlog"
	"../dloop"
)

type globalMailList struct {
	Lock sync.RWMutex
	core []convert.ChainType
}

var globalMailOpt *globalMailList

func initGlobalMail() {
	globalMailOpt = &globalMailList{
		core: make([]convert.ChainType, 0, 64),
	}
}

func PushMail(nodes []convert.ChainType) {
	if len(nodes) == 0 {
		return
	}
	globalMailOpt.Lock.Lock()
	globalMailOpt.core = append(globalMailOpt.core, nodes...)
	globalMailOpt.Lock.Unlock()
}

func makeMailHandle(debug, db bool) dloop.TimeHandle {
	target := dconfig.GetStringByKey("mailtarget")
	if len(target) == 0 {
		return nil
	}

	return func(n time.Time, quit <-chan int) (err error) {

		globalMailOpt.Lock.Lock()
		if len(globalMailOpt.core) == 0 {
			globalMailOpt.Lock.Unlock()
			return nil
		}
		one := make([]convert.ChainType, 0, 64)
		one, globalMailOpt.core = globalMailOpt.core, one
		globalMailOpt.Lock.Unlock()

		t := time.Now()
		subject := fmt.Sprintf("Log(%d) at %v", len(one), t)

		buff, err := json.Marshal(one)
		if err != nil {
			dlog.DebugLog("mail", uint64(t.Unix())).WithField("data", one).WithError(err).Warning("marshal json")
			return
		}
		req := mail.NewEnvelope("MailReporter", subject, string(buff))
		req.PushDestination(target)
		req.HeadSet("Content-Type", "text/plain; charset=UTF-8")

		err = mail.SendMail(req)
		if err != nil {
			dlog.DebugLog("mail", uint64(t.Unix())).WithField("data", one).WithError(err).Warning("send mail failed")
			return
		}

		dlog.Info(fmt.Sprintf("mail count: %d", len(one)))

		return nil
	}
}
