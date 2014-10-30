// 读取连接，解析出消息头，返回消息正文和消息类型
package msg

import (
	"bytes"           // bytes.NewBuffer
	"encoding/binary" // binary.Read
	"errors"          // errors.New

	"io"  // io.EOF
	"net" // TCP

	"time" // time.Sleep

	"fmt"
)

const (
	MAX_BUFFER   = 1024 // 读取缓存最大值
	SIZE_OF_TYPE = 4    // sizeof int32
	SIZE_OF_SIZE = 4    // sizeof int32
	SIZE_OF_HEAD = SIZE_OF_TYPE + SIZE_OF_SIZE
)

const (
	MAX_BUFFER_BIG   = 2147483647
	SIZE_OF_SIZE_BIG = 8
)

// 消息的结构
type Msg struct {
	Type    int32  // 消息类型
	Size    int32  // 消息大小（含消息类型和消息大小自身
	Content []byte // 消息正文
}

// 用以存储的消息结构
type MsgBig struct {
	Size    int32  //消息大小
	Content []byte //消息正文
}

// 解包消息，返回Msg类型
func UnPack(b []byte) (Msg, error) {
	m := Msg{}
	buf := bytes.NewBuffer(b)
	// 消息类型
	mType := buf.Next(SIZE_OF_TYPE)
	bufType := bytes.NewBuffer(mType)
	binary.Read(bufType, binary.LittleEndian, &m.Type)
	// 消息大小
	mSize := buf.Next(SIZE_OF_SIZE)
	bufSize := bytes.NewBuffer(mSize)
	binary.Read(bufSize, binary.LittleEndian, &m.Size)
	// 超限则返回错误
	if m.Size > MAX_BUFFER {
		return m, errors.New("OVER_MAX_BUFFER")
	}
	// 消息正文
	mContent := buf.Bytes()
	rest := int(m.Size - int32(SIZE_OF_HEAD))
	if rest > 0 {
		if rest > len(mContent)-1 {
			m.Content = mContent
		} else {
			m.Content = mContent[:rest]
		}
	}
	return m, nil
}

// 解包，不限制大小（用于消息存储
// func UnPackUnLimited(b []byte) (Msg, error) {
// 	m := Msg{}
// 	buf := bytes.NewBuffer(b)
// 	// 消息类型
// 	mType := buf.Next(SIZE_OF_TYPE)
// 	bufType := bytes.NewBuffer(mType)
// 	binary.Read(bufType, binary.LittleEndian, &m.Type)
// 	// 消息大小
// 	mSize := buf.Next(SIZE_OF_SIZE)
// 	bufSize := bytes.NewBuffer(mSize)
// 	binary.Read(bufSize, binary.LittleEndian, &m.Size)
// 	// // 超限则返回错误
// 	// if m.Size > MAX_BUFFER {
// 	// 	return m, errors.New("OVER_MAX_BUFFER")
// 	// }
// 	// 消息正文
// 	mContent := buf.Bytes()
// 	rest := int(m.Size - int32(SIZE_OF_HEAD))
// 	if rest > 0 {
// 		if rest > len(mContent)-1 {
// 			m.Content = mContent
// 		} else {
// 			m.Content = mContent[:rest]
// 		}
// 	}
// 	return m, nil
// }

// 打包消息，返回[]byte
func Pack(mType int32, mContent []byte) []byte {
	buf := new(bytes.Buffer)
	// 消息类型
	binary.Write(buf, binary.LittleEndian, mType)
	// 消息大小
	mSize := int32(SIZE_OF_HEAD + len(mContent))
	binary.Write(buf, binary.LittleEndian, mSize)
	// 消息正文
	binary.Write(buf, binary.LittleEndian, mContent)
	b := buf.Bytes()
	return b
}

func MsgRequest(addr net.TCPAddr, b []byte) *net.TCPConn {
	conn, e := net.DialTCP("tcp", nil, &addr)
	if e != nil {
		fmt.Printf("SingleRequest.DialTCP:%v", e)
		return nil
	}
	defer conn.Close()

	SingleWrite(conn, b)

	// m := SingleRead(conn)
	return conn
}

func SingleRequest(addr net.TCPAddr, b []byte) Msg {
	conn, e := net.DialTCP("tcp", nil, &addr)
	if e != nil {
		fmt.Printf("SingleRequest.DialTCP:%v", e)
		return Msg{}
	}
	defer conn.Close()

	SingleWrite(conn, b)

	m := SingleRead(conn)
	return m
}

func SingleWrite(conn *net.TCPConn, b []byte) []byte {
	conn.Write(b)
	return b
}

func SingleRead(conn *net.TCPConn) Msg {

	m := Msg{}

	b := make([]byte, SIZE_OF_HEAD)
	for { // 循环到读取到内容为止
		i, e := conn.Read(b)
		if e != nil && e != io.EOF { // 网络有错,则退出循环
			fmt.Printf("msg.SingleRead:%v", e)
			return Msg{}
		}

		if i > 0 { // 读到内容则退出读取循环
			break
		}
		time.Sleep(50 * time.Microsecond)
	}

	buf := bytes.NewBuffer(b)

	// 消息类型
	mType := buf.Next(SIZE_OF_TYPE)
	bufType := bytes.NewBuffer(mType)
	binary.Read(bufType, binary.LittleEndian, &m.Type)

	// 消息大小
	mSize := buf.Next(SIZE_OF_SIZE)
	bufSize := bytes.NewBuffer(mSize)
	binary.Read(bufSize, binary.LittleEndian, &m.Size)

	size := int(m.Size) - SIZE_OF_HEAD
	if size <= 0 {
		return m
	}

	b = make([]byte, size)
	_, e := conn.Read(b)
	if e != nil && e != io.EOF { // 网络有错,则退出循环
		fmt.Printf("msg.SingleRead:%v", e)
		return Msg{}
	}
	m.Content = b

	return m
}

func CopyBytes(a, b []byte) []byte {
	n := len(a)
	result := make([]byte, n+len(b))
	copy(result, a)
	copy(result[n:], b)
	return result
}

func PackMsgBig(mContent []byte) []byte {
	buf := new(bytes.Buffer)
	// 消息大小
	mSize := int32(SIZE_OF_SIZE_BIG + len(mContent))
	binary.Write(buf, binary.LittleEndian, mSize)
	// 消息正文
	binary.Write(buf, binary.LittleEndian, mContent)
	b := buf.Bytes()
	return b
}

func UnpackMsgBig(b []byte) (MsgBig, error) {
	m := MsgBig{}
	buf := bytes.NewBuffer(b)
	// 消息大小
	mSize := buf.Next(SIZE_OF_SIZE_BIG)
	bufSize := bytes.NewBuffer(mSize)
	binary.Read(bufSize, binary.LittleEndian, &m.Size)
	// 超限则返回错误
	if m.Size > MAX_BUFFER_BIG {
		return m, errors.New("OVER_MAX_BUFFER_BIG")
	}
	// 消息正文
	mContent := buf.Bytes()
	rest := int(m.Size - int32(SIZE_OF_SIZE_BIG))
	if rest > 0 {
		if rest > len(mContent)-1 {
			m.Content = mContent
		} else {
			m.Content = mContent[:rest]
		}
	}
	return m, nil
}
