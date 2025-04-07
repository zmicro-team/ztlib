package md

import (
	"fmt"
	"strings"
)

// MarkdownBuilder 用于逐步构建 Markdown 字符串。
type MarkdownBuilder struct {
	builder strings.Builder
}

// NewMarkdownBuilder 创建一个新的 MarkdownBuilder 实例。
func NewMarkdownBuilder() *MarkdownBuilder {
	return &MarkdownBuilder{}
}

// String 返回累积的 Markdown 字符串。
func (mb *MarkdownBuilder) String() string {
	return mb.builder.String()
}

// Reset 清除底层的 builder，允许复用。
func (mb *MarkdownBuilder) Reset() {
	mb.builder.Reset()
}

// WriteString 向 builder 追加一个原始字符串。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) WriteString(s string) *MarkdownBuilder {
	mb.builder.WriteString(s)
	return mb
}

// Newline 添加一个换行符。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) Newline() *MarkdownBuilder {
	mb.builder.WriteString("\n")
	return mb
}

// Paragraph 添加一个字符串，后跟两个换行符（标准的段落分隔）。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) Paragraph(s string) *MarkdownBuilder {
	mb.builder.WriteString(s)
	mb.builder.WriteString("\n\n")
	return mb
}

// Heading 添加一个 Markdown 标题（例如 "# 标题 1", "## 标题 2"）。
// 级别 `lv` 会被限制在 1 到 6 之间。
// 自动在标题后添加一个换行符。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) Heading(lv int, s string) *MarkdownBuilder {
	if lv <= 0 {
		lv = 1
	}
	if lv > 6 {
		lv = 6
	}
	// 使用 Fprintf 直接写入 builder 以提高效率
	fmt.Fprintf(&mb.builder, "%s %s\n", strings.Repeat("#", lv), s)
	return mb
}

// Bold 添加粗体文本（例如 "**text**"）。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) Bold(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "**%s**", s)
	return mb
}

// Italic 添加斜体文本（例如 "*text*"）。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) Italic(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "*%s*", s)
	return mb
}

// Link 添加一个 Markdown 链接（例如 "[text](Url)"）。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) Link(text, url string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "[%s](%s)", text, url)
	return mb
}

// QuoteText 添加一个块引用行（例如 "> text\n"）。
// 自动在引用的文本后添加一个换行符。
// 对于多行引用，为每行调用此方法，或使用 WriteString 进行手动格式化。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) QuoteText(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "> %s\n", s)
	return mb
}

// QuoteCode 添加内联代码格式（例如 "`code`"）。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) QuoteCode(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "`%s`", s)
	return mb
}

// CodeBlock 添加一个围栏代码块。
// `language` 是可选的（可以为空）。
// `code` 是代码块的内容。
// 自动添加换行符以确保正确格式化。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) CodeBlock(language, code string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "```%s\n%s\n```\n", language, code)
	return mb
}

// UnorderedListItem 添加一个无序列表项（例如 "- item\n"）。
// 不会自动处理嵌套列表的缩进。
// 自动添加一个换行符。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) UnorderedListItem(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "- %s\n", s)
	return mb
}

// OrderedListItem 添加一个有序列表项，使用 "1."（与 CommonMark 兼容）。
// 不会自动处理缩进或 "1." 之外的自动编号。
// 自动添加一个换行符。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) OrderedListItem(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, "1. %s\n", s)
	return mb
}

// HorizontalRule 添加一条水平分割线 ("---\n")。
// 自动添加一个换行符。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) HorizontalRule() *MarkdownBuilder {
	mb.builder.WriteString("---\n")
	return mb
}

// --- 特定平台 / HTML ---
// 注意：使用像 <font> 这样的 HTML 可能无法在所有 Markdown 渲染器中正常工作，
// 并且 <font> 标签本身在 HTML5 中已被弃用。如果可能，请考虑替代方案。

// InfoText 添加被 <font color="info"> 标签包裹的文本。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) InfoText(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, `<font color="info">%s</font>`, s)
	return mb
}

// CommentText 添加被 <font color="comment"> 标签包裹的文本。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) CommentText(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, `<font color="comment">%s</font>`, s)
	return mb
}

// WarningText 添加被 <font color="warning"> 标签包裹的文本。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) WarningText(s string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, `<font color="warning">%s</font>`, s)
	return mb
}

// MentionByUserid 添加特定平台的用户提及（例如 "<@userid>"）。
// 返回 builder 以支持链式调用。
func (mb *MarkdownBuilder) MentionByUserid(userid string) *MarkdownBuilder {
	fmt.Fprintf(&mb.builder, `<@%s>`, userid)
	return mb
}

// 示例用法（可以在不同的文件/包或 _test.go 文件中）：
/*
func main() {
    builder := md.NewMarkdownBuilder()

    builder.Heading(1, "文档标题")
    builder.Paragraph("这是一个介绍性段落。")

    builder.Heading(2, "特性")
    builder.UnorderedListItem("使用 strings.Builder 提高效率。")
    builder.UnorderedListItem("支持链式调用。")
    builder.UnorderedListItem("包含常见的 Markdown 元素。")
    builder.Newline() // 添加额外的换行符以增加间距

    builder.WriteString("这是一些粗体文本: ").Bold("重要!").WriteString(" 以及一个链接: ")
    builder.Link("Go 语言官网", "https://golang.org").WriteString(".")
    builder.Newline().Newline() // 等同于 Paragraph("")

    builder.QuoteText("这是一行引用。")
    builder.QuoteText("如果多次调用，可以跨越多行。")
    builder.Newline()

    builder.WriteString("内联代码看起来像这样 ").QuoteCode("var x = 10").WriteString(".")
    builder.Newline().Newline()

    builder.CodeBlock("go", `package main

import "fmt"

func main() {
    fmt.Println("你好，代码块！")
}`)

    builder.Heading(3, "特定平台功能")
    builder.MentionByUserid("U12345").WriteString(": 请检查这个。 ")
    builder.InfoText("这是一条提示信息。")
    builder.Newline()

    fmt.Println(builder.String())

    // --- 重置并复用 ---
    builder.Reset()
    builder.Heading(2, "重置后的另一个部分")
    builder.Paragraph("Builder 现在是空的。")
    fmt.Println("\n--- 重置后 ---")
    fmt.Println(builder.String())
}
*/
