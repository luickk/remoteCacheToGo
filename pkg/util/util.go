package util

import (
  "encoding/json"
  "io"
  "bufio"
  "net"
  "encoding/binary"
)

// serializable pp req
type SPushPullReq struct {
	Key string
	Operation string
	Data []byte
}

func EncodeMsg(msg *SPushPullReq) ([]byte, error) {
	serializedMsg, err := json.Marshal(&msg)
	if err != nil {
		return []byte{}, err
	}
	return serializedMsg, nil
}


func DecodeMsg(msg *SPushPullReq, data []byte) error {
	err := json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	return nil
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

func WriteFrame(writer *bufio.Writer, data []byte) error {
	dataLen := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataLen, uint64(len(data)))
	_, err := writer.Write(dataLen)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func readFixedSize(r io.Reader, len int) (b []byte, err error) {
    b = make([]byte, len)
    _, err = io.ReadFull(r, b)
    if err != nil {
        return
    }
    return
}

func ReadFrame(conn net.Conn) ([]byte, error) {
	lenData, err := readFixedSize(conn, 8)
	if err != nil {
		return []byte{}, err
	}
	dataLen := binary.LittleEndian.Uint64(lenData)

	dataBuffer, err := readFixedSize(conn, int(dataLen))
	if err != nil {
		return []byte{}, err
	}
	return dataBuffer, nil
}
