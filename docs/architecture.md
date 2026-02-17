# kaf-cli 架构文档

本文档描述 kaf-cli 的项目架构、模块设计和代码组织方式。

## 1. 项目结构

```
kaf-cli/
├── cmd/                    # 应用程序入口
│   ├── cli/               # 命令行版本
│   │   └── main.go        # CLI入口点
│   └── mcp/               # MCP服务器版本
│       └── main.go        # MCP入口点
├── internal/              # 内部包（私有）
│   ├── converter/         # 格式转换器
│   │   ├── interface.go   # 转换器接口
│   │   ├── dispatcher.go  # 调度器
│   │   ├── epub.go        # EPUB转换器
│   │   ├── mobi.go        # MOBI转换器
│   │   ├── mobi_utils.go  # MOBI工具函数
│   │   └── azw3.go        # AZW3转换器
│   ├── core/              # 核心逻辑
│   │   ├── convert.go     # 转换流程控制
│   │   └── parser.go      # 文本解析
│   ├── model/             # 数据模型
│   │   └── book.go        # Book和Section定义
│   ├── mcp/               # MCP相关
│   │   ├── service.go     # MCP服务
│   │   └── logger.go      # MCP日志
│   └── utils/             # 工具函数
│       ├── cover.go       # 封面生成
│       ├── defaults.go    # 默认值处理
│       ├── env.go         # 环境变量
│       ├── exec.go        # 执行命令
│       ├── file.go        # 文件操作
│       ├── html.go        # HTML处理
│       ├── kindle.go      # Kindle工具
│       ├── lang.go        # 语言处理
│       └── regex.go       # 正则工具
├── docs/                  # 文档
│   ├── features.md        # 功能文档
│   ├── architecture.md    # 架构文档（本文档）
│   └── development.md     # 开发指南
├── lib/                   # 库文件
├── assets/                # 静态资源
└── kaf.go                 # 库入口点
```

## 2. 模块设计

### 2.1 入口层 (cmd/)

#### CLI入口 (cmd/cli/main.go)
- **职责**: 命令行参数解析、程序启动
- **主要功能**:
  - 解析命令行参数（flag）
  - 支持拖放模式（单文件参数）
  - 调用核心转换流程
- **关键函数**:
  - `NewBookArgs()`: 创建Book并绑定flag
  - `main()`: 程序入口

#### MCP入口 (cmd/mcp/main.go)
- **职责**: MCP服务器启动
- **主要功能**:
  - 初始化MCP服务
  - 提供AI助手接口

### 2.2 核心层 (internal/core/)

#### 转换控制 (convert.go)
- **职责**: 转换流程的编排
- **主要函数**:
  - `Check()`: 预检查，验证输入、解析信息、处理封面
  - `validateInput()`: 验证输入文件
  - `parseBookInfoFromFilename()`: 从文件名提取书名作者
  - `setDefaultValues()`: 设置默认值
  - `handleCover()`: 处理封面（本地/orly/无）
  - `compileRegex()`: 编译正则表达式

#### 文本解析 (parser.go)
- **职责**: 读取并解析TXT文件
- **主要函数**:
  - `Parse()`: 主解析函数
  - `readBuffer()`: 读取文件并处理编码
  - `sanitizeHTMLTags()`: 智能HTML标签处理
- **解析流程**:
  1. 读取文件并检测编码
  2. 逐行读取内容
  3. 识别章节标题（卷/章节）
  4. 构建Section列表
  5. 组织卷-章节结构

### 2.3 转换器层 (internal/converter/)

#### 接口定义 (interface.go)
```go
type Converter interface {
    Build(book model.Book) error
}
```

#### 调度器 (dispatcher.go)
- **职责**: 根据配置选择合适的转换器
- **逻辑**:
  - 根据`format`参数确定输出格式
  - 检查kindlegen可用性
  - 依次调用各格式转换器

#### EPUB转换器 (epub.go)
- **职责**: 生成EPUB格式电子书
- **核心功能**:
  - 构建EPUB结构（go-epub库）
  - 生成CSS样式
  - 处理章节页眉图片
  - 添加封面
- **关键函数**:
  - `Build()`: 主构建函数
  - `wrapTitle()`: 包装章节标题
  - `parseChapterTitle()`: 解析章节标题
  - `generateHeaderImageHTML()`: 生成页眉图片HTML（新增）
  - `findChapterHeaderImage()`: 查找章节页眉图片（新增）

#### MOBI转换器 (mobi.go)
- **职责**: 使用第三方库生成MOBI
- **说明**: 当kindlegen不可用时使用

#### AZW3转换器 (azw3.go)
- **职责**: 生成AZW3格式
- **特点**: 支持大文件分卷

### 2.4 模型层 (internal/model/)

#### Book结构体
- **核心字段**:
  - 基本信息：Filename, Bookname, Author
  - 匹配规则：Match, VolumeMatch, ExclusionPattern
  - 样式设置：Align, Indent, Bottom, LineHeight, Font
  - 封面设置：Cover, CoverOrlyColor, CoverOrlyIdx
  - CSS相关：CustomCSSFile, ExtendedCSS, CSSVariables（新增）
  - 页眉图片：ChapterHeaderImage等（新增）
- **方法**:
  - `SetDefault()`: 设置默认值
  - `ToString()`: 输出转换信息

#### Section结构体
- **字段**: Title, Content, Sections（子章节）
- **用途**: 表示卷或章节

### 2.5 工具层 (internal/utils/)

#### 封面生成 (cover.go)
- 本地封面验证
- Orly封面在线生成

#### 文件操作 (file.go)
- `IsExists()`: 检查文件存在
- `AddPart()`: 添加段落内容

#### HTML处理 (html.go)
- HTML标签转义
- 属性处理

#### Kindle工具 (kindle.go)
- `LookKindlegen()`: 查找kindlegen工具

## 3. 数据流

### 3.1 转换流程

```
用户输入 → 参数解析 → 预检查 → 文本解析 → 格式转换 → 输出文件
```

详细步骤：
1. **参数解析** (cmd/cli/main.go)
   - 解析命令行参数
   - 创建Book实例

2. **预检查** (core/convert.go:Check)
   - 验证输入文件（.txt后缀）
   - 从文件名提取书名作者
   - 处理封面（本地/orly）
   - 编译正则表达式

3. **文本解析** (core/parser.go:Parse)
   - 检测文件编码
   - 逐行读取并识别章节
   - 构建Section列表
   - 组织卷-章节层次结构

4. **格式转换** (converter/dispatcher.go:Convert)
   - 根据format参数选择转换器
   - 调用各格式转换器的Build方法

5. **EPUB生成** (converter/epub.go:Build)
   - 创建EPUB实例
   - 生成CSS（默认+自定义+扩展）
   - 添加封面
   - 遍历章节生成HTML
   - 处理页眉图片（新增）
   - 写入文件

### 3.2 CSS生成流程

```
默认CSS → 字体CSS → 自定义CSS文件 → 内联扩展CSS → CSS变量
```

## 4. 扩展点

### 4.1 新增转换器
要实现新的电子书格式支持：

1. 在`internal/converter/`创建新文件
2. 实现`Converter`接口
3. 在`dispatcher.go`中添加调度逻辑

```go
type NewConverter struct{}

func NewNewConverter() *NewConverter {
    return &NewConverter{}
}

func (c NewConverter) Build(book model.Book) error {
    // 实现转换逻辑
    return nil
}
```

### 4.2 新增Book字段
要添加新的配置选项：

1. 在`model/book.go`的`Book`结构体添加字段
2. 在`cmd/cli/main.go`添加flag绑定
3. 在相应的转换器中使用新字段
4. 更新`docs/features.md`文档

### 4.3 新增CSS样式
要扩展CSS支持：

1. 在`epub.go`的默认CSS中添加新类
2. 在Book结构体中添加样式配置字段
3. 在`Build()`方法中应用样式

## 5. 依赖关系

### 5.1 外部依赖
- `github.com/go-shiori/go-epub`: EPUB生成
- `github.com/766b/mobi`: MOBI生成
- `github.com/leotaku/mobi`: AZW3生成
- `golang.org/x/text`: 编码处理
- `golang.org/x/net/html`: HTML处理

### 5.2 内部依赖
```
cmd/cli → internal/core → internal/model
              ↓
      internal/converter → internal/utils
              ↓
      internal/model
```

## 6. 配置优先级

### 6.1 参数优先级（从高到低）
1. 命令行参数
2. 环境变量
3. 默认值

### 6.2 CSS优先级（从高到低）
1. 用户自定义CSS文件（custom-css-file）
2. 内联扩展CSS（extended-css）
3. 字体相关CSS
4. 默认CSS

## 7. 错误处理

### 7.1 错误类型
- `ErrInvalidFile`: 无效输入文件
- `ErrMissingConfig`: 缺少必要配置

### 7.2 错误处理策略
- 预检查阶段返回错误
- 转换阶段记录并继续（尽可能生成可用输出）
- 致命错误立即退出

## 8. 性能考虑

### 8.1 优化点
- 流式读取大文件
- 正则表达式预编译
- 临时文件自动清理

### 8.2 大文件处理
- AZW3自动分卷（每2000章一个文件）
- 避免一次性加载整个文件到内存
