package extractor

import (
	"time"
)

// Extractor 定义了数据提取的基本行为
type Extractor interface {
	// ExtractTitle 提取标题
	ExtractTitle(content []byte) (string, error)
	
	// ExtractContent 提取内容
	ExtractContent(content []byte) (string, error)
	
	// ExtractLinks 提取链接
	ExtractLinks(content []byte) ([]string, error)
	
	// ExtractImages 提取图片
	ExtractImages(content []byte) ([]string, error)
	
	// ExtractTime 提取时间
	ExtractTime(content []byte) (time.Time, error)
}