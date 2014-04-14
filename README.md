gomsg
=====

Pack and UnPack Msg

Msg
------

<pre>
  <code>
  // 消息的结构
  type Msg struct {
  	Type    int32  // 消息类型
  	Size    int32  // 消息大小（含消息类型和消息大小自身
  	Content []byte // 消息正文
  }
  </code>
</pre>


Pack(mType int32, mContent []byte) []byte
------------------------------------------

use Pack to pack struct Msg to bytes for Conn.Write().


func UnPack(b []byte) (Msg, error) 
-----------------------------------

use UnPack to unpack msg from bytes which read from Conn.Read().
