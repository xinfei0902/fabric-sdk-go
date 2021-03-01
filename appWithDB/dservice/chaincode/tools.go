package chaincode

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"../../derrors"
	"../../fabclient"
	"../../web"
)

// ToolsCC for normal handle makes
type ToolsCC struct {
	path     string
	document string
}

// Path for web.ServiceHandle
func (make *ToolsCC) Path() string {
	return make.path
}

type ToolsInput struct {
	Async     bool     `json:"async"`
	Chaincode string   `json:"ccname"`
	Invoke    bool     `json:"invoke"`
	Function  string   `json:"function"`
	Args      []string `json:"args"`
}

// MakeChainCodeHandle for web.ServiceHandle
func (opt *ToolsCC) MakeChainCodeHandle() web.ChainCodeHandle {
	return func(_ map[string][]string, body []byte) (id string, ret interface{}, err error) {
		if len(body) == 0 {
			err = derrors.ErrorEmptyValue
			return
		}

		input := &ToolsInput{}
		err = json.Unmarshal(body, input)
		if err != nil {
			return
		}
		if len(input.Chaincode) == 0 || len(input.Function) == 0 {
			err = derrors.ErrorEmptyValue
			return
		}

		args := make([][]byte, len(input.Args))
		for i, v := range input.Args {
			args[i] = []byte(v)
		}
		if len(args) == 0 {
			args = nil
		}

		objResp, err := fabclient.CallChainCode(input.Async, input.Invoke, input.Chaincode, input.Function, args)
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
func (make *ToolsCC) MakeHTTPHandle() http.HandlerFunc {
	return nil
}

// Doc for web.ServiceHandle
func (make *ToolsCC) Doc() string {
	return make.document
}
