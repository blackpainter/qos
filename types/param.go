package types

import (
	btypes "github.com/QOSGroup/qbase/types"
)

type KeyValuePair struct {
	Key   []byte
	Value interface{}
}

type KeyValuePairs []KeyValuePair

type ParamSet interface {
	KeyValuePairs() KeyValuePairs
	ValidateKeyValue(key string, value string) (interface{}, btypes.Error)
	GetParamSpace() string
	Validate() btypes.Error
}
