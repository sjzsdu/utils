package markdown

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// NodeInfo 定义了节点的基本信息接口
type NodeInfo interface {
	// GetName 获取节点名称
	GetName() string
	// GetPath 获取节点路径
	GetPath() string
	// IsDir 判断是否为目录
	IsDir() bool
	// GetFileInfo 获取文件信息
	GetFileInfo() os.FileInfo
	// ReadContent 读取节点内容
	ReadContent() ([]byte, error)
}

// ProjectTree 定义了项目树的接口
type ProjectTree interface {
	// FindNode 根据路径查找节点
	FindNode(path string) (NodeInfo, error)
	// Visit 遍历项目树中的所有节点
	Visit(visitor func(path string, node NodeInfo, depth int) error) error
}

//go:embed templates/*.html
var templateFS embed.FS

// MarkdownFile 表示一个markdown文件的信息
type MarkdownFile struct {
	Path         string
	Name         string
	Size         int64
	RelativePath string
	Title        string // 从 MD 文件中提取的主标题
	Description  string // 从 MD 文件中提取的描述（第一段文字）
}

// ServerOptions 定义Markdown服务器选项
type ServerOptions struct {
	// TemplatesDir 模板文件目录
	TemplatesDir string
	// ShowContentOnly 是否仅显示内容
	ShowContentOnly bool
	// CustomTemplates 自定义模板
	CustomTemplates *template.Template
}

// DefaultServerOptions 返回默认的服务器选项
func DefaultServerOptions() ServerOptions {
	return ServerOptions{
		TemplatesDir:    "templates",
		ShowContentOnly: false,
	}
}

// MarkdownServer 处理HTTP请求，调用Manager和Renderer
type MarkdownServer struct {
	manager         Manager
	renderer        Renderer
	templates       *template.Template
	markdownContent string
	showContentOnly bool
	projectTree     ProjectTree // 项目树接口
}

// 常用图片类型的MIME映射
var mimeTypes = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"svg":  "image/svg+xml",
	"bmp":  "image/bmp",
	"webp": "image/webp",
	"ico":  "image/x-icon",
	"tif":  "image/tiff",
	"tiff": "image/tiff",
}

// NewMarkdownServer 创建新的Markdown服务器
func NewMarkdownServer(manager Manager, renderer Renderer, options ...ServerOptions) (*MarkdownServer, error) {
	// 使用默认选项
	opt := DefaultServerOptions()
	if len(options) > 0 {
		opt = options[0]
	}

	var templates *template.Template

	// 如果提供了自定义模板，直接使用
	if opt.CustomTemplates != nil {
		templates = opt.CustomTemplates
	} else {
		// 加载模板
		tmplContentList, err := templateFS.ReadFile(fmt.Sprintf("%s/list.html", opt.TemplatesDir))
		if err != nil {
			return nil, fmt.Errorf("读取list.html模板失败: %v", err)
		}

		tmplContentView, err := templateFS.ReadFile(fmt.Sprintf("%s/view.html", opt.TemplatesDir))
		if err != nil {
			return nil, fmt.Errorf("读取view.html模板失败: %v", err)
		}

		// 创建模板函数映射
		funcMap := template.FuncMap{
			"div": func(a, b interface{}) float64 {
				var af, bf float64
				switch v := a.(type) {
				case int64:
					af = float64(v)
				case float64:
					af = v
				case int:
					af = float64(v)
				}
				switch v := b.(type) {
				case int64:
					bf = float64(v)
				case float64:
					bf = v
				case int:
					bf = float64(v)
				}
				if bf != 0 {
					return af / bf
				}
				return 0
			},
			"multiply": func(a, b interface{}) int {
				var ai, bi int
				switch v := a.(type) {
				case int:
					ai = v
				case int64:
					ai = int(v)
				case float64:
					ai = int(v)
				}
				switch v := b.(type) {
				case int:
					bi = v
				case int64:
					bi = int(v)
				case float64:
					bi = int(v)
				}
				return ai * bi
			},
		}

		// 解析模板
		// 先解析list模板
		listTmpl, err := template.New("list").Funcs(funcMap).Parse(string(tmplContentList))
		if err != nil {
			return nil, fmt.Errorf("解析list模板失败: %v", err)
		}

		// 再解析view模板
		viewTmpl, err := listTmpl.New("view").Parse(string(tmplContentView))
		if err != nil {
			return nil, fmt.Errorf("解析view模板失败: %v", err)
		}

		templates = viewTmpl
	}

	return &MarkdownServer{
		manager:         manager,
		renderer:        renderer,
		templates:       templates,
		showContentOnly: opt.ShowContentOnly,
	}, nil
}

// SetMarkdownContent 设置直接提供的Markdown内容
func (s *MarkdownServer) SetMarkdownContent(content string, showOnly bool) {
	s.markdownContent = content
	s.showContentOnly = showOnly
}

// SetProjectTree 设置项目树接口
func (s *MarkdownServer) SetProjectTree(projectTree ProjectTree) {
	s.projectTree = projectTree
}

// HandleMarkdownList 处理markdown文件列表页面
func (s *MarkdownServer) HandleMarkdownList(w http.ResponseWriter, r *http.Request, proj ProjectTree) error {
	markdownFiles, err := s.getMarkdownFiles(proj)
	if err != nil {
		return fmt.Errorf("获取markdown文件失败: %v", err)
	}

	data := struct {
		Files []MarkdownFile
		Total int
	}{}

	data.Files = markdownFiles
	data.Total = len(markdownFiles)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "list", data); err != nil {
		return fmt.Errorf("模板渲染失败: %v", err)
	}

	return nil
}

// HandleMarkdownView 处理markdown文件查看页面
func (s *MarkdownServer) HandleMarkdownView(w http.ResponseWriter, r *http.Request, proj ProjectTree) error {
	// 从URL中提取文件路径
	filePath := strings.TrimPrefix(r.URL.Path, "/view")
	if filePath == "" || filePath == "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	}

	// 检查是否是通过--content参数提供的文档
	if s.markdownContent != "" {
		// 获取默认文件名
		defaultFileName := "/document.md"
		// 尝试从内容中提取标题作为文件名
		title, _ := s.renderer.ExtractTitleAndDescription(s.markdownContent)
		if title != "" {
			// 将标题转换为有效的文件名
			fileName := strings.ToLower(title)
			fileName = strings.ReplaceAll(fileName, " ", "-")
			// 移除特殊字符
			fileName = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(fileName, "")
			if fileName != "" {
				defaultFileName = "/" + fileName + ".md"
			}
		}

		// 如果请求的是这个特殊文档
		if filePath == defaultFileName {
			// 处理Markdown内容，修复Mermaid图表中的语法问题
			processedContent := s.renderer.ProcessContent(s.markdownContent, "./")

			// 获取最新项目树
			var markdownFiles []MarkdownFile
			if !s.showContentOnly {
				markdownFiles, _ = s.getMarkdownFiles(proj)
			}

			data := struct {
				FilePath      string
				Content       template.HTML
				RawPath       string
				MarkdownFiles []MarkdownFile
			}{}

			data.FilePath = filePath
			data.Content = processedContent
			data.RawPath = "/raw" + filePath
			data.MarkdownFiles = markdownFiles

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := s.templates.ExecuteTemplate(w, "view", data); err != nil {
				return fmt.Errorf("模板渲染失败: %v", err)
			}

			return nil
		}
	}

	// 查找文件节点
	node, err := proj.FindNode(filePath)
	if err != nil {
		return fmt.Errorf("文件不存在: %v", err)
	}

	// 读取文件内容（确保获取最新内容）
	content, err := node.ReadContent()
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 处理Markdown内容，修复Mermaid图表中的语法问题
	processedContent := s.renderer.ProcessContent(string(content), filepath.Dir(filePath))

	// 获取所有markdown文件列表
	markdownFiles, err := s.getMarkdownFiles(proj)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %v", err)
	}

	data := struct {
		FilePath      string
		Content       template.HTML
		RawPath       string
		MarkdownFiles []MarkdownFile
	}{}

	data.FilePath = filePath
	data.Content = processedContent
	data.RawPath = "/raw" + filePath
	data.MarkdownFiles = markdownFiles

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "view", data); err != nil {
		return fmt.Errorf("模板渲染失败: %v", err)
	}

	return nil
}

// HandleMarkdownRaw 处理原始markdown内容
func (s *MarkdownServer) HandleMarkdownRaw(w http.ResponseWriter, r *http.Request, proj ProjectTree) error {
	// 从URL中提取文件路径
	filePath := strings.TrimPrefix(r.URL.Path, "/raw")
	if filePath == "" || filePath == "/" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 检查是否是通过--content参数提供的文档
	if s.markdownContent != "" {
		// 获取默认文件名
		defaultFileName := "/document.md"
		// 尝试从内容中提取标题作为文件名
		title, _ := s.renderer.ExtractTitleAndDescription(s.markdownContent)
		if title != "" {
			// 将标题转换为有效的文件名
			fileName := strings.ToLower(title)
			fileName = strings.ReplaceAll(fileName, " ", "-")
			// 移除特殊字符
			fileName = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(fileName, "")
			if fileName != "" {
				defaultFileName = "/" + fileName + ".md"
			}
		}

		// 如果请求的是这个特殊文档
		if filePath == defaultFileName {
			// 从文件路径中提取文件名
			fileName := filepath.Base(filePath)

			// 设置响应头，支持文件下载
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
			w.Write([]byte(s.markdownContent))
			return nil
		}
	}

	// 查找文件节点
	node, err := proj.FindNode(filePath)
	if err != nil {
		return fmt.Errorf("文件不存在: %v", err)
	}

	// 读取文件内容（确保获取最新内容）
	content, err := node.ReadContent()
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 从文件路径中提取文件名
	fileName := filepath.Base(filePath)

	// 确保文件名有.md扩展名
	if !strings.HasSuffix(fileName, ".md") && !strings.HasSuffix(fileName, ".markdown") {
		fileName += ".md"
	}

	// 设置响应头，支持文件下载
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	w.Write(content)

	return nil
}

// HandleMarkdownContent 处理直接提供的markdown内容
func (s *MarkdownServer) HandleMarkdownContent(w http.ResponseWriter, r *http.Request, proj ProjectTree) error {
	// 处理Markdown内容，修复Mermaid图表中的语法问题
	processedContent := s.renderer.ProcessContent(s.markdownContent, "./")

	// 准备数据
	var markdownFiles []MarkdownFile
	if !s.showContentOnly {
		// 如果不是仅显示内容，获取所有markdown文件列表
		markdownFiles, _ = s.getMarkdownFiles(proj)
	}

	data := struct {
		FilePath      string
		Content       template.HTML
		RawPath       string
		MarkdownFiles []MarkdownFile
	}{}

	data.FilePath = "直接提供的内容"
	data.Content = processedContent
	data.RawPath = "/raw-content" // 设置一个固定路径用于下载
	data.MarkdownFiles = markdownFiles

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "view", data); err != nil {
		return fmt.Errorf("模板渲染失败: %v", err)
	}

	return nil
}

// HandleRawContentDownload 处理直接提供的markdown内容的下载
func (s *MarkdownServer) HandleRawContentDownload(w http.ResponseWriter, r *http.Request) error {
	// 设置响应头，支持文件下载
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="document.md"`)
	w.Write([]byte(s.markdownContent))
	return nil
}

// HandleImages 处理Markdown文档中的本地图片请求
func (s *MarkdownServer) HandleImages(w http.ResponseWriter, r *http.Request, proj ProjectTree) error {
	// 从URL中提取图片文件路径
	// URL格式: /images/[图片路径]
	imagePath := strings.TrimPrefix(r.URL.Path, "/images")
	if imagePath == "" || imagePath == "/" {
		http.Error(w, "图片路径不能为空", http.StatusBadRequest)
		return nil
	}

	// 查找图片文件节点
	node, err := proj.FindNode(imagePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("图片不存在: %v", err), http.StatusNotFound)
		return nil
	}

	// 检查是否为文件
	if node.IsDir() {
		http.Error(w, "路径不是图片文件", http.StatusBadRequest)
		return nil
	}

	// 读取图片内容
	content, err := node.ReadContent()
	if err != nil {
		http.Error(w, fmt.Sprintf("读取图片失败: %v", err), http.StatusInternalServerError)
		return nil
	}

	// 设置正确的Content-Type
	contentType := "application/octet-stream"
	if ext := strings.ToLower(filepath.Ext(node.GetName())); ext != "" {
		if mime, ok := mimeTypes[ext[1:]]; ok {
			contentType = mime
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

	// 返回图片内容
	w.Write(content)
	return nil
}

// Handler 返回配置好的HTTP处理器
func (s *MarkdownServer) Handler() http.Handler {
	mux := http.NewServeMux()

	// 无论是否提供了markdown内容，都注册所有路由
	if s.markdownContent != "" {
		// 如果提供了内容，首页直接显示该内容
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				if s.projectTree == nil {
					http.Error(w, "项目树未初始化", http.StatusInternalServerError)
					return
				}
				if err := s.HandleMarkdownContent(w, r, s.projectTree); err != nil {
					http.Error(w, fmt.Sprintf("处理Markdown内容失败: %v", err), http.StatusInternalServerError)
					return
				}
			} else if r.URL.Path == "/raw-content" {
				// 处理直接提供内容的下载
				if err := s.HandleRawContentDownload(w, r); err != nil {
					http.Error(w, fmt.Sprintf("下载失败: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				http.NotFound(w, r)
			}
		})
	} else {
		// 否则显示文件列表
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				if s.projectTree == nil {
					http.Error(w, "项目树未初始化", http.StatusInternalServerError)
					return
				}
				if err := s.HandleMarkdownList(w, r, s.projectTree); err != nil {
					http.Error(w, fmt.Sprintf("获取文件列表失败: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				http.NotFound(w, r)
			}
		})
	}

	// 专门的文件列表路由，无论是否提供了内容都显示列表
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		if s.projectTree == nil {
			http.Error(w, "项目树未初始化", http.StatusInternalServerError)
			return
		}
		if err := s.HandleMarkdownList(w, r, s.projectTree); err != nil {
			http.Error(w, fmt.Sprintf("获取文件列表失败: %v", err), http.StatusInternalServerError)
			return
		}
	})

	// 查看markdown文件
	mux.HandleFunc("/view/", func(w http.ResponseWriter, r *http.Request) {
		if s.projectTree == nil {
			http.Error(w, "项目树未初始化", http.StatusInternalServerError)
			return
		}
		if err := s.HandleMarkdownView(w, r, s.projectTree); err != nil {
			http.Error(w, fmt.Sprintf("查看文件失败: %v", err), http.StatusInternalServerError)
			return
		}
	})

	// 原始markdown内容
	mux.HandleFunc("/raw/", func(w http.ResponseWriter, r *http.Request) {
		if s.projectTree == nil {
			http.Error(w, "项目树未初始化", http.StatusInternalServerError)
			return
		}
		if err := s.HandleMarkdownRaw(w, r, s.projectTree); err != nil {
			http.Error(w, fmt.Sprintf("获取原始内容失败: %v", err), http.StatusInternalServerError)
			return
		}
	})

	// 本地图片访问
	mux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		if s.projectTree == nil {
			http.Error(w, "项目树未初始化", http.StatusInternalServerError)
			return
		}
		if err := s.HandleImages(w, r, s.projectTree); err != nil {
			http.Error(w, fmt.Sprintf("处理图片失败: %v", err), http.StatusInternalServerError)
			return
		}
	})

	return mux
}

// StartServer 启动Markdown文档服务（已过时，建议使用Handler()方法）
func (s *MarkdownServer) StartServer(port int) error {
	// 使用新的Handler()方法
	mux := s.Handler()

	// 尝试绑定端口，直到成功或达到最大尝试次数
	maxAttempts := 20
	var server *http.Server

	for attempt := 0; attempt < maxAttempts; attempt++ {
		currentPort := port + attempt

		// 创建临时监听器来检查端口是否可用
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", currentPort))
		if err != nil {
			// 端口被占用，尝试下一个
			fmt.Printf("端口 %d 已被占用，尝试下一个端口...\n", currentPort)
			continue
		}

		// 端口可用，关闭监听器并使用该端口启动服务器
		listener.Close()

		fmt.Printf("正在启动Markdown文档服务，端口: %d\n", currentPort)

		server = &http.Server{
			Addr:    fmt.Sprintf(":%d", currentPort),
			Handler: mux,
		}

		// 启动服务器（使用goroutine避免阻塞）
		go func(p int) {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("服务器运行失败: %v\n", err)
			}
		}(currentPort)

		fmt.Printf("Markdown文档服务已启动: http://localhost:%d\n", currentPort)
		fmt.Println("按 Ctrl+C 停止服务...")
		break
	}

	if server == nil {
		return fmt.Errorf("尝试 %d 次后仍未找到可用端口", maxAttempts)
	}

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("\n正在关闭Markdown文档服务...")

	return nil
}

// getMarkdownFiles 获取项目中所有的markdown文件
func (s *MarkdownServer) getMarkdownFiles(proj ProjectTree) ([]MarkdownFile, error) {
	var markdownFiles []MarkdownFile

	err := proj.Visit(func(path string, node NodeInfo, depth int) error {
		if !node.IsDir() && strings.HasSuffix(strings.ToLower(node.GetName()), ".md") {
			file := MarkdownFile{
				Path:         node.GetPath(),
				Name:         node.GetName(),
				RelativePath: path,
				Size:         0,
			}

			// 尝试获取文件大小
			if info := node.GetFileInfo(); info != nil {
				file.Size = info.Size()
			}

			// 读取内容提取标题和描述
			if content, err := node.ReadContent(); err == nil {
				if file.Size == 0 {
					file.Size = int64(len(content))
				}

				// 提取标题和描述
				title, desc := s.renderer.ExtractTitleAndDescription(string(content))
				file.Title = title
				file.Description = desc
			}

			markdownFiles = append(markdownFiles, file)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 如果提供了markdown内容，将其添加到文件列表
	if s.markdownContent != "" {
		// 使用默认文件名
		defaultFileName := "document.md"
		// 尝试从内容中提取标题作为文件名
		title, _ := s.renderer.ExtractTitleAndDescription(s.markdownContent)
		if title != "" {
			// 将标题转换为有效的文件名
			fileName := strings.ToLower(title)
			fileName = strings.ReplaceAll(fileName, " ", "-")
			// 移除特殊字符
			fileName = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(fileName, "")
			if fileName != "" {
				defaultFileName = fileName + ".md"
			}
		}

		// 添加到文件列表，确保RelativePath以斜杠开头
		file := MarkdownFile{
			Path:         "/" + defaultFileName,
			Name:         defaultFileName,
			RelativePath: "/" + defaultFileName,
			Size:         int64(len(s.markdownContent)),
			Title:        title,
		}

		markdownFiles = append(markdownFiles, file)
	}

	// 按路径排序
	sort.Slice(markdownFiles, func(i, j int) bool {
		return markdownFiles[i].RelativePath < markdownFiles[j].RelativePath
	})

	return markdownFiles, nil
}
