package chaincode

import (
	"net/http"

	"github.com/pkg/errors"

	"../../fabclient"
	"../../web"
)

// SingleKeyMake for normal handle makes
type SingleKeyMake struct {
	path     string
	invoke   bool
	node     string
	function string
	key      string
	document string
}

// NewSingleKeyMake create one make
func NewSingleKeyMake(path string, invoke bool, node, function, key, document string) *SingleKeyMake {
	return &SingleKeyMake{
		path, invoke, node, function, key, document,
	}
}

// Path for web.ServiceHandle
func (make *SingleKeyMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *SingleKeyMake) MakeChainCodeHandle() web.ChainCodeHandle {
	if len(make.key) > 0 {
		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			var key []byte

			v, ok := args[make.key]
			if ok && len(v) > 0 && len(v[0]) > 0 {
				key = []byte(v[0])
			}

			_, ok = args["async"]

			objResp, err := fabclient.CallChainCode(ok, make.invoke, make.node, make.function, [][]byte{key, body})
			if err != nil {
				return "", nil, err
			}

			if objResp.Success == false {
				err = errors.New(objResp.Message)
				return
			}

			id = objResp.TxID
			ret = objResp.Payload
			return
		}
	}
	return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
		_, ok := args["async"]
		objResp, err := fabclient.CallChainCode(ok, make.invoke, make.node, make.function, [][]byte{body})
		if err != nil {
			return "", nil, err
		}

		if objResp.Success == false {
			err = errors.New(objResp.Message)
			return
		}

		id = objResp.TxID
		ret = objResp.Payload
		return
	}
}

// MakeHTTPHandle for web.ServiceHandle
func (make *SingleKeyMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *SingleKeyMake) Doc() string {
	return make.document
}
