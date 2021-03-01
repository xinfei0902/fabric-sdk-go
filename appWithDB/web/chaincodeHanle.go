package web

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"../derrors"
)

func logChainCode(r *http.Request, start, end int64, id string, err error, rec interface{}) {
	one := logrus.WithField("uri", r.RequestURI)
	one = one.WithField("method", r.Method)
	one = one.WithField("time", start)
	one = one.WithField("duration", end-start)
	one = one.WithField("remote", r.RemoteAddr)
	if rec != nil {
		one = one.WithField("recover", rec)
		one.Infoln("recover")
		return
	}

	if len(id) > 0 {
		one = one.WithField("id", id)
	}

	if err != nil {
		one.Infoln("failed")
		return
	}

	one.Infoln("success")
}

// ChainCodeHandle define chaincode work API
type ChainCodeHandle func(args map[string][]string, body []byte) (id string, ret interface{}, err error)

// ServiceHandle define interface of handleMakers
type ServiceHandle interface {
	Path() string
	MakeChainCodeHandle() ChainCodeHandle
	MakeHTTPHandle() http.HandlerFunc
	Doc() string
}

// ResiterMaker http handles API
func ResiterMaker(maker ServiceHandle) error {
	if maker == nil {
		return derrors.ErrorEmptyValue
	}

	path := maker.Path()
	if len(path) == 0 {
		return derrors.ErrorEmptyPath
	}

	core := maker.MakeChainCodeHandle()
	if core == nil {
		one := maker.MakeHTTPHandle()
		if one == nil {
			return derrors.ErrorEmptyValue
		}
		return PushHandleFunc(path, one)
	}

	hf := decorateCCRRToHF(path, core)

	return PushHandleFunc(path, hf)
}

func decorateCCRRToHF(path string, core ChainCodeHandle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// one, h, err := r.FormFile("file")

		var id string
		var ret interface{}
		var start int64
		var err error
		var end int64
		defer func() {
			rc := recover()
			logChainCode(r, start, end, id, err, rc)
		}()
		args, body := GetParamsBody(r)
		start = time.Now().Unix()
		id, ret, err = core(args, body)
		end = time.Now().Unix()

		one := logrus.WithField("uri", r.RequestURI)
		one = one.WithField("method", r.Method)
		one = one.WithField("start", start)
		one = one.WithField("end", end)
		one = one.WithField("remote", r.RemoteAddr)

		one = one.WithField("success", err == nil)
		one.Infoln("cc call")

		OutputEnter(w, id, ret, err)
	}
}
