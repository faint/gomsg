// 单次与指定TCPAddr交换数据
package msg

import (
	"io"   // io.EOF
	"net"  // TCP
	"time" // time.Sleep
)

func SingleRequest(addr net.TCPAddr, b []byte) []byte {
	conn, e := net.DialTCP("tcp", nil, &addr)
	if e != nil {
		return []byte{}
	}
	defer conn.Close()

	conn.Write(b)

	read := false
	for !read { // 读取回执
		i, e := conn.Read(b)
		if i == 0 || e == io.EOF { // 读空则睡眠，让出循环
			time.Sleep(50 * time.Millisecond)
			continue
		}

		return b
		read = true // 设置已读
	}
	return []byte{}
}
