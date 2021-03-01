package chaincode

import (
	"fmt"
	"net/http"

	"../../fabclient"
	"../../web"
	"github.com/pkg/errors"
)

// CCPublicOperMake for normal handle makes
type CCPublicOperMake struct {
	path      string
	invoke    bool
	chaincode string
	function  string
	document  string
}

// NewCCPublicOperMake create one make
func NewCCPublicOperMake(path string, invoke bool, chaincode string, function string, document string) *CCPublicOperMake {
	return &CCPublicOperMake{path, invoke, chaincode, function, document}
}

// Path for web.ServiceHandle
func (make *CCPublicOperMake) Path() string {
	return make.path
}

// MakeChainCodeHandle for web.ServiceHandle
func (make *CCPublicOperMake) MakeChainCodeHandle() web.ChainCodeHandle {
	return func(args map[string][]string, body []byte) (id string, ret interface{}, err error) {
		_, ok := args["async"]

		objResp, err := fabclient.CallChainCode(ok, make.invoke, make.chaincode, make.function, [][]byte{body})
		if err != nil {
			return "", nil, err
		}

		if objResp.Success == false {
			err = errors.New(objResp.Message)
			return
		}

		id = objResp.TxID
		ret = objResp.Payload
		fmt.Println("=================ccQuery=====================")
		fmt.Println(ret)
		fmt.Println("=================ccQuery=====================")
		return
	}
}

// MakeHTTPHandle for web.ServiceHandle
func (make *CCPublicOperMake) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *CCPublicOperMake) Doc() string {
	return make.document
}
