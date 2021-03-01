package dservice

import (
	"encoding/json"
	"net/http"

	"../convert"
	"../dstore"
	"../web"
)

func makePeerReport(debug, db bool) http.HandlerFunc {
	if db == false {
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, body := web.GetParamsBody(r)

		dbObj := &convert.PeerSystemInformation{}
		err := json.Unmarshal(body, dbObj)
		if err != nil {
			web.OutputEnter(w, "", nil, err)
			return
		}

		if len(dbObj.Nodes) > 0 {
			mailList := make([]convert.ChainType, 0, len(dbObj.Nodes))
			for _, one := range dbObj.Nodes {
				if len(one.Error) > 0 {
					mailList = append(mailList, one)
				}
			}
			PushMail(mailList)
		}

		err = dstore.PushPeerInfomation(dbObj)
		web.OutputEnter(w, "", nil, err)
		return
	}
}
