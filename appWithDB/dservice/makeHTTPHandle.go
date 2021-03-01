package dservice

import (
	"net/http"
	"static"

	"../convert"
	"../derrors"
	"../fabclient"
	"../tools"
	"../web"
)

type makeHTTPHandle func(debug, db bool) http.HandlerFunc

type httpHandlePair struct {
	p string
	f makeHTTPHandle
}

func makeBlockHeight(debug, db bool) http.HandlerFunc {
	// all
	// read from peers memdb
	// if no exist
	// read from block
	return func(w http.ResponseWriter, r *http.Request) {
		args, _ := web.GetParamsBody(r)
		v := web.GetParamToList(args, "org", ",")
		var peers []string
		if len(v) > 0 {
			peers = fabclient.QueryPeersTargets(v)
		}

		ret, err := fetchPeers(peers)

		web.OutputEnter(w, "", ret, err)
	}
}

func makeBlockInfo(debug, db bool) http.HandlerFunc {
	// if db
	// read from blocks memdb
	// if no exist
	// read from db
	// if no exist
	// read from block

	// if no db
	// read from blocks memdb
	// if no exist
	// read from block

	var core func(key, value string, remote bool) (*convert.BlockShortResponse, error)

	if db {
		core = func(key, value string, remote bool) (ret *convert.BlockShortResponse, err error) {
			if false == remote {
				tmp, err := fetchBlockFromMemdb(key, value)
				if err == nil && tmp != nil {
					ret = convert.EventBlockToShortResponse(tmp)
					return ret, nil
				}
			}
			if false == remote {
				tmp, err := fetchBlockFromDB(key, value)
				if err == nil && tmp != nil {
					ret = convert.EventTableToShortResponse(tmp)
					return ret, nil
				}
			}
			tmp, err := fetchBlock(key, value)
			if err != nil {
				return
			}
			ret = convert.EventBlockToShortResponse(tmp)
			return
		}
	} else {
		core = func(key, value string, remote bool) (ret *convert.BlockShortResponse, err error) {
			if false == remote {
				tmp, err := fetchBlockFromMemdb(key, value)
				if err == nil && tmp != nil {
					ret = convert.EventBlockToShortResponse(tmp)
					return ret, nil
				}
			}

			tmp, err := fetchBlock(key, value)
			if err != nil {
				return
			}
			ret = convert.EventBlockToShortResponse(tmp)
			return
		}
	}

	var keys = []string{"number", "txid", "hash"}
	return func(w http.ResponseWriter, r *http.Request) {
		args, _ := web.GetParamsBody(r)

		key, value, exist := web.GetKVbyParams(args, keys)
		if false == exist || len(value) == 0 || len(value[0]) == 0 {
			web.OutputEnter(w, "", nil, derrors.ErrorEmptyValue)
			return
		}

		remote := web.GetParamToList(args, "remote", ",")
		b := len(remote) > 0 && len(remote[0]) > 0 && remote[0] == "true"

		ret, err := core(key, value[0], b)

		web.OutputEnter(w, "", ret, err)
	}
}

func makeBlockSync(debug, db bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ret, err := getSyncEmptyZone()
		web.OutputEnter(w, "", ret, err)
	}
}

func makeBlockRange(debug, db bool) http.HandlerFunc {
	// if db
	// read from db
	// if no exist
	// read from block

	// if no db
	// read from block
	var core func(start, end uint64, remote bool) ([]*convert.BlockShortResponse, error)

	if db {
		core = func(start, end uint64, remote bool) ([]*convert.BlockShortResponse, error) {
			if false == remote {
				tmp, err := fetchBlockRangeFromDB(start, end)
				if err == nil && len(tmp) > 0 {
					ret := make([]*convert.BlockShortResponse, 0, len(tmp))
					for _, one := range tmp {
						two := convert.EventTableToShortResponse(one)
						ret = append(ret, two)
					}
					return ret, nil
				}
			}

			tmp, err := fetchBlockRange(start, end)
			if err != nil {
				return nil, err
			}
			ret := make([]*convert.BlockShortResponse, 0, len(tmp)+1)

			for _, one := range tmp {
				two := convert.EventBlockToShortResponse(one)
				ret = append(ret, two)
			}
			return ret, nil
		}
	} else {
		core = func(start, end uint64, remote bool) ([]*convert.BlockShortResponse, error) {
			tmp, err := fetchBlockRange(start, end)
			if err != nil {
				return nil, err
			}
			ret := make([]*convert.BlockShortResponse, 0, len(tmp)+1)

			for _, one := range tmp {
				two := convert.EventBlockToShortResponse(one)
				ret = append(ret, two)
			}
			return ret, nil
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		args, _ := web.GetParamsBody(r)

		var err error
		var start, end uint64
		value := web.GetParamToList(args, "start", ",")
		if len(value) > 0 && len(value[0]) > 0 {
			start, err = tools.StringToUInt64(value[0])
			if err != nil {
				start = 0
			}
		}

		value = web.GetParamToList(args, "end", ",")
		if len(value) > 0 && len(value[0]) > 0 {
			end, err = tools.StringToUInt64(value[0])
			if err != nil {
				end = 0
			}
		}

		remote := web.GetParamToList(args, "remote", ",")
		b := len(remote) > 0 && len(remote[0]) > 0 && remote[0] == "true"

		height, err := fetchHeight()
		if err != nil {
			web.OutputEnter(w, "", nil, err)
			return
		}

		start, end = stdStartEndWithHeight(start, end, height)

		ret, err := core(start, end, b)

		web.OutputEnter(w, "", ret, err)
	}
}

func makeTranscationDetail(debug, db bool) http.HandlerFunc {
	// if db
	// read from db
	// if no exist
	// read from block

	// if no db
	// read from block

	var core func(txid string, remote bool) (*convert.TranResponse, error)

	if db {
		core = func(txid string, remote bool) (*convert.TranResponse, error) {
			if false == remote {
				tmp, err := fetchTranscationFromDB(txid)
				if err == nil && tmp != nil {
					one := convert.TransactionsTableToResponse([]convert.TranInBlockTable{
						*tmp,
					})
					ret := &one[0]
					return ret, nil
				}
			}

			// memdb

			tmp, err := fetchTranscation(txid)
			if err != nil {
				return nil, err
			}
			one := convert.TransactionsBlockToResponse([]fabclient.MiddleTranNode{
				*tmp,
			})
			ret := &one[0]
			return ret, nil
		}
	} else {
		core = func(txid string, remote bool) (*convert.TranResponse, error) {

			// memdb

			tmp, err := fetchTranscation(txid)
			if err != nil {
				return nil, err
			}
			one := convert.TransactionsBlockToResponse([]fabclient.MiddleTranNode{
				*tmp,
			})
			ret := &one[0]
			return ret, nil
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		args, _ := web.GetParamsBody(r)

		txid := web.GetParamToList(args, "txid", ",")
		if len(txid) == 0 || len(txid[0]) == 0 {
			web.OutputEnter(w, "", nil, derrors.ErrorEmptyValue)
			return
		}

		remote := web.GetParamToList(args, "remote", ",")
		b := len(remote) > 0 && len(remote[0]) > 0 && remote[0] == "true"

		ret, err := core(txid[0], b)
		web.OutputEnter(w, "", ret, err)
	}
}

func makePeers(debug, db bool) http.HandlerFunc {
	// all
	// read from peers memdb
	// if no exist
	// read from peers
	return func(w http.ResponseWriter, r *http.Request) {
		args, _ := web.GetParamsBody(r)
		v := web.GetParamToList(args, "org", ",")
		var peers []string
		if len(v) > 0 {
			peers = fabclient.QueryPeersTargets(v)
		}

		ret, err := fetchPeers(peers)

		web.OutputEnter(w, "", ret, err)
	}
}

func makeTest(debug, db bool) http.HandlerFunc {
	// all
	// read from peers memdb
	// if no exist
	// read from peers
	banner := []byte(static.Banner("Welcome!"))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body.Close()
		}
		w.Write(banner)
	}
}

func makeChannel(debug, db bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body.Close()
		}

		obj, err := fabclient.QueryChannel()
		web.OutputEnter(w, "", obj, err)
	}
}

func makeChainCode(debug, db bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body.Close()
		}
		obj, err := fabclient.QueryChainCodeList()
		web.OutputEnter(w, "", obj, err)
	}
}
