package dservice

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"../dcache"
	"../dlog"
	"../dstore"
	"../fabclient"
	"../tools"
)

func syncBlockIntoDB(n time.Time) (err error) {
	pairs, err := getSyncEmptyZone()
	if err != nil || len(pairs) == 0 {
		return
	}

	logMark := uint64(n.Unix())

	TableList := make([]*fabclient.MiddleCommonBlock, 0, 50)

	logObj := dlog.DebugLog("pushBlockIntoDB", logMark)

	for _, pair := range pairs {
		for i := pair.Start; i <= pair.End; i++ {
			middle, err := fetchBlock("number", tools.UInt64ToString(i))
			if err != nil {
				continue
			}

			ok, err := isMiddleInDB(middle)
			if err != nil {
				return err
			}
			if ok {
				continue
			}

			TableList = append(TableList, middle)

			if len(TableList) >= 50 {
				err = pumpBlockIntoDBMore(TableList)
				if err != nil {
					logObj.WithError(err).Warning("sync block into DB failed")
				}
				TableList = make([]*fabclient.MiddleCommonBlock, 0, 50)
			}
		}

		logObj.WithField("start", pair.Start).WithField("end", pair.End).Info("push emptyzone into DB")
	}

	if len(TableList) > 0 {
		err = pumpBlockIntoDBMore(TableList)
		if err != nil {
			logObj.WithError(err).Warning("sync block into DB failed")
		}
	}

	err = nil

	return
}

func syncBlockIntoMemdb() (err error) {
	height, err := fetchHeight()
	start, end := stdStartEndWithHeight(0, 0, height)

	if end <= start {
		return nil
	}

	list := make([]*fabclient.MiddleCommonBlock, 0, end-start)
	for i := start; i < end; i++ {
		one, err := dcache.FetchBlock(i)
		if err == nil && one != nil {
			continue
		}

		one, err = fabclient.QueryBlockByHeight(i)
		if err != nil {
			continue
		}
		list = append(list, one)
	}
	if len(list) > 0 {
		err = dcache.PushBlockMore(list)
	}

	dcache.CheckSizeBlocks(start, end)

	return err
}

///// DB
const globalEmtpyZoneFormat = `select COLMNName, kind from
(
select COLMNName,1 as kind from (select COLMNName from TABLEName order by COLMNName asc) t where not exists (select 1 from TABLEName where COLMNName=t.COLMNName-1)
union
select COLMNName,2 as kind from (select COLMNName from TABLEName order by COLMNName asc) t where not exists (select 1 from TABLEName where COLMNName=t.COLMNName+1)
) t order by COLMNName, kind;`

type emptyZonePair struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
}

func getEmptyZoneByNumber(tableName, colmnName string) (ret []emptyZonePair, err error) {
	sql := globalEmtpyZoneFormat
	sql = strings.Replace(sql, "COLMNName", colmnName, -1)
	sql = strings.Replace(sql, "TABLEName", tableName, -1)

	var Height uint64
	var Kind uint64

	var last uint64
	var write uint64

	var cb = func() {
		switch Kind {
		case 1:
			// Less than 0
			// discard
			if Height == 0 {
				return
			}
			if last < write {
				last = write
			}
			ret = append(ret, emptyZonePair{
				Start: last,
				End:   Height - 1,
			})
			last = 0
			write = Height + 1
		case 2:
			if last == 0 {
				last = Height + 1
			}
		default:
			// never get here
		}
	}
	err = dstore.LoopExcuteROWs(cb, sql, &Height, &Kind)
	if err != nil {
		return
	}

	if 2 == Kind {
		ret = append(ret, emptyZonePair{
			Start: last,
			End:   0,
		})
	}
	return
}

func getSyncEmptyZone() (ret []emptyZonePair, err error) {
	high, err := fetchHeight()
	if err != nil {
		return
	}

	ret, err = getEmptyZoneByNumber(gorm.ToDBName("EventBlockTables"), gorm.ToDBName("Height"))
	if err != nil {
		return
	}

	if len(ret) == 0 {
		return
	}

	// [0, high)
	// but 0 block is not parsed
	// so
	// (0, high)
	length := len(ret)
	if ret[length-1].End == 0 {
		tmp := ret[length-1]
		if tmp.Start < high {
			tmp.End = high - 1
			ret[length-1] = tmp
		} else {
			ret = ret[:length-1]
		}
	}

	return
}

func isMiddleInDB(middle *fabclient.MiddleCommonBlock) (ok bool, err error) {
	if middle == nil {
		return
	}
	ret, err := fetchBlockFromDB("number", fmt.Sprintf("%v", middle.Number))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	ok = tools.Base64Encode(middle.DataHash) == ret.DataHash
	return
}
