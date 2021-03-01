package dcache

import (
	"../fabclient"
)

// PushBlock into cache
func PushBlock(one *fabclient.MiddleCommonBlock) error {
	txn := globalOpt.Txn(true)
	defer txn.Commit()

	err := txn.Insert(TableBlockKey, one)
	if err != nil {
		txn.Abort()
		return err
	}

	return nil
}

// PushBlockMore into cache
func PushBlockMore(ones []*fabclient.MiddleCommonBlock) error {
	txn := globalOpt.Txn(true)
	defer txn.Commit()

	for _, one := range ones {
		err := txn.Insert(TableBlockKey, one)
		if err != nil {
			txn.Abort()
			return err
		}
	}

	return nil
}

// CheckSizeBlocks then delete other block from cache
func CheckSizeBlocks(start, end uint64) error {
	txn := globalOpt.Txn(true)
	defer txn.Commit()

	list := make([]interface{}, 0, 64)

	it, err := txn.Get(TableBlockKey, HeightBlockKey)
	if err != nil {
		txn.Abort()
		return err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		v, ok := obj.(*fabclient.MiddleCommonBlock)
		if !ok {
			continue
		}

		if v.Number < start || v.Number > end {
			list = append(list, obj)
		}
	}

	for _, one := range list {
		err = txn.Delete(TableBlockKey, one)
		if err != nil {
			txn.Abort()
			return err
		}
	}

	return nil
}

// FetchBlock from cache
func FetchBlock(height uint64) (*fabclient.MiddleCommonBlock, error) {
	txn := globalOpt.Txn(false)
	defer txn.Abort()

	one, err := txn.First(TableBlockKey, HeightBlockKey, height)
	if err != nil {
		return nil, err
	}

	if one == nil {
		return nil, nil
	}

	ret, ok := one.(*fabclient.MiddleCommonBlock)
	if !ok {
		return nil, nil
	}

	return ret, nil
}

// FetchBlockHash from cache
func FetchBlockHash(hash []byte) (*fabclient.MiddleCommonBlock, error) {
	txn := globalOpt.Txn(false)
	defer txn.Abort()

	one, err := txn.First(TableBlockKey, HashBlockKey, hash)
	if err != nil {
		return nil, err
	}

	if one == nil {
		return nil, nil
	}

	ret, ok := one.(*fabclient.MiddleCommonBlock)
	if !ok {
		return nil, nil
	}

	return ret, nil
}

// FetchBlockTxID  not support yet
func FetchBlockTxID(txid string) (*fabclient.MiddleCommonBlock, error) {
	// Todo here
	// Not support yet
	return nil, nil
}
