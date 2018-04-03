package api

import (
	"github.com/xeipuuv/gojsonschema"
)

type Request interface {
	IsRequest()
	IsValid() (*gojsonschema.Result, error)

	Reset()
	String() string
	ProtoMessage()
}

type Response interface {
	Reset()
	String() string
	ProtoMessage()
}
