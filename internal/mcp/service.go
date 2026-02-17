package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/feewg/kaf-cli/internal/converter"
	"github.com/feewg/kaf-cli/internal/core"
	"github.com/feewg/kaf-cli/internal/model"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ConverterService struct {
	version string
}

func (s *ConverterService) RegisterTools(srv *server.MCPServer, version string) {
	s.version = version

	// 注册转换工具
	s.registerConvertTool(srv)

	// 注册文件夹批量转换工具
	s.registerBatchConvertTool(srv)

	// 添加提示词模板
	s.registerPrompts(srv)
}

// registerConvertTool 注册转换工具，支持所有 CLI 参数
func (s *ConverterService) registerConvertTool(srv *server.MCPServer) {
	tool := mcpgo.NewTool("kaf_convert",
		mcpgo.WithDescription("电子书格式转换器，支持把txt文件转换成epub/mobi/azw3电子书格式\n转换成功后返回生成的电子书文件路径"),
		// 必填参数
		mcpgo.WithString("filename",
			mcpgo.Required(),
			mcpgo.Description("txt小说文件路径，支持相对路径和绝对路径"),
		),
		// 可选参数 - 基本信息
		mcpgo.WithString("bookname",
			mcpgo.Description("书名，为空时自动从文件名识别"),
		),
		mcpgo.WithString("author",
			mcpgo.Description("作者，为空时自动从文件名识别，默认YSTYLE"),
		),
		mcpgo.WithString("format",
			mcpgo.Description("输出格式: all(全部), epub, mobi, azw3，默认all"),
		),
		mcpgo.WithString("out",
			mcpgo.Description("输出文件名(不含后缀)，默认为书名"),
		),
		// 可选参数 - 章节匹配
		mcpgo.WithString("match",
			mcpgo.Description("章节匹配正则表达式，不填自动识别"),
		),
		mcpgo.WithString("volume_match",
			mcpgo.Description("卷匹配规则，默认启用，设为false禁用"),
		),
		mcpgo.WithString("exclude",
			mcpgo.Description("排除无效章节的正则表达式"),
		),
		mcpgo.WithString("unknow_title",
			mcpgo.Description("未知章节默认名称，默认'章节正文'"),
		),
		// 可选参数 - 封面设置
		mcpgo.WithString("cover",
			mcpgo.Description("封面图片: 本地路径、orly(在线生成)、none(无封面)"),
		),
		mcpgo.WithString("cover_orly_color",
			mcpgo.Description("orly封面主题色: 1-16或hex颜色代码"),
		),
		mcpgo.WithNumber("cover_orly_idx",
			mcpgo.Description("orly封面动物图案: 0-41"),
		),
		// 可选参数 - 字体和样式
		mcpgo.WithString("font",
			mcpgo.Description("嵌入字体文件路径"),
		),
		mcpgo.WithNumber("max",
			mcpgo.Description("标题最大字数，默认35"),
		),
		mcpgo.WithNumber("indent",
			mcpgo.Description("段落缩进字数，默认2"),
		),
		mcpgo.WithString("align",
			mcpgo.Description("标题对齐: left, center, right，默认center"),
		),
		mcpgo.WithString("bottom",
			mcpgo.Description("段落间距，默认1em"),
		),
		mcpgo.WithString("line_height",
			mcpgo.Description("行高，默认1.5rem"),
		),
		mcpgo.WithString("lang",
			mcpgo.Description("语言: en,de,fr,it,es,zh,ja,pt,ru,nl，默认zh"),
		),
		mcpgo.WithBoolean("tips",
			mcpgo.Description("添加制作教程，默认true"),
		),
		mcpgo.WithBoolean("separate_chapter_number",
			mcpgo.Description("分离章节序号和标题样式，默认false"),
		),
		mcpgo.WithString("custom_css_file",
			mcpgo.Description("自定义CSS文件路径"),
		),
		// 新增 - 扩展CSS样式
		mcpgo.WithString("extended_css",
			mcpgo.Description("内联扩展CSS样式（直接传入CSS代码）"),
		),
		mcpgo.WithString("css_variables",
			mcpgo.Description("CSS变量定义，格式: --var1:value1;--var2:value2"),
		),
		// 新增 - 章节页眉图片
		mcpgo.WithString("chapter_header_image",
			mcpgo.Description("章节页眉图片路径，所有章节显示相同图片"),
		),
		mcpgo.WithString("chapter_header_image_folder",
			mcpgo.Description("章节页眉图片文件夹，按章节名匹配图片"),
		),
		mcpgo.WithString("chapter_header_image_position",
			mcpgo.Description("页眉图片位置: left, center, right，默认center"),
		),
		mcpgo.WithString("chapter_header_image_height",
			mcpgo.Description("页眉图片高度，如: 100px, 2em，默认auto"),
		),
		mcpgo.WithString("chapter_header_image_width",
			mcpgo.Description("页眉图片宽度，如: 50%, 200px，默认100%"),
		),
		mcpgo.WithString("chapter_header_image_mode",
			mcpgo.Description("图片模式: single(所有章节相同), folder(按章节名匹配)，默认single"),
		),
	)

	srv.AddTool(tool, s.handleConvert)
}

// registerBatchConvertTool 注册文件夹批量转换工具
func (s *ConverterService) registerBatchConvertTool(srv *server.MCPServer) {
	tool := mcpgo.NewTool("kaf_batch_convert",
		mcpgo.WithDescription("批量转换文件夹中的txt小说文件\n按规范命名自动识别书名、作者、封面、页眉图片等\n支持批量生成epub/mobi/azw3格式电子书"),
		// 必填参数
		mcpgo.WithString("folder",
			mcpgo.Required(),
			mcpgo.Description("小说文件夹路径，包含txt文件和可选资源文件"),
		),
		// 可选参数 - 输出控制
		mcpgo.WithString("format",
			mcpgo.Description("输出格式: all(全部), epub, mobi, azw3，默认all"),
		),
		mcpgo.WithString("output_folder",
			mcpgo.Description("输出文件夹，默认使用输入文件夹"),
		),
		// 可选参数 - 全局样式
		mcpgo.WithString("font",
			mcpgo.Description("嵌入字体文件路径（全局）"),
		),
		mcpgo.WithString("align",
			mcpgo.Description("标题对齐: left, center, right，默认center"),
		),
		mcpgo.WithNumber("indent",
			mcpgo.Description("段落缩进字数，默认2"),
		),
		mcpgo.WithString("bottom",
			mcpgo.Description("段落间距，默认1em"),
		),
		mcpgo.WithString("line_height",
			mcpgo.Description("行高，默认1.5rem"),
		),
		mcpgo.WithString("lang",
			mcpgo.Description("语言: en,de,fr,it,es,zh,ja,pt,ru,nl，默认zh"),
		),
		mcpgo.WithBoolean("separate_chapter_number",
			mcpgo.Description("分离章节序号和标题样式，默认false"),
		),
		mcpgo.WithString("custom_css_file",
			mcpgo.Description("全局自定义CSS文件路径"),
		),
		mcpgo.WithString("extended_css",
			mcpgo.Description("全局内联扩展CSS样式"),
		),
		mcpgo.WithString("css_variables",
			mcpgo.Description("全局CSS变量定义"),
		),
	)

	srv.AddTool(tool, s.handleBatchConvert)
}

// handleConvert 处理转换请求
func (s *ConverterService) handleConvert(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	logger.Info("convert request received", "params", req.Params.Arguments)

	// 提取参数
	args := req.Params.Arguments

	filename, ok := args["filename"].(string)
	if !ok || filename == "" {
		return nil, errors.New("filename is required")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}

	// 创建 Book 对象
	book, err := model.NewBookSimple(filename)
	if err != nil {
		return nil, err
	}

	// 应用所有可选参数
	s.applyBookParameters(book, args)

	// 执行转换
	logger.Info("starting conversion", "book", book.Bookname, "format", book.Format)

	// 执行转换流程
	if err := core.Check(book, s.version); err != nil {
		logger.Error("check failed", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := core.Parse(book); err != nil {
		logger.Error("parse failed", "error", err)
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	conv := converter.Dispatcher{Book: book}
	if err := conv.Convert(); err != nil {
		logger.Error("convert failed", "error", err)
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	// 获取输出文件信息
	outputFiles := s.getOutputFiles(book)

	logger.Info("conversion completed", "output", outputFiles)

	// 构建结果
	resultText := fmt.Sprintf(`转换成功！

书籍信息:
- 书名: %s
- 作者: %s
- 输出格式: %s

输出文件:
%s
`,
		book.Bookname,
		book.Author,
		s.getFormatsStr(book.Format),
		strings.Join(outputFiles, "\n"),
	)

	return mcpgo.NewToolResultText(resultText), nil
}

// handleBatchConvert 处理批量转换请求
func (s *ConverterService) handleBatchConvert(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	logger.Info("batch convert request received", "params", req.Params.Arguments)

	args := req.Params.Arguments

	folder, ok := args["folder"].(string)
	if !ok || folder == "" {
		return nil, errors.New("folder is required")
	}

	// 检查文件夹是否存在
	if info, err := os.Stat(folder); os.IsNotExist(err) || !info.IsDir() {
		return nil, fmt.Errorf("folder not found or not a directory: %s", folder)
	}

	// 获取输出文件夹
	outputFolder := folder
	if v, ok := args["output_folder"].(string); ok && v != "" {
		outputFolder = v
		// 确保输出文件夹存在
		if err := os.MkdirAll(outputFolder, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output folder: %w", err)
		}
	}

	// 扫描文件夹获取所有书籍
	books := s.scanBooks(folder, outputFolder)
	if len(books) == 0 {
		return mcpgo.NewToolResultText("未找到符合规范的txt文件。\n\n规范命名格式:\n- 书名.txt\n- 《书名》作者：作者名.txt\n- 《书名》（校对版全本）作者：作者名.txt"), nil
	}

	// 获取全局样式参数
	globalParams := s.extractGlobalParams(args)

	// 批量转换
	var results []string
	var successCount, failCount int

	for _, bookInfo := range books {
		logger.Info("processing book", "book", bookInfo.Book.Bookname)

		// 应用全局样式参数
		s.applyGlobalParams(bookInfo.Book, globalParams)

		// 执行转换
		if err := s.convertBook(bookInfo.Book); err != nil {
			logger.Error("convert failed", "book", bookInfo.Book.Bookname, "error", err)
			results = append(results, fmt.Sprintf("❌ %s: %s", bookInfo.Book.Bookname, err.Error()))
			failCount++
			continue
		}

		// 获取输出文件
		outputFiles := s.getOutputFiles(bookInfo.Book)
		results = append(results, fmt.Sprintf("✅ %s:\n   %s", bookInfo.Book.Bookname, strings.Join(outputFiles, "\n   ")))
		successCount++
	}

	// 构建结果
	resultText := fmt.Sprintf(`批量转换完成！

统计:
- 成功: %d
- 失败: %d
- 总计: %d

详细结果:
%s
`,
		successCount,
		failCount,
		len(books),
		strings.Join(results, "\n\n"),
	)

	return mcpgo.NewToolResultText(resultText), nil
}

// BookInfo 包含书籍信息和相关文件路径
type BookInfo struct {
	Book         *model.Book
	CoverPath    string
	HeaderPath   string
	HeaderFolder string
}

// scanBooks 扫描文件夹获取所有书籍
// 支持两种模式：
// 1. 单文件夹模式：所有txt在一个文件夹，资源使用统一命名（cover.jpg, header.png等）
// 2. 子文件夹模式：每个txt在独立子文件夹，子文件夹内有各自的资源
func (s *ConverterService) scanBooks(folder, outputFolder string) []BookInfo {
	var books []BookInfo

	entries, err := os.ReadDir(folder)
	if err != nil {
		logger.Error("failed to read folder", "error", err)
		return books
	}

	// 首先检查是否是"单文件夹多书籍"模式
	// 查找通用的资源文件（cover.jpg, header.png等）
	globalResources := s.findGlobalResources(folder)

	// 处理所有txt文件和子文件夹
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否是子文件夹模式（子文件夹内有txt文件）
			subBooks := s.scanSubFolder(filepath.Join(folder, entry.Name()), outputFolder, globalResources)
			books = append(books, subBooks...)
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}

		// 创建书籍
		bookPath := filepath.Join(folder, name)
		book, err := model.NewBookSimple(bookPath)
		if err != nil {
			logger.Error("failed to create book", "file", name, "error", err)
			continue
		}

		// 设置输出路径
		if outputFolder != folder {
			book.Out = filepath.Join(outputFolder, book.Bookname)
		}

		// 为单文件夹模式的书籍查找资源
		// 优先使用与书名相关的资源，其次使用通用资源
		info := BookInfo{Book: book}
		info = s.findResourcesForBook(info, folder, book.Bookname, globalResources)

		// 应用资源路径
		s.applyBookResources(book, info)

		books = append(books, info)
	}

	return books
}

// findGlobalResources 查找文件夹中的通用资源文件
func (s *ConverterService) findGlobalResources(folder string) map[string]string {
	resources := make(map[string]string)

	entries, err := os.ReadDir(folder)
	if err != nil {
		return resources
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		base := strings.ToLower(strings.TrimSuffix(name, ext))

		// 检查是否是通用资源命名
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".webp":
			// 通用封面命名
			if base == "cover" || base == "封面" || base == "bookcover" {
				resources["cover"] = filepath.Join(folder, name)
			}
			// 通用页眉命名
			if base == "header" || base == "页眉" || base == "chapter_header" {
				resources["header"] = filepath.Join(folder, name)
			}
		}
	}

	return resources
}

// scanSubFolder 扫描子文件夹（子文件夹模式）
func (s *ConverterService) scanSubFolder(subFolder, outputFolder string, parentGlobalResources map[string]string) []BookInfo {
	var books []BookInfo

	entries, err := os.ReadDir(subFolder)
	if err != nil {
		return books
	}

	// 查找子文件夹内的通用资源
	localResources := s.findGlobalResources(subFolder)

	// 如果子文件夹没有资源，继承父文件夹的资源
	if _, hasCover := localResources["cover"]; !hasCover {
		if parentCover, ok := parentGlobalResources["cover"]; ok {
			localResources["cover"] = parentCover
		}
	}
	if _, hasHeader := localResources["header"]; !hasHeader {
		if parentHeader, ok := parentGlobalResources["header"]; ok {
			localResources["header"] = parentHeader
		}
	}

	// 查找页眉图片文件夹
	headerFolder := ""
	for _, entry := range entries {
		if entry.IsDir() {
			name := strings.ToLower(entry.Name())
			if name == "headers" || name == "header" || name == "页眉" || name == "chapter_headers" {
				headerFolder = filepath.Join(subFolder, entry.Name())
				break
			}
		}
	}

	// 处理子文件夹内的txt文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}

		bookPath := filepath.Join(subFolder, name)
		book, err := model.NewBookSimple(bookPath)
		if err != nil {
			continue
		}

		// 设置输出路径
		if outputFolder != "" {
			book.Out = filepath.Join(outputFolder, book.Bookname)
		}

		info := BookInfo{
			Book:         book,
			CoverPath:    localResources["cover"],
			HeaderPath:   localResources["header"],
			HeaderFolder: headerFolder,
		}

		s.applyBookResources(book, info)
		books = append(books, info)
	}

	return books
}

// findResourcesForBook 为书籍查找资源（支持书名相关和通用资源）
func (s *ConverterService) findResourcesForBook(info BookInfo, folder, bookName string, globalResources map[string]string) BookInfo {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return info
	}

	// 1. 首先尝试查找与书名相关的资源
	_ = s.cleanBookNameForMatch(bookName)

	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否是页眉图片文件夹
			dirName := strings.ToLower(entry.Name())
			if strings.Contains(dirName, "header") || strings.Contains(dirName, "页眉") {
				// 检查是否与书名匹配
				if s.isResourceForBook(entry.Name(), bookName) {
					info.HeaderFolder = filepath.Join(folder, entry.Name())
				}
			}
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" && ext != ".webp" {
			continue
		}

		base := strings.TrimSuffix(name, ext)
		lowerBase := strings.ToLower(base)

		// 检查是否是封面
		if strings.Contains(lowerBase, "cover") || strings.Contains(lowerBase, "封面") {
			if s.isResourceForBook(base, bookName) {
				info.CoverPath = filepath.Join(folder, name)
			}
		}

		// 检查是否是页眉图片
		if strings.Contains(lowerBase, "header") || strings.Contains(lowerBase, "页眉") {
			if s.isResourceForBook(base, bookName) {
				info.HeaderPath = filepath.Join(folder, name)
			}
		}
	}

	// 2. 如果没找到书名相关资源，使用通用资源
	if info.CoverPath == "" {
		info.CoverPath = globalResources["cover"]
	}
	if info.HeaderPath == "" {
		info.HeaderPath = globalResources["header"]
	}

	// 3. 查找通用命名的页眉文件夹
	if info.HeaderFolder == "" {
		for _, entry := range entries {
			if entry.IsDir() {
				dirName := strings.ToLower(entry.Name())
				if dirName == "headers" || dirName == "header" || dirName == "页眉" {
					info.HeaderFolder = filepath.Join(folder, entry.Name())
					break
				}
			}
		}
	}

	return info
}

// isResourceForBook 检查资源是否属于某本书
func (s *ConverterService) isResourceForBook(resourceName, bookName string) bool {
	cleanResource := s.cleanBookNameForMatch(resourceName)
	cleanBook := s.cleanBookNameForMatch(bookName)

	// 精确匹配
	if cleanResource == cleanBook {
		return true
	}

	// 资源名包含书名
	if strings.Contains(cleanResource, cleanBook) {
		return true
	}

	// 书名包含资源名（去除cover/header等后缀后）
	resourceWithoutSuffix := s.removeResourceSuffix(cleanResource)
	if resourceWithoutSuffix != "" && strings.Contains(cleanBook, resourceWithoutSuffix) {
		return true
	}

	return false
}

// cleanBookNameForMatch 清理书名用于匹配
func (s *ConverterService) cleanBookNameForMatch(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "《", "")
	name = strings.ReplaceAll(name, "》", "")
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	return name
}

// removeResourceSuffix 移除资源文件的后缀（cover, header等）
func (s *ConverterService) removeResourceSuffix(name string) string {
	suffixes := []string{"cover", "封面", "header", "页眉", "chapterheader", "chapter_header"}
	for _, suffix := range suffixes {
		name = strings.ReplaceAll(name, suffix, "")
	}
	return strings.TrimSpace(name)
}

// applyBookResources 应用资源到书籍
func (s *ConverterService) applyBookResources(book *model.Book, info BookInfo) {
	if info.CoverPath != "" {
		book.Cover = info.CoverPath
	}
	if info.HeaderPath != "" {
		book.ChapterHeaderImage = info.HeaderPath
	}
	if info.HeaderFolder != "" {
		book.ChapterHeaderImageFolder = info.HeaderFolder
		book.ChapterHeaderImageMode = "folder"
	}
}

// convertBook 执行单本书的转换
func (s *ConverterService) convertBook(book *model.Book) error {
	if err := core.Check(book, s.version); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := core.Parse(book); err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}

	conv := converter.Dispatcher{Book: book}
	if err := conv.Convert(); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	return nil
}

// applyBookParameters 从参数应用到 Book 对象
func (s *ConverterService) applyBookParameters(book *model.Book, args map[string]interface{}) {
	// 基本信息
	if v, ok := args["bookname"].(string); ok && v != "" {
		book.Bookname = v
	}
	if v, ok := args["author"].(string); ok && v != "" {
		book.Author = v
	}
	if v, ok := args["format"].(string); ok && v != "" {
		book.Format = v
	}
	if v, ok := args["out"].(string); ok && v != "" {
		book.Out = v
	}

	// 章节匹配
	if v, ok := args["match"].(string); ok && v != "" {
		book.Match = v
	}
	if v, ok := args["volume_match"].(string); ok {
		book.VolumeMatch = v
	}
	if v, ok := args["exclude"].(string); ok && v != "" {
		book.ExclusionPattern = v
	}
	if v, ok := args["unknow_title"].(string); ok && v != "" {
		book.UnknowTitle = v
	}

	// 封面设置
	if v, ok := args["cover"].(string); ok && v != "" {
		book.Cover = v
	}
	if v, ok := args["cover_orly_color"].(string); ok && v != "" {
		book.CoverOrlyColor = v
	}
	if v, ok := args["cover_orly_idx"].(float64); ok {
		book.CoverOrlyIdx = int(v)
	}

	// 字体和样式
	if v, ok := args["font"].(string); ok && v != "" {
		book.Font = v
	}
	if v, ok := args["max"].(float64); ok {
		book.Max = uint(v)
	}
	if v, ok := args["indent"].(float64); ok {
		book.Indent = uint(v)
	}
	if v, ok := args["align"].(string); ok && v != "" {
		book.Align = v
	}
	if v, ok := args["bottom"].(string); ok && v != "" {
		book.Bottom = v
	}
	if v, ok := args["line_height"].(string); ok && v != "" {
		book.LineHeight = v
	}
	if v, ok := args["lang"].(string); ok && v != "" {
		book.Lang = v
	}
	if v, ok := args["tips"].(bool); ok {
		book.Tips = v
	}
	if v, ok := args["separate_chapter_number"].(bool); ok {
		book.SeparateChapterNumber = v
	}
	if v, ok := args["custom_css_file"].(string); ok && v != "" {
		book.CustomCSSFile = v
	}

	// 新增 - 扩展CSS样式
	if v, ok := args["extended_css"].(string); ok && v != "" {
		book.ExtendedCSS = v
	}
	if v, ok := args["css_variables"].(string); ok && v != "" {
		book.CSSVariables = v
	}

	// 新增 - 章节页眉图片
	if v, ok := args["chapter_header_image"].(string); ok && v != "" {
		book.ChapterHeaderImage = v
	}
	if v, ok := args["chapter_header_image_folder"].(string); ok && v != "" {
		book.ChapterHeaderImageFolder = v
	}
	if v, ok := args["chapter_header_image_position"].(string); ok && v != "" {
		book.ChapterHeaderImagePosition = v
	}
	if v, ok := args["chapter_header_image_height"].(string); ok && v != "" {
		book.ChapterHeaderImageHeight = v
	}
	if v, ok := args["chapter_header_image_width"].(string); ok && v != "" {
		book.ChapterHeaderImageWidth = v
	}
	if v, ok := args["chapter_header_image_mode"].(string); ok && v != "" {
		book.ChapterHeaderImageMode = v
	}
}

// extractGlobalParams 提取全局样式参数
func (s *ConverterService) extractGlobalParams(args map[string]interface{}) map[string]interface{} {
	globalParams := make(map[string]interface{})

	globalKeys := []string{
		"format", "font", "align", "indent", "bottom",
		"line_height", "lang", "separate_chapter_number",
		"custom_css_file", "extended_css", "css_variables",
	}

	for _, key := range globalKeys {
		if v, ok := args[key]; ok {
			globalParams[key] = v
		}
	}

	return globalParams
}

// applyGlobalParams 应用全局参数到书籍
func (s *ConverterService) applyGlobalParams(book *model.Book, params map[string]interface{}) {
	if v, ok := params["format"].(string); ok && v != "" {
		book.Format = v
	}
	if v, ok := params["font"].(string); ok && v != "" {
		book.Font = v
	}
	if v, ok := params["align"].(string); ok && v != "" {
		book.Align = v
	}
	if v, ok := params["indent"].(float64); ok {
		book.Indent = uint(v)
	}
	if v, ok := params["bottom"].(string); ok && v != "" {
		book.Bottom = v
	}
	if v, ok := params["line_height"].(string); ok && v != "" {
		book.LineHeight = v
	}
	if v, ok := params["lang"].(string); ok && v != "" {
		book.Lang = v
	}
	if v, ok := params["separate_chapter_number"].(bool); ok {
		book.SeparateChapterNumber = v
	}
	if v, ok := params["custom_css_file"].(string); ok && v != "" {
		book.CustomCSSFile = v
	}
	if v, ok := params["extended_css"].(string); ok && v != "" {
		book.ExtendedCSS = v
	}
	if v, ok := params["css_variables"].(string); ok && v != "" {
		book.CSSVariables = v
	}
}

// getOutputFiles 获取输出文件列表
func (s *ConverterService) getOutputFiles(book *model.Book) []string {
	var outputFiles []string
	formats := []string{"epub"}
	if book.Format == "all" || book.Format == "" {
		formats = []string{"epub", "mobi", "azw3"}
	} else if book.Format != "epub" {
		formats = []string{book.Format}
	}

	for _, format := range formats {
		fpath := book.Out + "." + format
		if info, err := os.Stat(fpath); err == nil {
			absPath, _ := filepath.Abs(fpath)
			outputFiles = append(outputFiles, fmt.Sprintf("- [%s](%s) (%d bytes)", filepath.Base(fpath), absPath, info.Size()))
		}
	}

	return outputFiles
}

// getFormatsStr 获取格式字符串
func (s *ConverterService) getFormatsStr(format string) string {
	if format == "all" || format == "" {
		return "epub, mobi, azw3"
	}
	return format
}

// registerPrompts 注册提示词模板
func (s *ConverterService) registerPrompts(srv *server.MCPServer) {
	// 官方网站提示词
	website := mcpgo.NewPrompt("kaf-mcp", mcpgo.WithPromptDescription("kaf-mcp官方网站和代码仓库"))
	srv.AddPrompt(website, func(ctx context.Context, request mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
		return mcpgo.NewGetPromptResult(
			"kaf-mcp官方网站",
			[]mcpgo.PromptMessage{
				mcpgo.NewPromptMessage(
					mcpgo.RoleAssistant,
					mcpgo.NewTextContent("https://github.com/feewg/kaf-cli"),
				),
			},
		), nil
	})

	// 批量转换规范提示词
	batchPrompt := mcpgo.NewPrompt("batch_convert_guide", mcpgo.WithPromptDescription("文件夹批量转换规范说明"))
	srv.AddPrompt(batchPrompt, func(ctx context.Context, request mcpgo.GetPromptRequest) (*mcpgo.GetPromptResult, error) {
		guide := `文件夹批量转换规范:

支持两种组织结构：

1. 单文件夹模式（所有书籍在一个文件夹）:
   novels/
   ├── cover.jpg              # 通用封面（所有书籍使用）
   ├── header.png             # 通用页眉（所有书籍使用）
   ├── 斗破苍穹.txt
   ├── 武动乾坤.txt
   └── ...

2. 子文件夹模式（每本书独立文件夹）:
   novels/
   ├── 斗破苍穹/
   │   ├── book.txt           # 小说文件（任意命名）
   │   ├── cover.jpg          # 专属封面
   │   └── header.png         # 专属页眉
   ├── 武动乾坤/
   │   ├── book.txt
   │   ├── cover.jpg
   │   └── headers/           # 章节页眉图片文件夹
   │       ├── 第一章.png
   │       └── ...
   └── ...

资源文件命名规则:
- 封面: cover.jpg, cover.png, 封面.jpg 等
- 页眉: header.jpg, header.png, 页眉.jpg 等
- 页眉文件夹: headers/, header/, 页眉/

匹配优先级:
1. 子文件夹内资源优先于父文件夹通用资源
2. 书名相关命名优先于通用命名
3. 精确匹配优先于模糊匹配`

		return mcpgo.NewGetPromptResult(
			"批量转换规范",
			[]mcpgo.PromptMessage{
				mcpgo.NewPromptMessage(
					mcpgo.RoleAssistant,
					mcpgo.NewTextContent(guide),
				),
			},
		), nil
	})
}
