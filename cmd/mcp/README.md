# KAF MCP Server

基于 Model Context Protocol (MCP) 的电子书格式转换服务，支持将 TXT 小说文件转换为 EPUB、MOBI、AZW3 等电子书格式。

## 功能特性

- **完整的 MCP 协议支持**: 提供丰富的工具集供 AI 助手调用
- **多格式转换**: 支持 EPUB、MOBI、AZW3 和批量格式输出
- **批量处理**: 支持同时转换多个文件
- **任务管理**: 支持异步任务和进度查询
- **配置预设**: 保存和加载常用转换配置
- **文件资源管理**: 读取和列出工作目录文件
- **跨平台**: 支持 Windows、macOS、Linux

## 安装

### 自动安装 (推荐)

```bash
curl -sSL https://github.com/feewg/kaf-cli/releases/download/mcp-1.0.0/install-mcp.sh | bash -s 1.0.0
```

### 手动下载

从 [Releases](https://github.com/feewg/kaf-cli/releases) 页面下载适合您系统的版本：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Windows | x64 | `kaf-mcp_1.0.0_windows_amd64.exe` |
| Windows | x86 | `kaf-mcp_1.0.0_windows_386.exe` |
| macOS | Intel | `kaf-mcp_1.0.0_darwin_amd64` |
| macOS | Apple Silicon | `kaf-mcp_1.0.0_darwin_arm64` |
| Linux | x64 | `kaf-mcp_1.0.0_linux_amd64` |
| Linux | ARM64 | `kaf-mcp_1.0.0_linux_arm64` |

下载后重命名为 `kaf-mcp`（或 Windows 下 `kaf-mcp.exe`），并添加到系统 PATH。

## 配置

### Cherry Studio / Claude Desktop

编辑配置文件：

**macOS**: `~/Library/Application Support/CherryStudio/config/mcp.json`

**Windows**: `%APPDATA%/CherryStudio/config/mcp.json`

```json
{
  "mcpServers": {
    "kaf-mcp": {
      "command": "kaf-mcp",
      "env": {
        "KAF_DIR": "/path/to/your/books",
        "LOGGER": "false"
      }
    }
  }
}
```

### 环境变量

- `KAF_DIR`: 设置工作目录，TXT 文件默认从此目录读取，输出文件也保存到此目录
- `LOGGER`: 设置为 `true` 启用日志记录到文件（默认 `false`，日志输出到 stderr）

## MCP 工具列表

### 1. kaf_convert - 单个文件转换

转换单个 TXT 文件为电子书格式。

**参数：**
- `filename` (必填): TXT 文件路径
- `bookname`: 书名
- `author`: 作者
- `format`: 输出格式 (`all`, `epub`, `mobi`, `azw3`)
- `out`: 输出文件名
- `match`: 章节匹配正则表达式
- `volume_match`: 卷匹配规则
- `exclude`: 排除无效章节的正则
- `cover`: 封面图片路径或 `orly` 或 `none`
- `cover_orly_color`: orly 封面颜色
- `cover_orly_idx`: orly 封面动物图案索引
- `font`: 嵌入字体文件路径
- `max`: 标题最大字数
- `indent`: 段落缩进字数
- `align`: 标题对齐方式
- `bottom`: 段落间距
- `line_height`: 行高
- `lang`: 语言设置
- `tips`: 是否添加教程
- `separate_chapter_number`: 是否分离章节序号
- `custom_css_file`: 自定义 CSS 文件路径

**新增参数（v1.x）：**
- `extended_css`: 内联扩展 CSS 样式代码
- `css_variables`: CSS 变量定义（格式：`--var1:value1;--var2:value2`）
- `chapter_header_image`: 章节页眉图片路径（所有章节相同）
- `chapter_header_image_folder`: 章节页眉图片文件夹（按章节名匹配）
- `chapter_header_image_position`: 页眉图片位置 (`left`, `center`, `right`)
- `chapter_header_image_height`: 页眉图片高度（如 `100px`, `2em`）
- `chapter_header_image_width`: 页眉图片宽度（如 `50%`, `200px`）
- `chapter_header_image_mode`: 图片模式 (`single`, `folder`)

### 2. kaf_batch_convert - 文件夹批量转换（新增）

批量转换文件夹中的 TXT 小说文件，自动识别配套资源（封面、页眉图片等）。

**参数：**
- `folder` (必填): 小说文件夹路径
- `format`: 输出格式
- `output_folder`: 输出文件夹（默认使用输入文件夹）
- `font`: 嵌入字体文件路径（全局）
- `indent`: 段落缩进字数（全局）
- `align`: 标题对齐方式（全局）
- `bottom`: 段落间距（全局）
- `line_height`: 行高（全局）
- `lang`: 语言设置（全局）
- `separate_chapter_number`: 是否分离章节序号（全局）
- `custom_css_file`: 自定义 CSS 文件路径（全局）
- `extended_css`: 内联扩展 CSS 样式（全局）
- `css_variables`: CSS 变量定义（全局）

**文件夹组织结构：**

支持两种组织方式：

**方式一：单文件夹模式（推荐批量处理）**
```
novels/
├── cover.jpg              # 通用封面（所有书籍使用）
├── header.png             # 通用页眉（所有书籍使用）
├── 斗破苍穹.txt
├── 武动乾坤.txt
└── ...
```

**方式二：子文件夹模式（每本书独立）**
```
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
```

**资源文件命名规则：**
- **通用封面**：`cover.jpg`, `cover.png`, `封面.jpg` 等
- **通用页眉**：`header.jpg`, `header.png`, `页眉.jpg` 等
- **页眉文件夹**：`headers/`, `header/`, `页眉/`
- **书名相关命名**：`《书名》封面.jpg`, `《书名》页眉.png` 等（优先级高于通用命名）

**文件名前缀处理：**
- 支持去除文件名前缀，格式为 `前缀@实际文件名.txt`
- 例：`soushu2024@《斗破苍穹》作者：天蚕土豆.txt` 会识别为 `《斗破苍穹》作者：天蚕土豆.txt`
- 适用于从某些网站下载的文件带有来源前缀的情况

**自动匹配优先级：**
1. 子文件夹内资源优先于父文件夹通用资源
2. 书名相关命名优先于通用命名
3. 精确匹配优先于模糊匹配

### 3. kaf_get_job - 获取任务状态

查询转换任务的详细状态和结果。

**参数：**
- `job_id` (必填): 任务 ID

### 4. kaf_list_jobs - 列出任务列表

列出所有转换任务，支持状态过滤。

**参数：**
- `status`: 按状态过滤 (`pending`, `processing`, `completed`, `failed`)
- `limit`: 返回最大数量

### 5. kaf_read_file - 读取文件

读取转换后的电子书文件内容（用于验证或预览）。

**参数：**
- `file_path` (必填): 文件路径
- `as_base64`: 是否以 base64 编码返回

### 6. kaf_list_files - 列出文件

列出工作目录中的文件。

**参数：**
- `directory`: 目录路径，默认为工作目录
- `pattern`: 文件匹配模式，如 `*.txt`, `*.epub`

### 7. kaf_save_preset - 保存配置预设

保存常用的转换配置为预设。

**参数：**
- `name` (必填): 预设名称
- `config` (必填): 配置对象，包含转换参数

### 8. kaf_load_preset - 加载配置预设

加载保存的配置预设。

**参数：**
- `name` (必填): 预设名称

### 9. kaf_list_presets - 列出预设

列出所有保存的配置预设。

### 10. kaf_system_info - 系统信息

获取 KAF 转换器的系统信息和版本。

## 使用示例

### 基本转换

```
使用 kaf_convert 工具转换小说：
- filename: /path/to/novel.txt
- bookname: 我的小说
- author: 作者名
- format: epub
```

### 批量转换

```
使用 kaf_batch_convert 工具批量转换：
- folder: /path/to/novels
- format: epub
- align: center
- separate_chapter_number: true
```

### 使用章节页眉图片

```
使用 kaf_convert 工具转换并添加页眉图片：
- filename: /path/to/novel.txt
- bookname: 我的小说
- chapter_header_image: /path/to/header.png
- chapter_header_image_position: center
- chapter_header_image_height: 150px
```

或使用文件夹模式匹配各章节不同图片：

```
使用 kaf_convert 工具转换并添加章节页眉：
- filename: /path/to/novel.txt
- bookname: 我的小说
- chapter_header_image_folder: /path/to/headers/
- chapter_header_image_mode: folder
- chapter_header_image_position: center
```

### 使用配置预设

```
先使用 kaf_save_preset 保存预设：
- name: 我的配置
- config:
    format: epub
    author: 我的笔名
    cover: orly
    cover_orly_color: "#FF5733"

然后使用 kaf_convert 时直接引用预设。或者先使用 kaf_load_preset 加载配置。
```

## 故障排除

### MCP 连接失败

1. 确认 `kaf-mcp` 在系统 PATH 中：
   ```bash
   which kaf-mcp  # macOS/Linux
   where kaf-mcp  # Windows
   ```

2. 检查配置文件路径是否正确

3. 查看日志：设置 `LOGGER=true` 环境变量，日志会保存到用户主目录的 `kaf-mcp.log`

### 文件找不到

- 确认 `KAF_DIR` 环境变量设置正确
- 使用绝对路径而非相对路径
- 检查文件权限

### 转换失败

- 确认输入文件是有效的 UTF-8 或 GBK 编码的 TXT 文件
- 检查磁盘空间是否充足
- 查看任务状态获取详细错误信息

## 许可证

MIT License - 详见 [LICENSE](../../LICENSE)
