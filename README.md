gomsg
=====

  // 消息的结构
  type Msg struct {
  	Type    int32  // 消息类型
  	Size    int32  // 消息大小（含消息类型和消息大小自身
  	Content []byte // 消息正文
  }

Pack and UnPack Msg

use Pack to pack struct Msg to bytes for Conn.Write().

use UnPack to unpack msg from bytes which read from Conn.Read().
