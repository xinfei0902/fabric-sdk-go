package dcache

import (
	"../fabclient"
)

// SetPeers into cache
func SetPeers(input []fabclient.SystemInformationStruct) (err error) {
	txn := globalOpt.Txn(true)
	defer txn.Commit()

	_, err = txn.DeleteAll(TablePeerKey, AddressPeerKey)
	if err != nil {
		txn.Abort()
		return
	}

	if len(input) == 0 {
		return
	}

	for _, one := range input {
		err = txn.Insert(TablePeerKey, one)
		if err != nil {
			txn.Abort()
			return
		}
	}

	return
}

// GetPeers from cache
func GetPeers(targets []string) (ret []fabclient.SystemInformationStruct, err error) {

	txn := globalOpt.Txn(false)
	defer txn.Abort()

	ret = make([]fabclient.SystemInformationStruct, 0, len(targets))

	// get all
	if len(targets) == 0 {
		it, err := txn.Get(TablePeerKey, AddressPeerKey)
		if err != nil {
			return nil, err
		}

		ret = make([]fabclient.SystemInformationStruct, 0, 4)
		for v := it.Next(); v != nil; v = it.Next() {
			one, ok := v.(fabclient.SystemInformationStruct)
			if !ok {
				continue
			}
			ret = append(ret, one)
		}

		return ret, nil
	}

	for _, t := range targets {
		v, err := txn.First(TablePeerKey, AddressPeerKey, t)
		if err != nil {
			return nil, err
		}
		one, ok := v.(fabclient.SystemInformationStruct)
		if !ok {
			continue
		}
		ret = append(ret, one)
	}
	return
}
