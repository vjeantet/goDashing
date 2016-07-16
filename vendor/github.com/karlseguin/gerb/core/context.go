package core

import (
	"gopkg.in/karlseguin/bytepool.v3"
	"io"
)

type Context struct {
	Writer   io.Writer
	Data     map[string]interface{}
	Counters map[string]int
	Contents map[string]*bytepool.Bytes
}
