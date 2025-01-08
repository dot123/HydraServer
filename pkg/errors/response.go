package errors

import (
	"errors"
	"fmt"
	"github.com/topfreegames/pitaya/v2"
	pitayaErrors "github.com/topfreegames/pitaya/v2/errors"
)

// NewResponseError 自定义响应错误
func NewResponseError(code int, err error) *pitayaErrors.Error {
	if err == nil {
		return pitaya.Error(errors.New(""), fmt.Sprintf("%d", code))
	}
	return pitaya.Error(err, fmt.Sprintf("%d", code))
}
