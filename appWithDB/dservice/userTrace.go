package dservice

import (
	"encoding/json"
	"net/http"

	"../web"
)

type UserTraceRequest struct {
	Address string `json:"address"`
	Start   int64  `json:"start"`
	End     int64  `json:"end"`
}

func makeUserTrace(debug, db bool) http.HandlerFunc {
	if db == false {
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		_, body := web.GetParamsBody(r)

		input := &UserTraceRequest{}
		err := json.Unmarshal(body, input)
		if err != nil {
			web.OutputEnter(w, "", nil, err)
			return
		}

		ret, err := fetchUserTraceFromDB(input.Address, input.Start, input.End)

		web.OutputEnter(w, "", ret, err)
		return
	}
}
