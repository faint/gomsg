// Package msg 读取连接，解析出消息头，返回消息正文和消息类型
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
	maxBuffer = 1024 // 读取缓存最大值
	sizeType  = 4    // sizeof int32
	sizeSize  = 4    // sizeof int32
	sizeHead  = sizeType + sizeSize
)

const (
	maxBufferBIG = 2147483647
	sizeSizeBIG  = 8
)

// Msg 消息的结构
type Msg struct {
	Type    int32  // 消息类型
	Size    int32  // 消息大小（含消息类型和消息大小自身
	Content []byte // 消息正文
}

// Big 用以存储的消息结构
type Big struct {
	Size    int32  //消息大小
	Content []byte //消息正文
}

// UnPack 解包消息，返回Msg类型
func UnPack(b []byte) (Msg, error) {
	m := Msg{}
	buf := bytes.NewBuffer(b)
	// 消息类型
	mType := buf.Next(sizeType)
	bufType := bytes.NewBuffer(mType)
	binary.Read(bufType, binary.LittleEndian, &m.Type)
	// 消息大小
	mSize := buf.Next(sizeSize)
	bufSize := bytes.NewBuffer(mSize)
	binary.Read(bufSize, binary.LittleEndian, &m.Size)
	// 超限则返回错误
	if m.Size > maxBuffer {
		return m, errors.New("OVER_maxBuffer")
	}
	// 消息正文
	mContent := buf.Bytes()
	rest := int(m.Size - int32(sizeHead))
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
// 	mType := buf.Next(sizeType)
// 	bufType := bytes.NewBuffer(mType)
// 	binary.Read(bufType, binary.LittleEndian, &m.Type)
// 	// 消息大小
// 	mSize := buf.Next(sizeSize)
// 	bufSize := bytes.NewBuffer(mSize)
// 	binary.Read(bufSize, binary.LittleEndian, &m.Size)
// 	// // 超限则返回错误
// 	// if m.Size > maxBuffer {
// 	// 	return m, errors.New("OVER_maxBuffer")
// 	// }
// 	// 消息正文
// 	mContent := buf.Bytes()
// 	rest := int(m.Size - int32(sizeHead))
// 	if rest > 0 {
// 		if rest > len(mContent)-1 {
// 			m.Content = mContent
// 		} else {
// 			m.Content = mContent[:rest]
// 		}
// 	}
// 	return m, nil
// }

// Pack 打包消息，返回[]byte
func Pack(mType int32, mContent []byte) []byte {
	buf := new(bytes.Buffer)
	// 消息类型
	binary.Write(buf, binary.LittleEndian, mType)
	// 消息大小
	mSize := int32(sizeHead + len(mContent))
	binary.Write(buf, binary.LittleEndian, mSize)
	// 消息正文
	binary.Write(buf, binary.LittleEndian, mContent)
	b := buf.Bytes()
	return b
}

// Request ...
func Request(addr net.TCPAddr, b []byte) *net.TCPConn {
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

// SingleRequest ...
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

// SingleWrite ...
func SingleWrite(conn *net.TCPConn, b []byte) []byte {
	conn.Write(b)
	return b
}

// SingleRead ...
func SingleRead(conn *net.TCPConn) Msg {

	m := Msg{}

	b := make([]byte, sizeHead)
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
	mType := buf.Next(sizeType)
	bufType := bytes.NewBuffer(mType)
	binary.Read(bufType, binary.LittleEndian, &m.Type)

	// 消息大小
	mSize := buf.Next(sizeSize)
	bufSize := bytes.NewBuffer(mSize)
	binary.Read(bufSize, binary.LittleEndian, &m.Size)

	size := int(m.Size) - sizeHead
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

// CopyBytes ...
func CopyBytes(a, b []byte) []byte {
	n := len(a)
	result := make([]byte, n+len(b))
	copy(result, a)
	copy(result[n:], b)
	return result
}

// PackBig ...
func PackBig(mContent []byte) []byte {
	buf := new(bytes.Buffer)
	// 消息大小
	mSize := int64(sizeSizeBIG + len(mContent))
	binary.Write(buf, binary.LittleEndian, mSize)
	// 消息正文
	binary.Write(buf, binary.LittleEndian, mContent)
	b := buf.Bytes()
	return b
}

// UnpackBig ...
func UnpackBig(b []byte) (Big, error) {
	m := Big{}
	buf := bytes.NewBuffer(b)
	// 消息大小
	mSize := buf.Next(sizeSizeBIG)
	bufSize := bytes.NewBuffer(mSize)
	binary.Read(bufSize, binary.LittleEndian, &m.Size)
	// 超限则返回错误
	if m.Size > maxBufferBIG {
		return m, errors.New("OVER_maxBufferBIG")
	}
	// 消息正文
	mContent := buf.Bytes()
	rest := int(m.Size - int32(sizeSizeBIG))
	if rest > 0 {
		if rest > len(mContent)-1 {
			m.Content = mContent
		} else {
			m.Content = mContent[:rest]
		}
	}
	return m, nil
}
