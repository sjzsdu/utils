package sources

import (
	"io"
	"net/http"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// readResponseBody 读取HTTP响应体
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, ErrNonOkStatusCode
	}
	return io.ReadAll(resp.Body)
}

// NewBaseItem 创建基础的Item实例
func NewBaseItem(id, title, url, source string) models.Item {
	now := time.Now()
	return models.Item{
		ID:          id,
		Title:       title,
		URL:         url,
		Source:      source,
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: now,
	}
}

// 定义一些常见的错误
var (
	ErrNonOkStatusCode = NewError("non-ok status code")
	ErrEmptyContent    = NewError("empty content")
)

// Error 自定义错误类型
type Error struct {
	Message string
}

// NewError 创建新的错误
func NewError(message string) *Error {
	return &Error{Message: message}
}

// Error 实现error接口
func (e *Error) Error() string {
	return e.Message
}
