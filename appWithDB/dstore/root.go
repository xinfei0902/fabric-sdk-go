package dstore

import (
	"../convert"
	"../derrors"
)

func init() {
}

// Init Output service
func Init(kind KindStore, params map[string]interface{}) (err error) {
	switch kind {
	case PostgresDBKind:
		v, ok := params[KeyLinkString]
		if !ok {
			return derrors.ErrorEmptyValue
		}
		link, ok := v.(string)
		if !ok || len(link) == 0 {
			return derrors.ErrorEmptyValue
		}

		err = newDBClientPool(PostgresDBKind.String(), link)
		if err != nil {
			return
		}
		err = migrateTables(convert.GetBlocksTables())
	default:
		err = derrors.ErrorNotSupport
	}

	return
}
