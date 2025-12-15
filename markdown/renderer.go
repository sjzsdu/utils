package markdown

import (
	"html/template"
	"path/filepath"
	"strings"
)

// Renderer 定义Markdown渲染的接口
type Renderer interface {
	// ExtractTitleAndDescription 从Markdown内容中提取标题和描述
	ExtractTitleAndDescription(content string) (title string, description string)
	// SanitizeForMermaid 处理Markdown内容，修复Mermaid图表中的语法问题
	SanitizeForMermaid(content string) string
	// ConvertLocalImagesToServerPath 将本地图片引用转换为服务器路径
	ConvertLocalImagesToServerPath(content, currentDir string) string
	// ConvertServerImagesToLocalPath 将服务器路径图片引用转换为本地路径
	ConvertServerImagesToLocalPath(content, currentDir string) string
	// ProcessContent 处理Markdown内容，包括Mermaid图表和图片路径转换
	ProcessContent(content, currentDir string) template.HTML
	// ProcessContentWithOptions 处理Markdown内容，支持自定义选项
	ProcessContentWithOptions(content, currentDir string, options ProcessOptions) template.HTML
}

// ProcessOptions 定义Markdown处理选项
type ProcessOptions struct {
	// SanitizeMermaid 是否处理Mermaid图表
	SanitizeMermaid bool
	// ConvertImages 是否转换图片路径
	ConvertImages bool
	// ImagePathConverter 自定义图片路径转换器
	ImagePathConverter func(content, currentDir string) string
}

// DefaultProcessOptions 返回默认的处理选项
func DefaultProcessOptions() ProcessOptions {
	return ProcessOptions{
		SanitizeMermaid: true,
		ConvertImages:   true,
	}
}

// MarkdownRenderer 实现Renderer接口，处理Markdown渲染逻辑
type MarkdownRenderer struct{}

// NewMarkdownRenderer 创建新的Markdown渲染器
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

// ExtractTitleAndDescription 从Markdown内容中提取标题和描述
func (r *MarkdownRenderer) ExtractTitleAndDescription(content string) (title string, description string) {
	lines := strings.Split(content, "\n")
	foundTitle := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 跳过空行
		if trimmed == "" {
			continue
		}

		// 提取第一个 # 标题作为标题
		if !foundTitle && strings.HasPrefix(trimmed, "#") {
			// 去掉 # 符号和空格
			title = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			foundTitle = true
			continue
		}

		// 提取第一段非空文本作为描述（跳过代码块、引用等）
		if foundTitle && description == "" {
			// 跳过代码块标记
			if strings.HasPrefix(trimmed, "```") {
				continue
			}
			// 跳过引用块
			if strings.HasPrefix(trimmed, ">") {
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			}
			// 跳过列表项
			if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "+") {
				continue
			}
			// 跳过标题
			if strings.HasPrefix(trimmed, "#") {
				continue
			}

			// 如果是普通文本，作为描述
			if len(trimmed) > 0 {
				description = trimmed
				// 限制描述长度
				if len(description) > 120 {
					description = description[:120] + "..."
				}
				break
			}
		}
	}

	// 如果没有找到标题，使用默认值
	if title == "" {
		title = "未命名文档"
	}

	// 如果没有找到描述
	if description == "" {
		description = "暂无描述"
	}

	return title, description
}

// SanitizeForMermaid 处理Markdown内容，修复Mermaid图表中的语法问题
func (r *MarkdownRenderer) SanitizeForMermaid(content string) string {
	// 只处理Mermaid代码块中的内容
	lines := strings.Split(content, "\n")
	var result strings.Builder
	inMermaidBlock := false
	isClassDiagram := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检测Mermaid代码块的开始
		if strings.HasPrefix(trimmed, "```mermaid") {
			inMermaidBlock = true
			isClassDiagram = false
			result.WriteString(line + "\n")
			continue
		}

		// 检测代码块的结束
		if inMermaidBlock && strings.HasPrefix(trimmed, "```") {
			inMermaidBlock = false
			isClassDiagram = false
			result.WriteString(line + "\n")
			continue
		}

		// 在Mermaid代码块内，进行处理
		if inMermaidBlock {
			// 检测是否是类图
			if strings.HasPrefix(trimmed, "classDiagram") {
				isClassDiagram = true
			}

			// 1. 替换interface{}为any（Go 1.18+）
			line = strings.ReplaceAll(line, "interface{}", "any")
			line = strings.ReplaceAll(line, "map[string]interface{}", "map[string]any")
			line = strings.ReplaceAll(line, "[]interface{}", "[]any")
			line = strings.ReplaceAll(line, "...interface{}", "...any")
			line = strings.ReplaceAll(line, "chan interface{}", "chan any")

			// 2. 只在类图中处理类名中的特殊字符
			if isClassDiagram {
				// 检查是否包含类图的关系符号
				if strings.Contains(line, "-->") || strings.Contains(line, "<--") ||
					strings.Contains(line, "..") || strings.Contains(line, "*--") || strings.Contains(line, "o--") || strings.Contains(line, "--o") || strings.Contains(line, "--*") ||
					strings.Contains(line, "<|--") || strings.Contains(line, "--|>") || strings.Contains(line, "<|..") || strings.Contains(line, "..|>") {
					// 处理指针类型的类名（*ClassName）
					line = r.sanitizeMermaidClassName(line)
				}
			}
		}

		result.WriteString(line + "\n")
	}

	return result.String()
}

// sanitizeMermaidClassName 清理Mermaid类图中的类名特殊字符
func (r *MarkdownRenderer) sanitizeMermaidClassName(line string) string {
	// 匹配关系箭头后的类名
	patterns := []string{
		"-->", "<--", "..", "*--", "o--", "--o", "--*",
		"<|--", "--|>", "<|..", "..|>",
	}

	for _, pattern := range patterns {
		if !strings.Contains(line, pattern) {
			continue
		}

		parts := strings.Split(line, pattern)
		if len(parts) != 2 {
			continue
		}

		// 处理右侧部分（可能包含类名）
		rightPart := strings.TrimSpace(parts[1])

		// 如果以*开头，去掉*
		if strings.HasPrefix(rightPart, "*") {
			// 找到类名（可能后面还有:标签）
			tokens := strings.Fields(rightPart)
			if len(tokens) > 0 && strings.HasPrefix(tokens[0], "*") {
				// 去掉前导*
				tokens[0] = strings.TrimPrefix(tokens[0], "*")
				rightPart = strings.Join(tokens, " ")
			}
		}

		line = parts[0] + pattern + " " + rightPart
		break // 只处理第一个匹配
	}

	return line
}

// ConvertLocalImagesToServerPath 将本地图片引用转换为服务器路径
func (r *MarkdownRenderer) ConvertLocalImagesToServerPath(content, currentDir string) string {
	// 使用简单的字符串处理，避免复杂正则表达式
	var result strings.Builder

	// 遍历内容，寻找Markdown图片语法
	for i := 0; i < len(content); i++ {
		// 检查是否是图片开始标记：![
		if i+1 < len(content) && content[i] == '!' && content[i+1] == '[' {
			// 记录当前位置
			start := i
			i += 2 // 跳过![

			// 寻找alt text结束标记：]
			altEnd := strings.Index(content[i:], "]")
			if altEnd == -1 {
				// 不是完整的图片语法，继续
				result.WriteString(content[start:i])
				continue
			}

			altText := content[i : i+altEnd]
			i += altEnd + 1 // 跳过]和(

			// 检查是否是(，如果不是则不是完整的图片语法
			if i >= len(content) || content[i] != '(' {
				result.WriteString(content[start:i])
				continue
			}
			i++ // 跳过(

			// 寻找图片路径结束标记：)
			pathEnd := strings.Index(content[i:], ")")
			if pathEnd == -1 {
				// 不是完整的图片语法，继续
				result.WriteString(content[start:i])
				continue
			}

			imagePath := content[i : i+pathEnd]
			i += pathEnd + 1 // 跳过)

			// 检查是否是HTTP/HTTPS开头的图片，若是则不处理
			if strings.HasPrefix(strings.ToLower(imagePath), "http://") || strings.HasPrefix(strings.ToLower(imagePath), "https://") {
				// 外部图片，保持原样
				result.WriteString("![" + altText + "](" + imagePath + ")")
			} else {
				// 清理图片路径，移除可能的查询参数或锚点
				imagePath = strings.Split(imagePath, "?")[0]
				imagePath = strings.Split(imagePath, "#")[0]

				// 解析图片路径相对当前文件目录
				resolvedPath := imagePath
				if !strings.HasPrefix(imagePath, "/") {
					// 相对路径：将图片路径与当前文件目录结合
					resolvedPath = filepath.Join(currentDir, imagePath)
				}
				// 清理路径，处理..和. segments
				resolvedPath = filepath.Clean(resolvedPath)

				// 确保路径以/开头
				if !strings.HasPrefix(resolvedPath, "/") {
					resolvedPath = "/" + resolvedPath
				}

				// 转换为/images/路径
				result.WriteString("![" + altText + "](/images" + resolvedPath + ")")
			}
		} else {
			// 不是图片语法，直接写入
			result.WriteByte(content[i])
		}
	}

	return result.String()
}

// ConvertServerImagesToLocalPath 将服务器路径图片引用转换为本地路径
func (r *MarkdownRenderer) ConvertServerImagesToLocalPath(content, currentDir string) string {
	// 使用简单的字符串处理，避免复杂正则表达式
	var result strings.Builder

	// 遍历内容，寻找服务器图片语法
	for i := 0; i < len(content); i++ {
		// 检查是否是图片开始标记：![
		if i+1 < len(content) && content[i] == '!' && content[i+1] == '[' {
			// 记录当前位置
			start := i
			i += 2 // 跳过![

			// 寻找alt text结束标记：]
			altEnd := strings.Index(content[i:], "]")
			if altEnd == -1 {
				// 不是完整的图片语法，继续
				result.WriteString(content[start:i])
				continue
			}

			altText := content[i : i+altEnd]
			i += altEnd + 1 // 跳过]和(

			// 检查是否是(，如果不是则不是完整的图片语法
			if i >= len(content) || content[i] != '(' {
				result.WriteString(content[start:i])
				continue
			}
			i++ // 跳过(

			// 寻找图片路径结束标记：)
			pathEnd := strings.Index(content[i:], ")")
			if pathEnd == -1 {
				// 不是完整的图片语法，继续
				result.WriteString(content[start:i])
				continue
			}

			imagePath := content[i : i+pathEnd]
			i += pathEnd + 1 // 跳过)

			// 检查是否是/images/开头的服务器图片路径
			if strings.HasPrefix(imagePath, "/images") {
				// 移除/images前缀
				localPath := strings.TrimPrefix(imagePath, "/images")
				// 如果当前目录不是根目录，将路径转换为相对路径
				if currentDir != "" && currentDir != "/" {
					localPath = filepath.Join(currentDir, localPath)
					// 清理路径
					localPath = filepath.Clean(localPath)
					// 如果路径以当前目录开头，转换为相对路径
					if strings.HasPrefix(localPath, currentDir) {
						localPath = strings.TrimPrefix(localPath, currentDir)
						localPath = strings.TrimPrefix(localPath, "/")
						if localPath == "" {
							localPath = "."
						}
					}
				}
				// 输出本地图片路径
				result.WriteString("![" + altText + "](" + localPath + ")")
			} else {
				// 不是服务器图片路径，保持原样
				result.WriteString("![" + altText + "]" + imagePath + ")")
			}
		} else {
			// 不是图片语法，直接写入
			result.WriteByte(content[i])
		}
	}

	return result.String()
}

// ProcessContent 处理Markdown内容，包括Mermaid图表和图片路径转换
func (r *MarkdownRenderer) ProcessContent(content, currentDir string) template.HTML {
	// 使用默认选项处理内容
	return r.ProcessContentWithOptions(content, currentDir, DefaultProcessOptions())
}

// ProcessContentWithOptions 处理Markdown内容，支持自定义选项
func (r *MarkdownRenderer) ProcessContentWithOptions(content, currentDir string, options ProcessOptions) template.HTML {
	processedContent := content

	// 根据选项处理内容
	if options.SanitizeMermaid {
		// 修复Mermaid图表中的语法问题
		processedContent = r.SanitizeForMermaid(processedContent)
	}

	if options.ConvertImages {
		if options.ImagePathConverter != nil {
			// 使用自定义图片路径转换器
			processedContent = options.ImagePathConverter(processedContent, currentDir)
		} else {
			// 使用默认的图片路径转换
			processedContent = r.ConvertLocalImagesToServerPath(processedContent, currentDir)
		}
	}

	return template.HTML(processedContent)
}
