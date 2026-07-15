package common

import "errors"

var (
	ErrInvalidParam       = errors.New("参数无效")
	ErrInvalidRequestBody = errors.New("请求内容无效")
	ErrNotFound           = errors.New("记录不存在")
	ErrAlreadyExists      = errors.New("记录已经存在")
	ErrRunAlreadyFinished = errors.New("任务已经结束")
)
