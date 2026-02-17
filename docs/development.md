# kaf-cli 开发指南

本文档面向开发者，提供开发、调试和扩展 kaf-cli 的指南。

## 1. 开发环境

### 1.1 前置要求
- Go 1.21 或更高版本
- Git
- Make（可选）

### 1.2 项目克隆
```bash
git clone https://github.com/feewg/kaf-cli.git
cd kaf-cli
```

### 1.3 依赖安装
```bash
go mod download
```

## 2. 项目构建

### 2.1 构建CLI版本
```bash
# Windows
go build -o kaf-cli.exe ./cmd/cli

# Linux/Mac
go build -o kaf-cli ./cmd/cli
```

### 2.2 构建MCP版本
```bash
# Windows
go build -o kaf-mcp.exe ./cmd/mcp

# Linux/Mac
go build -o kaf-mcp ./cmd/mcp
```

### 2.3 交叉编译
```bash
# Windows → Linux
GOOS=linux GOARCH=amd64 go build -o kaf-cli-linux ./cmd/cli

# Windows → macOS
GOOS=darwin GOARCH=amd64 go build -o kaf-cli-darwin ./cmd/cli
```

## 3. 代码规范

### 3.1 命名规范
- 包名：小写，简短（`converter`, `parser`）
- 文件名：小写，下划线分隔（`epub.go`, `mobi_utils.go`）
- 结构体：大驼峰（`EpubConverter`, `Book`）
- 接口：名词或动词（`Converter`）
- 方法：大驼峰（`Build`, `Parse`）
- 私有：小驼峰（`wrapTitle`, `sanitizeHTML`）
- 常量：大写下划线（`DefaultMatchTips`）

### 3.2 注释规范
- 所有导出项必须有注释
- 注释以被注释项开头
- 函数注释说明功能和参数

```go
// EpubConverter 负责将Book转换为EPUB格式
type EpubConverter struct {
    // ...
}

// Build 生成EPUB电子书
// 根据book的配置生成epub文件
func (convert EpubConverter) Build(book model.Book) error {
    // ...
}
```

### 3.3 错误处理
- 使用`fmt.Errorf()`包装错误，添加上下文
- 预定义错误使用`errors.New()`
- 错误消息使用中文（面向中文用户）

```go
if err != nil {
    return fmt.Errorf("读取文件失败: %w", err)
}
```

## 4. 添加新功能

### 4.1 添加Book配置字段

步骤：
1. 在`internal/model/book.go`添加字段
2. 在`cmd/cli/main.go`添加flag绑定
3. 在转换器中使用
4. 更新文档

示例：
```go
// internal/model/book.go
type Book struct {
    // ... 现有字段
    NewFeature string // 新功能配置
}

// cmd/cli/main.go
flag.StringVar(&book.NewFeature, "new-feature", "", "新功能说明")

// internal/converter/epub.go
func (convert EpubConverter) Build(book model.Book) error {
    if book.NewFeature != "" {
        // 使用新功能
    }
}
```

### 4.2 添加CSS支持

在`epub.go`中扩展CSS：

```go
// 1. 在默认CSS中添加新类
CSSContent: `
    /* 现有样式 */
    
    /* 新样式 */
    .new-class {
        property: value;
    }
`,

// 2. 在Build方法中应用
if book.NewCSSField != "" {
    epubcss += fmt.Sprintf("\n.new-class { %s }\n", book.NewCSSField)
}
```

### 4.3 添加新的命令行参数

在`cmd/cli/main.go`中：

```go
// 找到flag.Parse()之前的位置
flag.StringVar(&book.NewField, "new-param", "default", "参数说明")
```

## 5. 调试技巧

### 5.1 使用Delve调试
```bash
# 安装Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试运行
dlv debug ./cmd/cli -- -filename test.txt

# 在main函数处设置断点
(dlv) break main.main
(dlv) continue
```

### 5.2 日志输出
```go
// 临时调试输出
fmt.Printf("Debug: book=%+v\n", book)

// 条件调试
if os.Getenv("KAF_DEBUG") != "" {
    log.Printf("Debug: %s", message)
}
```

### 5.3 单元测试
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/parser/

# 带覆盖率
go test -cover ./...
```

## 6. 新增文档

### 6.1 功能文档 (docs/features.md)
当添加新功能时，更新`features.md`：
- 在对应章节添加功能说明
- 添加参数表格
- 提供使用示例

### 6.2 架构文档 (docs/architecture.md)
当修改架构时，更新`architecture.md`：
- 更新模块设计
- 更新数据流图
- 更新依赖关系

### 6.3 README.md
更新主README：
- 在功能列表中添加新功能
- 在参数列表中添加新参数
- 添加使用示例

## 7. 发布流程

### 7.1 版本号管理
使用语义化版本：
- MAJOR: 不兼容的API更改
- MINOR: 向后兼容的功能添加
- PATCH: 向后兼容的问题修复

### 7.2 发布检查清单
- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] CHANGELOG已更新
- [ ] 版本号已更新
- [ ] 所有平台构建成功

### 7.3 构建发布包
```bash
# 清理
go clean

# 构建各平台版本
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/kaf-cli-windows.exe ./cmd/cli
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/kaf-cli-linux ./cmd/cli
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/kaf-cli-darwin ./cmd/cli
```

## 8. 常见问题

### 8.1 编码问题
**问题**: 中文字符显示乱码
**解决**: 确保文件使用UTF-8编码，或让程序自动检测编码

### 8.2 正则表达式
**问题**: 章节识别不准确
**解决**: 使用`-match`参数自定义正则，参考`internal/model/book.go`中的`DefaultMatchTips`

### 8.3 封面生成失败
**问题**: Orly封面生成失败
**解决**: 检查网络连接，或使用本地图片

## 9. 贡献指南

### 9.1 提交Issue
- 描述问题现象
- 提供复现步骤
- 提供环境信息（OS、版本）

### 9.2 提交PR
1. Fork项目
2. 创建特性分支 (`git checkout -b feature/xxx`)
3. 提交更改 (`git commit -am 'Add xxx'`)
4. 推送分支 (`git push origin feature/xxx`)
5. 创建Pull Request

### 9.3 代码审查标准
- 代码符合规范
- 有适当的测试
- 文档已更新
- 所有CI检查通过

## 10. 文件速查

| 文件 | 功能 | 修改场景 |
|------|------|----------|
| `model/book.go` | Book定义 | 新增配置项 |
| `cmd/cli/main.go` | CLI入口 | 新增命令行参数 |
| `converter/epub.go` | EPUB生成 | 修改EPUB输出 |
| `converter/azw3.go` | AZW3生成 | 修改AZW3输出 |
| `core/parser.go` | 文本解析 | 修改解析逻辑 |
| `docs/features.md` | 功能文档 | 新增功能 |
| `docs/architecture.md` | 架构文档 | 架构变更 |

## 11. 更新文档索引

### 11.1 文档阅读顺序
1. `README.md` - 快速开始
2. `docs/features.md` - 了解所有功能
3. `docs/architecture.md` - 了解项目架构
4. `docs/development.md` - 开发指南（本文档）

### 11.2 修改代码前必读
- 了解`Book`结构体的所有字段
- 了解转换流程
- 了解CSS生成逻辑
- 查看相关文档

### 11.3 快速定位
- **新增参数**: `cmd/cli/main.go` + `model/book.go`
- **新增样式**: `converter/epub.go` CSS部分
- **修改解析**: `core/parser.go`
- **新增格式**: `converter/` + `dispatcher.go`
