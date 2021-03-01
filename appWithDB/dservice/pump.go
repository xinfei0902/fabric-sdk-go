package dservice

import (
	"../convert"
	"../dcache"
	"../dstore"
	"../fabclient"

	"./chaincode"
)

func removePeers(peers []string, mark string) (ret []string) {
	if len(peers) == 0 {
		return
	}
	last := len(peers) - 1
	for i, one := range peers {
		if one == mark {
			if last == 0 {
				return
			}
			peers[i] = peers[last]
			ret = peers[0:last]
			return
		}
	}
	ret = peers
	return
}

func pumpPeersIntoCache(peers []string) (err error) {
	ret, err := fabclient.QuerySystemInfomationFromPeer(peers)
	if err != nil {
		return
	}

	err = dcache.SetPeers(ret)
	return
}

func pumpBlockIntoMemdb(block *fabclient.MiddleCommonBlock) error {
	return dcache.PushBlock(block)
}

func pumpBlockIntoDB(block *fabclient.MiddleCommonBlock) (err error) {
	if block == nil {
		return
	}

	one := convert.EventBlockToTable(block)
	return dstore.PushOne(one, chaincode.PushOption()...)
}

func pumpBlockIntoDBMore(blocks []*fabclient.MiddleCommonBlock) (err error) {
	if len(blocks) == 0 {
		return
	}
	ones := make([]*convert.EventBlockTable, len(blocks))
	for i, middle := range blocks {
		ones[i] = convert.EventBlockToTable(middle)
	}
	return dstore.PushMore(ones, chaincode.PushOption()...)
}
