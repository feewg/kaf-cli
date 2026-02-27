package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/feewg/kaf-cli/internal/config"
	"github.com/feewg/kaf-cli/internal/converter"
	"github.com/feewg/kaf-cli/internal/core"
	"github.com/feewg/kaf-cli/internal/model"
	"github.com/feewg/kaf-cli/internal/utils"
	"github.com/feewg/kaf-cli/pkg/analytics"
)

var (
	secret      string
	measurement string
	version     string
)

// CLIConfig 命令行全局配置
type CLIConfig struct {
	ConfigPath string // 指定的配置文件路径
}

func NewBookArgs(cliCfg *CLIConfig) *model.Book {
	var book model.Book
	flag.StringVar(&book.Filename, "filename", "", "txt 文件名")
	flag.StringVar(&book.Bookname, "bookname", "", "书名: 默认为txt文件名")
	flag.StringVar(&book.Author, "author", "YSTYLE", "作者")
	flag.StringVar(&book.Match, "match", "", "匹配标题的正则表达式, 不写可以自动识别, 如果没生成章节就参考教程。例: -match 第.{1,8}章 表示第和章字之间可以有1-8个任意文字")
	flag.StringVar(&book.VolumeMatch, "volume-match", model.VolumeMatch, "卷匹配规则,设置为false可以禁用卷识别")
	flag.StringVar(&book.ExclusionPattern, "exclude", model.DefaultExclusion, "排除无效章节/卷的正则表达式")
	flag.StringVar(&book.UnknowTitle, "unknow-title", "章节正文", "未知章节默认名称")
	flag.StringVar(&book.Cover, "cover", "cover.png", "封面图片可为: 本地图片, 和orly。 设置为orly时生成orly风格的封面, 需要连接网络。")
	flag.StringVar(&book.CoverOrlyColor, "cover-orly-color", "", "orly封面的主题色, 可以为1-16和hex格式的颜色代码, 不填时随机")
	flag.IntVar(&book.CoverOrlyIdx, "cover-orly-idx", -1, "orly封面的动物, 可以为0-41, 不填时随机, 具体图案可以查看: https://orly.nanmu.me")
	flag.UintVar(&book.Max, "max", 35, "标题最大字数")
	flag.UintVar(&book.Indent, "indent", 2, "段落缩进字数")
	flag.StringVar(&book.Align, "align", utils.GetEnv("KAF_CLI_ALIGN", "center"), "标题对齐方式: left、center、righ。环境变量KAF_CLI_ALIGN可修改默认值")
	flag.StringVar(&book.Bottom, "bottom", "1em", "段落间距(单位可以为em、px)")
	flag.StringVar(&book.LineHeight, "line-height", "", "行高(用于设置行间距, 默认为1.5rem)")
	flag.StringVar(&book.Font, "font", "", "嵌入字体, 之后epub的正文都将使用该字体")
	flag.StringVar(&book.Lang, "lang", utils.GetEnv("KAF_CLI_LANG", "zh"), "设置语言: en,de,fr,it,es,zh,ja,pt,ru,nl。环境变量KAF_CLI_LANG可修改默认值")
	flag.StringVar(&book.Format, "format", utils.GetEnv("KAF_CLI_FORMAT", "all"), "书籍格式: all、epub、mobi、azw3。环境变量KAF_CLI_FORMAT可修改默认值")
	flag.StringVar(&book.Out, "out", "", "输出文件名，不需要包含格式后缀")
	flag.BoolVar(&book.Tips, "tips", true, "添加本软件教程")
	flag.BoolVar(&book.SeparateChapterNumber, "separate-chapter-number", false, "是否分离章节序号和标题样式（序号单独一行显示）")
	flag.StringVar(&book.CustomCSSFile, "custom-css-file", "", "自定义 CSS 文件路径，用于覆盖默认样式")

	// 扩展CSS样式支持
	flag.StringVar(&book.ExtendedCSS, "extended-css", "", "内联扩展CSS样式（直接传入CSS代码）")
	flag.StringVar(&book.CSSVariables, "css-variables", "", "CSS变量定义，格式: --var1:value1;--var2:value2")

	// 章节页眉图片支持
	flag.StringVar(&book.ChapterHeaderImage, "chapter-header-image", "", "章节页眉图片路径，所有章节显示相同图片")
	flag.StringVar(&book.ChapterHeaderImageFolder, "chapter-header-image-folder", "", "章节页眉图片文件夹，按章节名匹配图片")
	flag.StringVar(&book.ChapterHeaderImagePosition, "chapter-header-image-position", "center", "页眉图片位置: left, center, right")
	flag.StringVar(&book.ChapterHeaderImageHeight, "chapter-header-image-height", "auto", "页眉图片高度，如: 100px, 2em")
	flag.StringVar(&book.ChapterHeaderImageWidth, "chapter-header-image-width", "100%", "页眉图片宽度，如: 50%, 200px")
	flag.StringVar(&book.ChapterHeaderImageMode, "chapter-header-image-mode", "single", "图片模式: single(所有章节相同), folder(按章节名匹配)")

	// YAML 配置文件支持
	flag.StringVar(&cliCfg.ConfigPath, "config", "", "YAML 配置文件路径，自动识别时可不指定")

	flag.Parse()
	return &book
}

func printHelp(version string) {
	fmt.Println("错误: 文件名不能为空")
	fmt.Println("软件版本: 	", version)
	fmt.Println("简洁模式: 	把文件拖放到kaf-cli上")
	fmt.Println("命令行简单模式: kaf-cli ebook.txt")
	fmt.Println("\n以下为kaf-cli的全部参数")
	var cliCfg CLIConfig
	NewBookArgs(&cliCfg)
	flag.PrintDefaults()
	fmt.Println("\nYAML 配置支持:")
	fmt.Println("  1. 使用 -config 指定配置文件: kaf-cli -config kaf.yaml")
	fmt.Println("  2. 自动识别: 将 kaf.yaml 放在txt文件同级目录下会自动加载")
	fmt.Println("  3. 生成示例配置: kaf-cli -example-config")
	if runtime.GOOS == "windows" {
		time.Sleep(time.Second * 10)
	}
}

func main() {
	// 检查是否是生成示例配置
	if len(os.Args) == 2 && os.Args[1] == "-example-config" {
		if err := os.WriteFile("kaf.yaml", []byte(config.ExampleConfig()), 0644); err != nil {
			fmt.Printf("生成示例配置失败: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println("已生成示例配置文件: kaf.yaml")
		fmt.Println("请根据需要进行修改，然后放在txt文件同级目录下即可自动识别")
		return
	}

	// 检查是否是批量处理模式
	if len(os.Args) == 3 && os.Args[1] == "-batch" {
		batchFolder := os.Args[2]
		runBatchConvert(batchFolder)
		return
	}

	var book *model.Book
	var err error
	var yamlCfg *config.Config
	var cliCfg CLIConfig

	if len(os.Args) == 2 && strings.HasSuffix(os.Args[1], ".txt") {
		// 简洁模式: kaf-cli ebook.txt
		// 自动查找配置文件
		filename := os.Args[1]
		yamlCfg, _, _ = config.AutoLoadForFile(filename)
		book, err = model.NewBookSimple(filename)
		if err != nil {
			fmt.Printf("错误: %s\n", err.Error())
			os.Exit(1)
		}
		// 应用 YAML 配置
		if yamlCfg != nil {
			yamlCfg.MergeWithBook(book)
			fmt.Printf("已加载配置文件\n")
		}
	} else {
		// 命令行模式
		book = NewBookArgs(&cliCfg)

		// 如果指定了配置文件，加载它
		if cliCfg.ConfigPath != "" {
			yamlCfg, err = config.LoadFromFile(cliCfg.ConfigPath)
			if err != nil {
				fmt.Printf("加载配置文件失败: %s\n", err.Error())
				os.Exit(1)
			}
			yamlCfg.MergeWithBook(book)
			fmt.Printf("已加载配置文件: %s\n", cliCfg.ConfigPath)
		} else if book.Filename != "" {
			// 如果没有指定配置，但指定了文件名，尝试自动查找
			yamlCfg, cfgPath, err := config.AutoLoadForFile(book.Filename)
			if err == nil && yamlCfg != nil {
				yamlCfg.MergeWithBook(book)
				fmt.Printf("已加载配置文件: %s\n", cfgPath)
			}
		}
	}

	if err := core.Check(book, version); err != nil {
		if err.Error() == "不是txt文件" {
			fmt.Printf("错误: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println(err)
		printHelp(version)
		os.Exit(1)
	}
	analytics.Analytics(version, secret, measurement, book.Format)
	book.ToString()
	if err := core.Parse(book); err != nil {
		fmt.Printf("错误: %s\n", err.Error())
		os.Exit(2)
	}
	conv := converter.Dispatcher{
		Book: book,
	}
	if err := conv.Convert(); err != nil {
		fmt.Printf("错误: %s\n", err.Error())
		os.Exit(1)
	}
}

// runBatchConvert 执行批量转换
func runBatchConvert(folder string) {
	// 检查文件夹是否存在
	if info, err := os.Stat(folder); os.IsNotExist(err) || !info.IsDir() {
		fmt.Printf("错误: 文件夹不存在或不是目录: %s\n", folder)
		os.Exit(1)
	}

	fmt.Println("正在扫描文件夹:", folder)
	books := scanBooks(folder, folder)

	if len(books) == 0 {
		fmt.Println("未找到符合规范的txt文件。")
		fmt.Println("\n支持的文件夹结构:")
		fmt.Println("1. 单文件夹模式: 所有txt文件在同一文件夹，可包含通用cover.jpg/header.png")
		fmt.Println("2. 子文件夹模式: 每本小说一个子文件夹，子文件夹内包含独立的资源文件")
		os.Exit(1)
	}

	fmt.Printf("找到 %d 本书籍，开始批量转换...\n\n", len(books))

	var successCount, failCount int
	for i, info := range books {
		fmt.Printf("[%d/%d] 正在转换: %s\n", i+1, len(books), info.Book.Bookname)

		// 应用从文件夹扫描时加载的配置（类似于通用封面的逻辑）
		if info.Config != nil {
			info.Config.MergeWithBook(info.Book)
		}

		// 尝试加载书籍同目录的 YAML 配置（优先级高于文件夹通用配置）
		if yamlCfg, cfgPath, err := config.AutoLoadForFile(info.Book.Filename); err == nil && yamlCfg != nil {
			yamlCfg.MergeWithBook(info.Book)
			fmt.Printf("  📄 配置: %s\n", filepath.Base(cfgPath))
		}

		if err := convertBook(info.Book, version); err != nil {
			fmt.Printf("  ❌ 失败: %s\n", err.Error())
			failCount++
		} else {
			fmt.Printf("  ✅ 成功\n")
			successCount++
		}
	}

	fmt.Println("\n=================================")
	fmt.Printf("批量转换完成！成功: %d, 失败: %d, 总计: %d\n", successCount, failCount, len(books))
}

// convertBook 执行单本书的转换
func convertBook(book *model.Book, version string) error {
	if err := core.Check(book, version); err != nil {
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
