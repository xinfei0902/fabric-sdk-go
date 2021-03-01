package web

import (
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strings"
	"time"

	"../derrors"
	"github.com/sirupsen/logrus"
)

var globalHandles map[string]http.Handler

var globalHandleFuncs map[string]http.HandlerFunc

func initHandle() {
	globalHandles = make(map[string]http.Handler)
	globalHandleFuncs = make(map[string]http.HandlerFunc)
}

func middleLevelHead(time int64, r *http.Request) {
	r.ParseForm()
	one := logrus.WithField("uri", r.RequestURI)
	one = one.WithField("method", r.Method)
	one = one.WithField("time", time)
	one = one.WithField("remote", r.RemoteAddr)
	// if len(r.Form) > 0 {
	// 	one = one.WithField("param", r.Form)
	// }
	// if len(r.PostForm) > 0 {
	// 	one = one.WithField("post_param", r.PostForm)
	// }

	one.Infoln("enter")
}

func middleLevelTail(now int64, r *http.Request) {
	one := logrus.WithField("uri", r.RequestURI)
	one = one.WithField("method", r.Method)
	one = one.WithField("time", time.Now().Unix())
	one = one.WithField("enter", now)
	one = one.WithField("remote", r.RemoteAddr)
	// if len(r.Form) > 0 {
	// 	one = one.WithField("param", r.Form)
	// }
	// if len(r.PostForm) > 0 {
	// 	one = one.WithField("post_param", r.PostForm)
	// }
	one.Infoln("leave")
}

// PushHandle into http default router
func PushHandle(path string, h http.Handler) error {
	if len(path) == 0 || h == nil {
		return derrors.ErrorEmptyValue
	}
	path = httpPathJoin(path)
	_, ok := globalHandles[path]
	if ok {
		return derrors.ErrorSameKeyExist
	}

	globalHandles[path] = h
	return nil
}

// PushHandleFunc into http default router
func PushHandleFunc(path string, hf http.HandlerFunc) error {
	if len(path) == 0 || hf == nil {
		return derrors.ErrorEmptyValue
	}

	one := func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().Unix()
		middleLevelHead(now, r)
		hf(w, r)
		middleLevelTail(now, r)
	}
	return pushHandleCore(path, one)
}

func pushHandleCore(path string, hf http.HandlerFunc) error {
	if len(path) == 0 || hf == nil {
		return derrors.ErrorEmptyValue
	}
	path = httpPathJoin(path)
	_, ok := globalHandleFuncs[path]
	if ok {
		return derrors.ErrorSameKeyExist
	}
	globalHandleFuncs[path] = hf
	return nil
}

// AddDebugServer into http default router
func AddDebugServer(path string) (err error) {
	if len(path) == 0 {
		path = "/debug/pprof"
	}

	err = PushHandleFunc(httpPathJoin(path), pprof.Index)
	if err != nil {
		return
	}
	err = PushHandleFunc(httpPathJoin(path, "profile"), pprof.Profile)
	err = PushHandleFunc(httpPathJoin(path, "symbol"), pprof.Symbol)
	err = PushHandleFunc(httpPathJoin(path, "trace"), pprof.Trace)
	err = PushHandleFunc(httpPathJoin(path, "cmdline"), pprof.Cmdline)

	for _, v := range []string{"heap", "goroutine", "block", "threadcreate"} {
		err = PushHandle(httpPathJoin(path, v), pprof.Handler(v))
		if err != nil {
			return
		}
	}

	return
}

func baseSignUPHandle() *http.ServeMux {
	// clear http handles
	ret := http.NewServeMux()

	for k, v := range globalHandleFuncs {
		ret.HandleFunc(k, v)
	}

	for k, v := range globalHandles {
		ret.Handle(k, v)
	}

	return ret
}

func httpPathJoin(elem ...string) string {
	return strings.Replace(filepath.Join(elem...), "\\", "/", -1)
}
