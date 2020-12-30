package util

import (
  "encoding/json"
	"unicode"
  "errors"
  "bytes"
  "encoding/binary"
)

// serializable pp req
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

func Padd(paddSize int, inp []byte) []byte {
  if len(inp) > paddSize {
    return []byte{}
  }
  padded := make([]byte, paddSize-len(inp))
  return append(inp, padded...)
}

func RemovePadding(inp []byte, fixedHeaderSize int) ([]byte, error) {
  if len (inp) > 0 {
    // since json string are null byte terminatee we can just trim all null bytes
    return  bytes.Trim(inp, "\x00"), nil
  } else {
    return []byte{}, errors.New("input not large enough")
  }
}

func readInt64(b []byte) (int64, error) {
  buf := bytes.NewBuffer(b)
  res, err := binary.ReadVarint(buf)
  if err != nil {
    return 0, err
  }
  return res, nil
}
