package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

//goland:noinspection GoMixedReceiverTypes
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%v mins", r)
	jsonValue = strconv.Quote(jsonValue)
	return []byte(jsonValue), nil
}

//goland:noinspection GoMixedReceiverTypes
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unQuotedJsonValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	parts := strings.Split(unQuotedJsonValue, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(i)
	return nil
}
