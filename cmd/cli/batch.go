package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/feewg/kaf-cli/internal/converter"
	"github.com/feewg/kaf-cli/internal/core"
	"github.com/feewg/kaf-cli/internal/model"
)

// BatchBookInfo 包含批量处理时的书籍信息和资源路径
type BatchBookInfo struct {
	Book         *model.Book
	CoverPath    string
	HeaderPath   string
	HeaderFolder string
}

// scanBooks 扫描文件夹获取所有书籍
// 支持两种模式：
// 1. 单文件夹模式：所有txt在一个文件夹，资源使用统一命名（cover.jpg, header.png等）
// 2. 子文件夹模式：每个txt在独立子文件夹，子文件夹内有各自的资源
func scanBooks(folder, outputFolder string) []BatchBookInfo {
	var books []BatchBookInfo

	entries, err := os.ReadDir(folder)
	if err != nil {
		fmt.Println("读取文件夹失败:", err)
		return books
	}

	// 首先检查是否是"单文件夹多书籍"模式
	// 查找通用的资源文件（cover.jpg, header.png等）
	globalResources := findGlobalResources(folder)

	// 处理所有txt文件和子文件夹
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否是子文件夹模式（子文件夹内有txt文件）
			subBooks := scanSubFolder(filepath.Join(folder, entry.Name()), outputFolder, globalResources)
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
			fmt.Printf("创建书籍失败 %s: %s\n", name, err)
			continue
		}

		// 设置输出路径
		if outputFolder != folder && outputFolder != "" {
			book.Out = filepath.Join(outputFolder, book.Bookname)
		}

		// 为单文件夹模式的书籍查找资源
		info := BatchBookInfo{Book: book}
		info = findResourcesForBook(info, folder, book.Bookname, globalResources)

		// 应用资源路径
		applyBookResources(book, info)

		books = append(books, info)
	}

	return books
}

// findGlobalResources 查找文件夹中的通用资源文件
func findGlobalResources(folder string) map[string]string {
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
func scanSubFolder(subFolder, outputFolder string, parentGlobalResources map[string]string) []BatchBookInfo {
	var books []BatchBookInfo

	entries, err := os.ReadDir(subFolder)
	if err != nil {
		return books
	}

	// 查找子文件夹内的通用资源
	localResources := findGlobalResources(subFolder)

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

		info := BatchBookInfo{
			Book:         book,
			CoverPath:    localResources["cover"],
			HeaderPath:   localResources["header"],
			HeaderFolder: headerFolder,
		}

		applyBookResources(book, info)
		books = append(books, info)
	}

	return books
}

// findResourcesForBook 为书籍查找资源（支持书名相关和通用资源）
func findResourcesForBook(info BatchBookInfo, folder, bookName string, globalResources map[string]string) BatchBookInfo {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return info
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否是页眉图片文件夹
			dirName := strings.ToLower(entry.Name())
			if strings.Contains(dirName, "header") || strings.Contains(dirName, "页眉") {
				// 检查是否与书名匹配
				if isResourceForBook(entry.Name(), bookName) {
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
			if isResourceForBook(base, bookName) {
				info.CoverPath = filepath.Join(folder, name)
			}
		}

		// 检查是否是页眉图片
		if strings.Contains(lowerBase, "header") || strings.Contains(lowerBase, "页眉") {
			if isResourceForBook(base, bookName) {
				info.HeaderPath = filepath.Join(folder, name)
			}
		}
	}

	// 如果没找到书名相关资源，使用通用资源
	if info.CoverPath == "" {
		info.CoverPath = globalResources["cover"]
	}
	if info.HeaderPath == "" {
		info.HeaderPath = globalResources["header"]
	}

	// 查找通用命名的页眉文件夹
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
func isResourceForBook(resourceName, bookName string) bool {
	cleanResource := cleanBookNameForMatch(resourceName)
	cleanBook := cleanBookNameForMatch(bookName)

	// 精确匹配
	if cleanResource == cleanBook {
		return true
	}

	// 资源名包含书名
	if strings.Contains(cleanResource, cleanBook) {
		return true
	}

	// 书名包含资源名（去除cover/header等后缀后）
	resourceWithoutSuffix := removeResourceSuffix(cleanResource)
	if resourceWithoutSuffix != "" && strings.Contains(cleanBook, resourceWithoutSuffix) {
		return true
	}

	return false
}

// cleanBookNameForMatch 清理书名用于匹配
func cleanBookNameForMatch(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "《", "")
	name = strings.ReplaceAll(name, "》", "")
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	return name
}

// removeResourceSuffix 移除资源文件的后缀（cover, header等）
func removeResourceSuffix(name string) string {
	suffixes := []string{"cover", "封面", "header", "页眉", "chapterheader", "chapter_header"}
	for _, suffix := range suffixes {
		name = strings.ReplaceAll(name, suffix, "")
	}
	return strings.TrimSpace(name)
}

// applyBookResources 应用资源到书籍
func applyBookResources(book *model.Book, info BatchBookInfo) {
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
