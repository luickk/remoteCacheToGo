package util

import (
  "encoding/json"
	"unicode"
)

// serializable pp req
// cache.conn.Write(append(append([]byte(ppCacheOp.Key + "-<-"), ppCacheOp.Data...), []byte("\rnr")...))
type SPushPullReq struct {
	Key string
	Operation string
	Data []byte
}



func EncodePushPullReq(sPushPullReq *SPushPullReq) ([]byte, error) {
	serializedPPR, err := json.Marshal(&sPushPullReq)
	if err != nil {
		return []byte{}, err
	}
	return serializedPPR, nil
}


func DecodePushPullReq(ppr *SPushPullReq, data []byte) error {
	err := json.Unmarshal(data, ppr)
	if err != nil {
		return err
	}
	return nil
}

// checks wether input consists only of characters
func CharacterWhiteList(input string) bool {
    for _, r := range input {
        if unicode.IsLetter(r) || unicode.IsNumber(r) {
            return true
        }
    }
    return false
}

// slice operation
// from https://stackoverflow.com/questions/8307478/how-to-find-out-element-position-in-slice
func Index(limit int, predicate func(i int) bool) int {
    for i := 0; i < limit; i++ {
        if predicate(i) {
            return i
        }
    }
    return -1
}
