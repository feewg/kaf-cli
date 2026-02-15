package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/feewg/kaf-cli/internal/converter"
	"github.com/feewg/kaf-cli/internal/core"
	"github.com/feewg/kaf-cli/internal/model"
)

type ConverterService struct {
	version string
}

func (s *ConverterService) RegisterTools(srv *server.MCPServer, version string) {
	s.version = version

	// 注册转换工具
	s.registerConvertTool(srv)

	// 添加提示词模板
	s.registerPrompts(srv)
}

// registerConvertTool 注册转换工具，支持所有 CLI 参数
func (s *ConverterService) registerConvertTool(srv *server.MCPServer) {
	tool := mcp.NewTool("kaf_convert",
		mcp.WithDescription("电子书格式转换器，支持把txt文件转换成epub/mobi/azw3电子书格式\n转换成功后返回生成的电子书文件路径"),
		// 必填参数
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("txt小说文件路径，支持相对路径和绝对路径"),
		),
		// 可选参数 - 基本信息
		mcp.WithString("bookname",
			mcp.Description("书名，为空时自动从文件名识别"),
		),
		mcp.WithString("author",
			mcp.Description("作者，为空时自动从文件名识别，默认YSTYLE"),
		),
		mcp.WithString("format",
			mcp.Description("输出格式: all(全部), epub, mobi, azw3，默认all"),
		),
		mcp.WithString("out",
			mcp.Description("输出文件名(不含后缀)，默认为书名"),
		),
		// 可选参数 - 章节匹配
		mcp.WithString("match",
			mcp.Description("章节匹配正则表达式，不填自动识别"),
		),
		mcp.WithString("volume_match",
			mcp.Description("卷匹配规则，默认启用，设为false禁用"),
		),
		mcp.WithString("exclude",
			mcp.Description("排除无效章节的正则表达式"),
		),
		mcp.WithString("unknow_title",
			mcp.Description("未知章节默认名称，默认'章节正文'"),
		),
		// 可选参数 - 样式设置
		mcp.WithString("cover",
			mcp.Description("封面图片: 本地路径、orly(在线生成)、none(无封面)"),
		),
		mcp.WithString("cover_orly_color",
			mcp.Description("orly封面主题色: 1-16或hex颜色代码"),
		),
		mcp.WithNumber("cover_orly_idx",
			mcp.Description("orly封面动物图案: 0-41"),
		),
		mcp.WithString("font",
			mcp.Description("嵌入字体文件路径"),
		),
		mcp.WithNumber("max",
			mcp.Description("标题最大字数，默认35"),
		),
		mcp.WithNumber("indent",
			mcp.Description("段落缩进字数，默认2"),
		),
		mcp.WithString("align",
			mcp.Description("标题对齐: left, center, right，默认center"),
		),
		mcp.WithString("bottom",
			mcp.Description("段落间距，默认1em"),
		),
		mcp.WithString("line_height",
			mcp.Description("行高，默认1.5rem"),
		),
		mcp.WithString("lang",
			mcp.Description("语言: en,de,fr,it,es,zh,ja,pt,ru,nl，默认zh"),
		),
		mcp.WithBoolean("tips",
			mcp.Description("添加制作教程，默认true"),
		),
		mcp.WithBoolean("separate_chapter_number",
			mcp.Description("分离章节序号和标题样式，默认false"),
		),
		mcp.WithString("custom_css_file",
			mcp.Description("自定义CSS文件路径"),
		),
	)

	srv.AddTool(tool, s.handleConvert)
}

// handleConvert 处理转换请求
func (s *ConverterService) handleConvert(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	// 如果格式是 all，可能会有多个输出文件
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
		strings.Join(formats, ", "),
		strings.Join(outputFiles, "\n"),
	)

	return mcp.NewToolResultText(resultText), nil
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
}

// registerPrompts 注册提示词模板
func (s *ConverterService) registerPrompts(srv *server.MCPServer) {
	// 官方网站提示词
	website := mcp.NewPrompt("kaf-mcp", mcp.WithPromptDescription("kaf-mcp官方网站和代码仓库"))
	srv.AddPrompt(website, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return mcp.NewGetPromptResult(
			"kaf-mcp官方网站",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleAssistant,
					mcp.NewTextContent("https://github.com/feewg/kaf-cli"),
				),
			},
		), nil
	})
}
