package dservice

import (
	"../convert"
	"../dcache"
	"../derrors"
	"../dstore"
	"../fabclient"
	"../tools"
)

func fetchPeers(peers []string) (ret []fabclient.SystemInformationStruct, err error) {
	tmp, err := dcache.GetPeers(peers)
	if err != nil || tmp == nil {
		tmp = []fabclient.SystemInformationStruct{}
	}

	if len(tmp) >= len(peers) {
		ret = tmp
		return
	}

	if len(tmp) > 0 {
		for _, one := range tmp {
			peers = removePeers(peers, one.PeerAddress)
		}
	}
	tmp2, err := fabclient.QuerySystemInfomationFromPeer(peers)
	ret = append(tmp, tmp2...)
	return
}

func fetchBlockFromMemdb(key, value string) (*fabclient.MiddleCommonBlock, error) {
	switch key {
	case "number":
		n, err := tools.StringToUInt64(value)
		if err != nil {
			return nil, err
		}
		return dcache.FetchBlock(n)
	case "txid":
		return dcache.FetchBlockTxID(value)
	case "hash":
		buff, err := tools.Base64Decode([]byte(value))
		if err != nil {
			return nil, err
		}
		return dcache.FetchBlockHash(buff)
	default:
		// not here
	}
	return nil, derrors.ErrorNeverGetHere
}

func fetchBlockFromDB(key, value string) (*convert.EventBlockTable, error) {
	switch key {
	case "number":
		n, err := tools.StringToUInt64(value)
		if err != nil {
			return nil, err
		}
		return dstore.FetchBlock(n)
	case "txid":
		return dstore.FetchBlockTxID(value)
	case "hash":
		return dstore.FetchBlockHash(value)
	default:
		// not here
	}
	return nil, derrors.ErrorNeverGetHere
}

func fetchBlockRangeFromDB(start, end uint64) ([]*convert.EventBlockTable, error) {
	return dstore.FetchBlockRange(start, end)
}

func fetchTranscationFromDB(txid string) (*convert.TranInBlockTable, error) {
	return dstore.FetchTranscations(txid)
}

func fetchHeight() (uint64, error) {
	peers := fabclient.QueryPeersTargets(nil)

	ret, err := fetchPeers(peers)
	if err != nil {
		return 0, err
	}
	if len(ret) == 0 {
		return 0, nil
	}

	var max uint64
	for _, one := range ret {
		if max < one.Height {
			max = one.Height
		}
	}
	return max, nil
}

func fetchTranscation(txid string) (*fabclient.MiddleTranNode, error) {
	_, ret, err := fabclient.QueryTransaction(txid)
	return ret, err
}

func fetchBlock(key, value string) (*fabclient.MiddleCommonBlock, error) {
	return fabclient.QueryBlockByKV(key, value)
}

func fetchBlockRange(start, end uint64) ([]*fabclient.MiddleCommonBlock, error) {
	if end <= start {
		return nil, nil
	}
	ret := make([]*fabclient.MiddleCommonBlock, 0, end-start)
	var lastErr error
	for i := start; i < end; i++ {
		one, err := dcache.FetchBlock(i)
		if err == nil && one != nil {
			ret = append(ret, one)
			continue
		}

		one, err = fabclient.QueryBlockByHeight(i)
		if err != nil {
			lastErr = err
			continue
		}
		ret = append(ret, one)
	}

	if len(ret) == 0 {
		return nil, lastErr
	}
	return ret, nil
}

func stdStartEndWithHeight(start, end, height uint64) (s, e uint64) {
	if end > height || end == 0 {
		end = height
		if end > 10 {
			start = end - 10
		} else {
			start = 1
		}

	} else if start == 0 {
		if end > 10 {
			start = end - 10
		} else {
			start = 1
		}
	} else {
		end = start + 10
	}
	return start, end
}

func fetchUserTraceFromDB(address string, start, end int64) ([]*convert.UserTrace, error) {
	if end <= 0 || start <= 0 || end < start || len(address) == 0 {
		return nil, derrors.ErrorEmptyValue
	}

	return dstore.FetchUserTrace(address, start, end)
}
