package markdown

import "errors"

// 定义包内使用的错误类型
var (
	ErrContentNotFound = errors.New("markdown content not found")
	ErrInvalidPath     = errors.New("invalid markdown path")
	ErrRenderFailed    = errors.New("markdown render failed")
)
