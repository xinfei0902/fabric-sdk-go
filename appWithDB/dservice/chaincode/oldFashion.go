package chaincode

import (
	"net/http"

	"github.com/pkg/errors"

	"../../fabclient"
	"../../web"
)

// OldFashionMake for normal handle makes
type OldFashionMake struct {
	path     string
	invoke   bool
	node     string
	function string
	key      string
	document string
}

// NewOldFashionMake create one make
func NewOldFashionMake(path string, invoke bool, node, function, key, document string) *OldFashionMake {
	return &OldFashionMake{
		path, invoke, node, function, key, document,
	}
}

// Path for web.ServiceHandle
func (make *OldFashionMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *OldFashionMake) MakeChainCodeHandle() web.ChainCodeHandle {
	call := "query"
	if make.invoke {
		call = "invoke"
	}
	if len(make.key) > 0 {

		return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
			var value []byte

			v, ok := args[make.key]
			if ok && len(v) > 0 && len(v[0]) > 0 {
				value = []byte(v[0])
			}

			_, ok = args["async"]

			objResp, err := fabclient.CallChainCode(ok, make.invoke, make.node, call, [][]byte{[]byte(make.function), value})
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
		objResp, err := fabclient.CallChainCode(ok, make.invoke, make.node, call, [][]byte{[]byte(make.function), body})
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
func (make *OldFashionMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *OldFashionMake) Doc() string {
	return make.document
}
