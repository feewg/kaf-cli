package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/feewg/kaf-cli/internal/model"
	"gopkg.in/yaml.v3"
)

// Config YAML 配置文件结构
type Config struct {
	// 基础配置
	Filename    string `yaml:"filename"`    // txt 文件名
	Bookname    string `yaml:"bookname"`    // 书名
	Author      string `yaml:"author"`      // 作者
	Match       string `yaml:"match"`       // 匹配标题的正则表达式
	VolumeMatch string `yaml:"volume_match"` // 卷匹配规则
	Exclude     string `yaml:"exclude"`     // 排除无效章节的正则表达式
	UnknowTitle string `yaml:"unknow_title"` // 未知章节默认名称

	// 封面配置
	Cover          string `yaml:"cover"`            // 封面图片
	CoverOrlyColor string `yaml:"cover_orly_color"` // orly封面主题色
	CoverOrlyIdx   int    `yaml:"cover_orly_idx"`   // orly封面动物索引

	// 排版配置
	Max        uint   `yaml:"max"`         // 标题最大字数
	Indent     uint   `yaml:"indent"`      // 段落缩进字数
	Align      string `yaml:"align"`       // 标题对齐方式
	Bottom     string `yaml:"bottom"`      // 段落间距
	LineHeight string `yaml:"line_height"` // 行高
	Font       string `yaml:"font"`        // 嵌入字体

	// 语言和格式
	Lang   string `yaml:"lang"`   // 语言
	Format string `yaml:"format"` // 书籍格式
	Out    string `yaml:"out"`    // 输出文件名

	// 其他选项
	Tips                  bool `yaml:"tips"`                     // 添加教程
	SeparateChapterNumber bool `yaml:"separate_chapter_number"`  // 分离章节序号和标题样式

	// 自定义CSS
	CustomCSSFile string `yaml:"custom_css_file"` // 自定义CSS文件路径
	ExtendedCSS   string `yaml:"extended_css"`    // 内联扩展CSS样式
	CSSVariables  string `yaml:"css_variables"`   // CSS变量定义

	// 章节页眉图片
	ChapterHeaderImage         string `yaml:"chapter_header_image"`          // 章节页眉图片路径
	ChapterHeaderImageFolder   string `yaml:"chapter_header_image_folder"`   // 章节页眉图片文件夹
	ChapterHeaderImagePosition string `yaml:"chapter_header_image_position"` // 页眉图片位置
	ChapterHeaderImageHeight   string `yaml:"chapter_header_image_height"`   // 页眉图片高度
	ChapterHeaderImageWidth    string `yaml:"chapter_header_image_width"`    // 页眉图片宽度
	ChapterHeaderImageMode     string `yaml:"chapter_header_image_mode"`     // 图片模式
}

// DefaultConfigNames 默认配置文件名（按优先级排序）
var DefaultConfigNames = []string{
	"kaf.yaml",
	"kaf.yml",
	".kaf.yaml",
	".kaf.yml",
}

// LoadFromFile 从指定路径加载 YAML 配置文件
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// LoadFromString 从字符串加载 YAML 配置
func LoadFromString(content string) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析配置内容失败: %w", err)
	}
	return &cfg, nil
}

// AutoLoad 自动查找并加载配置文件
// 查找顺序: 1. 指定文件夹下的配置文件 2. 当前目录下的配置文件
func AutoLoad(dir string) (*Config, string, error) {
	// 如果指定了目录，先在该目录下查找
	if dir != "" {
		for _, name := range DefaultConfigNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				cfg, err := LoadFromFile(path)
				return cfg, path, err
			}
		}
	}

	// 在当前目录查找
	for _, name := range DefaultConfigNames {
		if _, err := os.Stat(name); err == nil {
			cfg, err := LoadFromFile(name)
			return cfg, name, err
		}
	}

	return nil, "", fmt.Errorf("未找到配置文件")
}

// AutoLoadForFile 为指定文件自动查找配置文件
// 查找顺序: 1. 文件所在目录 2. 当前目录
func AutoLoadForFile(filePath string) (*Config, string, error) {
	// 获取文件所在目录
	dir := filepath.Dir(filePath)
	if dir == "" {
		dir = "."
	}

	// 先在文件所在目录查找
	for _, name := range DefaultConfigNames {
		configPath := filepath.Join(dir, name)
		if _, err := os.Stat(configPath); err == nil {
			cfg, err := LoadFromFile(configPath)
			return cfg, configPath, err
		}
	}

	// 再在当前目录查找
	return AutoLoad("")
}

// LoadFromFolder 从指定文件夹中查找并加载配置文件
// 类似于查找封面图片的逻辑，在文件夹中查找通用配置文件
func LoadFromFolder(folder string) (*Config, string, error) {
	// 在指定文件夹中查找
	for _, name := range DefaultConfigNames {
		configPath := filepath.Join(folder, name)
		if _, err := os.Stat(configPath); err == nil {
			cfg, err := LoadFromFile(configPath)
			return cfg, configPath, err
		}
	}

	return nil, "", fmt.Errorf("在文件夹 %s 中未找到配置文件", folder)
}

// MergeWithBook 将配置合并到 Book 对象
// 优先级: 1. 命令行参数（已设置的值） 2. YAML 配置 3. 默认值
func (c *Config) MergeWithBook(book *model.Book) {
	if c == nil || book == nil {
		return
	}

	// 使用反射遍历配置字段
	cfgVal := reflect.ValueOf(c).Elem()
	bookVal := reflect.ValueOf(book).Elem()
	cfgType := cfgVal.Type()

	for i := 0; i < cfgVal.NumField(); i++ {
		field := cfgType.Field(i)
		cfgField := cfgVal.Field(i)
		bookField := bookVal.FieldByName(field.Name)

		// 如果 Book 中对应的字段不存在，跳过
		if !bookField.IsValid() {
			continue
		}

		// 获取 YAML 标签中的字段名
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		yamlTag = strings.Split(yamlTag, ",")[0]

		// 根据字段类型合并值
		switch cfgField.Kind() {
		case reflect.String:
			mergeStringField(bookField, cfgField.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			mergeIntField(bookField, cfgField.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			mergeUintField(bookField, cfgField.Uint())
		case reflect.Bool:
			mergeBoolField(bookField, cfgField.Bool())
		}
	}

	// 特殊字段映射（YAML字段名和Book字段名不一致的情况）
	if c.Exclude != "" && book.ExclusionPattern == "" {
		book.ExclusionPattern = c.Exclude
	}
	if c.UnknowTitle != "" && book.UnknowTitle == "章节正文" {
		book.UnknowTitle = c.UnknowTitle
	}
}

// mergeStringField 合并字符串字段（只覆盖默认值）
func mergeStringField(field reflect.Value, value string) {
	if value == "" {
		return
	}
	// 如果当前是零值或默认值，则覆盖
	if field.String() == "" {
		field.SetString(value)
	}
}

// mergeIntField 合并整数字段
func mergeIntField(field reflect.Value, value int64) {
	// 对于 int 类型，-1 通常表示未设置
	if field.Int() == -1 || field.Int() == 0 {
		field.SetInt(value)
	}
}

// mergeUintField 合并无符号整数字段
func mergeUintField(field reflect.Value, value uint64) {
	if field.Uint() == 0 {
		field.SetUint(value)
	}
}

// mergeBoolField 合并布尔字段
func mergeBoolField(field reflect.Value, value bool) {
	// 布尔值只有在使用零值语义时才覆盖
	// 注意：由于 bool 零值是 false，无法区分"未设置"和"设置为 false"
	// 这里我们保守处理：只有当当前值为 false 时才覆盖
	if !field.Bool() {
		field.SetBool(value)
	}
}

// ToBook 将配置转换为 Book 对象
func (c *Config) ToBook(filename string) *model.Book {
	book := &model.Book{
		Filename:                   filename,
		Bookname:                   c.Bookname,
		Author:                     c.Author,
		Match:                      c.Match,
		VolumeMatch:                c.VolumeMatch,
		ExclusionPattern:           c.Exclude,
		UnknowTitle:                c.UnknowTitle,
		Cover:                      c.Cover,
		CoverOrlyColor:             c.CoverOrlyColor,
		CoverOrlyIdx:               c.CoverOrlyIdx,
		Max:                        c.Max,
		Indent:                     c.Indent,
		Align:                      c.Align,
		Bottom:                     c.Bottom,
		LineHeight:                 c.LineHeight,
		Font:                       c.Font,
		Lang:                       c.Lang,
		Format:                     c.Format,
		Out:                        c.Out,
		Tips:                       c.Tips,
		SeparateChapterNumber:      c.SeparateChapterNumber,
		CustomCSSFile:              c.CustomCSSFile,
		ExtendedCSS:                c.ExtendedCSS,
		CSSVariables:               c.CSSVariables,
		ChapterHeaderImage:         c.ChapterHeaderImage,
		ChapterHeaderImageFolder:   c.ChapterHeaderImageFolder,
		ChapterHeaderImagePosition: c.ChapterHeaderImagePosition,
		ChapterHeaderImageHeight:   c.ChapterHeaderImageHeight,
		ChapterHeaderImageWidth:    c.ChapterHeaderImageWidth,
		ChapterHeaderImageMode:     c.ChapterHeaderImageMode,
	}

	model.SetDefault(book)
	return book
}

// ExampleConfig 返回示例配置内容
func ExampleConfig() string {
	return `# KAF-CLI 配置文件示例
# 将此文件保存为 kaf.yaml 放在txt文件同级目录下即可自动识别

# 基础配置
filename: ""
bookname: ""
author: "YSTYLE"
match: ""
volume_match: "^第[0-9一二三四五六七八九十零〇百千两 ]+[卷部]"
exclude: "^第[0-9一二三四五六七八九十零〇百千两 ]+(部门|部队|部属|部分|部件|部落|部.*：$)"
unknow_title: "章节正文"

# 封面配置
cover: "cover.png"
cover_orly_color: ""
cover_orly_idx: -1

# 排版配置
max: 35
indent: 2
align: "center"
bottom: "1em"
line_height: ""
font: ""

# 语言和格式
lang: "zh"
format: "all"
out: ""

# 其他选项
tips: true
separate_chapter_number: false

# 自定义CSS
custom_css_file: ""
extended_css: ""
css_variables: ""

# 章节页眉图片
chapter_header_image: ""
chapter_header_image_folder: ""
chapter_header_image_position: "center"
chapter_header_image_height: "auto"
chapter_header_image_width: "100%"
chapter_header_image_mode: "single"
`
}
