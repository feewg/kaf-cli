# kaf-cli 功能文档

本文档详细描述了 kaf-cli 的所有功能，供开发者和用户参考。

## 1. 基础功能

### 1.1 文件转换
- **核心功能**: 将TXT文本文件转换为EPUB、MOBI、AZW3电子书格式
- **批量处理**: 支持多种输出格式同时生成
- **转换速度**: EPUB格式生成速度可达300章/秒以上

### 1.2 智能识别
- **编码识别**: 自动识别文件编码（UTF-8、GBK、GB18030等），解决中文乱码
- **章节识别**: 自动识别章节标题，支持多种格式
- **书名作者识别**: 从文件名自动提取书名和作者
  - 支持格式: `《书名》（校对版全本）作者：作者名.txt`
  - 支持格式: `《书名》作者：作者名.txt`

### 1.3 封面处理
- **自定义封面**: 支持本地图片作为封面（PNG、JPG格式）
- **Orly封面**: 在线生成Orly风格封面
  - 支持自定义主题色（1-16或十六进制颜色）
  - 支持自定义动物图案（0-41）
- **默认封面**: 自动查找目录下的`cover.png`文件

## 2. 章节处理

### 2.1 章节识别规则
- **默认规则**: 自动匹配常见章节格式
  - 第X章/回/节/集
  - 中文数字章节（第一章、卷一）
  - 阿拉伯数字章节（1.、1、）
  - 特殊章节（引子、楔子、序章、最终章、番外、完本感言）
  - 英文章节（Section、Chapter、Page）
- **自定义规则**: 支持正则表达式自定义章节匹配

### 2.2 卷识别
- **默认规则**: 匹配"第X卷/部"格式
- **禁用卷**: 可设置`volume-match=false`禁用卷识别
- **卷样式**: 使用双线边框精美样式

### 2.3 章节样式
- **标题对齐**: 支持左对齐、居中、右对齐
- **标题分离**: 支持将章节序号和标题分离显示
- **最大字数**: 可限制标题最大字数（默认35字）
- **未知章节**: 可设置未识别内容的默认章节名

## 3. 排版功能

### 3.1 段落样式
- **自动识别**: 智能识别段落
- **首行缩进**: 可自定义缩进字数（默认2字）
- **段落间距**: 可自定义段间距（默认1em）
- **行高**: 可自定义行间距（默认1.5rem）

### 3.2 HTML标签处理
- **智能转义**: 自动转义不支持的HTML标签
- **保留标签**: 保留EPUB支持的标签（img、br、p、span等）
- **安全性**: 防止XSS攻击

### 3.3 字体支持
- **字体嵌入**: 支持嵌入自定义字体文件
- **字体应用**: 嵌入后正文自动使用该字体

## 4. CSS样式系统

### 4.1 基础CSS类
- `h2.volume` - 卷名样式
- `h3.title` - 章节标题样式
- `h3.title span.chapter-number` - 章节序号样式
- `.content` - 正文段落样式
- `body` - 整体样式
- `.chapter-header-image` - 章节页眉图片样式（新增）

### 4.2 自定义CSS文件
- **参数**: `--custom-css-file`
- **功能**: 通过外部CSS文件覆盖默认样式
- **优先级**: 用户CSS > 默认CSS

### 4.3 扩展CSS功能（新增）
- **内联CSS**: `--extended-css` 直接传入CSS代码
- **CSS变量**: `--css-variables` 定义CSS变量
  - 格式: `--var1:value1;--var2:value2`

### 4.4 可用CSS变量
```css
:root {
  /* 可由用户自定义 */
}
```

## 5. 章节页眉图片（新增功能）

### 5.1 功能概述
支持在每个章节的页眉添加图片，可用于装饰章节开头或添加章节特定的插图。

### 5.2 使用模式

#### 单图片模式
- **参数**: `--chapter-header-image`
- **说明**: 所有章节使用同一张图片
- **示例**: `--chapter-header-image header.png`

#### 文件夹匹配模式
- **参数**: `--chapter-header-image-folder`
- **说明**: 按章节名从文件夹中匹配图片
- **匹配规则**:
  1. 完整章节名匹配（如"第一章.png"）
  2. 数字匹配（如章节"第一章"匹配"1.png"）
- **支持格式**: PNG、JPG、JPEG、GIF、WEBP

### 5.3 图片样式参数
- `--chapter-header-image-position`: 图片位置
  - `left` - 左对齐
  - `center` - 居中（默认）
  - `right` - 右对齐
- `--chapter-header-image-height`: 图片高度（如`100px`、`2em`）
- `--chapter-header-image-width`: 图片宽度（如`50%`、`200px`，默认`100%`）
- `--chapter-header-image-mode`: 图片模式
  - `single` - 所有章节相同（默认）
  - `folder` - 按章节名匹配

### 5.4 完整示例
```bash
# 所有章节使用同一张图片
kaf-cli -filename novel.txt --chapter-header-image header.png --chapter-header-image-position center

# 按章节名匹配图片
kaf-cli -filename novel.txt --chapter-header-image-folder ./headers/ --chapter-header-image-mode folder

# 自定义图片大小
kaf-cli -filename novel.txt --chapter-header-image header.png --chapter-header-image-height 150px --chapter-header-image-width 80%
```

## 6. 输出控制

### 6.1 输出格式
- `all` - 生成所有格式（默认）
- `epub` - 仅生成EPUB
- `mobi` - 生成EPUB和MOBI
- `azw3` - 仅生成AZW3

### 6.2 输出文件名
- **默认**: 使用书名作为文件名
- **自定义**: `--out` 参数指定输出文件名（不含扩展名）

### 6.3 语言设置
- **支持语言**: en, de, fr, it, es, zh, ja, pt, ru, nl
- **默认**: zh（中文）
- **环境变量**: `KAF_CLI_LANG`

## 7. 高级功能

### 7.1 排除规则
- **功能**: 排除无效章节/卷
- **默认规则**: 排除"部门、部队"等误识别内容
- **自定义**: `--exclude` 参数设置正则表达式

### 7.2 教程文本
- **功能**: 在书籍末尾添加制作教程
- **默认**: 开启
- **关闭**: `--tips=false`

### 7.3 拖放模式
- **Windows**: 将TXT文件拖到kaf-cli.exe上自动转换
- **自动封面**: 自动使用目录下的cover.png作为封面

## 8. 环境变量

| 变量名 | 功能 | 默认值 |
|--------|------|--------|
| `KAF_CLI_ALIGN` | 标题对齐方式 | center |
| `KAF_CLI_LANG` | 书籍语言 | zh |
| `KAF_CLI_FORMAT` | 输出格式 | all |

## 9. MCP支持

### 9.1 MCP服务器
- **功能**: 提供MCP协议支持，可与AI助手集成
- **入口**: `cmd/mcp/main.go`
- **配置**: 支持Cherry Studio等MCP客户端

### 9.2 可用工具
- 电子书转换（`kaf_convert`）
- 文件夹批量转换（`kaf_batch_convert`）- 新增
- 参数查询
- 配置管理

### 9.3 文件夹批量转换规范

MCP版本新增文件夹批量转换功能，支持按规范命名自动识别配套资源。

#### 文件命名规范
```
小说文件夹/
├── 《书名1》作者：作者1.txt          # 小说文件
├── 《书名1》封面.png                # 封面图片（自动匹配）
├── 《书名1》页眉.png                # 页眉图片（单图片模式）
├── 《书名1》页眉/                   # 页眉图片文件夹（多图片模式）
│   ├── 第一章.png
│   ├── 第二章.png
│   └── ...
├── 《书名2》作者：作者2.txt
├── 《书名2》封面.jpg
└── ...
```

#### 自动匹配规则
1. **书名提取**：支持 `《书名》作者：作者名.txt` 格式
2. **封面匹配**：查找 `书名+封面/cover` 图片
3. **页眉匹配**：查找 `书名+页眉/header` 图片或文件夹
4. **模糊匹配**：支持部分匹配，忽略大小写和特殊字符

## 10. 参数汇总

### 基础参数
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-filename` | TXT文件名 | 必填 |
| `-bookname` | 书名 | 自动识别 |
| `-author` | 作者 | YSTYLE |
| `-cover` | 封面图片 | cover.png |
| `-format` | 输出格式 | all |
| `-out` | 输出文件名 | 书名 |

### 章节参数
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-match` | 章节匹配正则 | 自动 |
| `-volume-match` | 卷匹配正则 | 内置规则 |
| `-exclude` | 排除规则 | 内置规则 |
| `-max` | 标题最大字数 | 35 |
| `-unknow-title` | 未知章节名 | 章节正文 |

### 样式参数
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-align` | 标题对齐 | center |
| `-indent` | 段落缩进 | 2 |
| `-bottom` | 段落间距 | 1em |
| `-line-height` | 行高 | 1.5rem |
| `-font` | 嵌入字体 | - |
| `-separate-chapter-number` | 分离序号 | false |
| `-custom-css-file` | 自定义CSS文件 | - |

### 新增CSS参数（v1.x）
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-extended-css` | 内联CSS代码 | - |
| `-css-variables` | CSS变量定义 | - |

### 新增页眉图片参数（v1.x）
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-chapter-header-image` | 页眉图片路径 | - |
| `-chapter-header-image-folder` | 图片文件夹 | - |
| `-chapter-header-image-position` | 图片位置 | center |
| `-chapter-header-image-height` | 图片高度 | auto |
| `-chapter-header-image-width` | 图片宽度 | 100% |
| `-chapter-header-image-mode` | 匹配模式 | single |

## 11. 版本历史

### v1.0.0
- 基础TXT转EPUB/MOBI/AZW3功能
- 自动章节识别
- 自定义CSS支持

### v1.x.0（当前）
- 新增扩展CSS支持（`--extended-css`、`--css-variables`）
- 新增章节页眉图片支持
- 完善文档系统
