package foundationrender

import (
	"bytes"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
)

// 有些题目的表格非常复杂，不太好使用table插件转为markdown，因此直接保留原始html表格以获得更好的兼容性
type tablePlugin struct{}

func NewTablePlugin() converter.Plugin {
	return &tablePlugin{}
}

func (b *tablePlugin) Name() string {
	return "table_preserve"
}

func (b *tablePlugin) Init(conv *converter.Converter) error {

	// 声明 table 为块级，避免 markdown 把它合并进段落
	conv.Register.TagType("table", converter.TagTypeBlock, converter.PriorityStandard)

	// 自定义渲染器：直接输出原始 HTML，不做 Markdown 转换
	conv.Register.Renderer(
		func(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
			if n.Type == html.ElementNode && n.Data == "table" {
				htmlStr := renderNodeAsHTML(n)
				_, _ = w.WriteString(htmlStr)
				return converter.RenderSuccess
			}
			return converter.RenderTryNext
		},
		converter.PriorityStandard, // 高优先，确保优先于内置表格插件
	)

	return nil
}

// 将节点序列化为完整 HTML 字符串
func renderNodeAsHTML(n *html.Node) string {
	var buf bytes.Buffer
	_ = html.Render(&buf, n)
	return buf.String()
}
