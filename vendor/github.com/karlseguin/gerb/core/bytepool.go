package core

import (
	"gopkg.in/karlseguin/bytepool.v3"
)

var BytePool = bytepool.New(65536, 64)
