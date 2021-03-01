package web

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"../djson"
	"../tools"
)

// responseObj return object
type responseObj struct {
	Success bool        `json:"success"`
	TxID    string      `json:"txid,omitempty"`
	Payload interface{} `json:"data,omitempty"`
	Msg     string      `json:"msg,omitempty"`
}

// outputJSONObj write obj by JSON stream
func outputJSONObj(w http.ResponseWriter, obj interface{}) {
	buff, err := djson.Marshal(obj)
	if err != nil {
		logrus.Debugln("OutputInvokeObj", "json failed", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(buff)

	logrus.Debugln("Reponse done")
}

// outputFailed by msg
func outputFailed(w http.ResponseWriter, msg string) {
	one := responseObj{
		Success: false,
		Msg:     msg,
	}
	outputJSONObj(w, &one)
}

// outputSuccess by TxID and payload
func outputSuccess(w http.ResponseWriter, id string, payload interface{}) {
	ret := responseObj{
		Success: true,
	}
	if len(id) > 0 {
		ret.TxID = id
	}
	switch payload.(type) {
	case []byte:
		tmp := payload.([]byte)
		if len(tmp) > 0 {
			obj, err := tools.TryParseStringToObj(tmp)
			if err != nil {
				ret.Payload = string(tmp)
			} else {
				ret.Payload = obj
			}
		}
	default:
		if payload != nil {
			ret.Payload = payload
		}
	}
	outputJSONObj(w, &ret)
}

func OutputEnter(w http.ResponseWriter, id string, payload interface{}, msg error) {
	if msg != nil {
		outputFailed(w, msg.Error())
		return
	}
	outputSuccess(w, id, payload)
}

func GetParamsBody(r *http.Request) (map[string][]string, []byte) {
	r.ParseForm()
	ret := r.PostForm
	if len(r.Form) > 0 {
		for k, value := range r.Form {
			if len(value) == 0 {
				continue
			}
			for _, v := range value {
				ret.Add(k, v)
			}
		}
	}
	if r.Body == nil {
		return ret, nil
	}
	defer r.Body.Close()

	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Debugln("read body all")
	}
	return ret, buff
}

func GetParamToList(args map[string][]string, key string, sep string) (value []string) {
	key = strings.ToLower(strings.TrimSpace(key))

	value = make([]string, 0, len(args))

	for k, v := range args {
		k = strings.ToLower(strings.TrimSpace(k))
		if key != k {
			continue
		}

		if len(v) == 0 {
			continue
		}

		for _, one := range v {
			if len(one) == 0 {
				continue
			}
			pairs := strings.Split(one, sep)
			for _, two := range pairs {
				if len(two) == 0 {
					continue
				}

				value = append(value, two)
			}
		}
	}

	return
}

func GetKVbyParams(args map[string][]string, keylist []string) (key string, value []string, exist bool) {
	if len(args) == 0 {
		return
	}
	if len(keylist) == 0 {
		return
	}
	for _, one := range keylist {
		value, exist = args[one]
		if exist && len(value) > 0 {
			key = one
			return
		}
	}
	return
}
