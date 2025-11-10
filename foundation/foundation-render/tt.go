package foundationrender

import (
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
)

type ttPlugin struct{}

func NewTTPlugin() converter.Plugin {
	return &ttPlugin{}
}

func (b *ttPlugin) Name() string {
	return "tt"
}

func (b *ttPlugin) Init(conv *converter.Converter) error {
	conv.Register.TagType("tt", converter.TagTypeInline, converter.PriorityStandard)

	conv.Register.Renderer(
		func(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
			if n.Type != html.ElementNode || n.Data != "tt" {
				return converter.RenderTryNext
			}

			var builder strings.Builder
			isMultiline := false

			// 遍历子节点
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					builder.WriteString(c.Data)
				} else if c.Type == html.ElementNode && c.Data == "br" {
					isMultiline = true
					builder.WriteString("\n")
				} else {
					// 对其他元素直接忽略或可以递归
				}
			}

			content := builder.String()
			content = strings.TrimSpace(content) // 去掉首尾多余空格

			if isMultiline {
				_, err := w.WriteString("```\n")
				if err != nil {
					return converter.RenderTryNext
				}
				_, err := w.WriteString(content)
				if err != nil {
					return converter.RenderTryNext
				}
				_, err := w.WriteString("\n```")
				if err != nil {
					return converter.RenderTryNext
				}
			} else {
				w.WriteString("`" + content + "`")
			}

			return converter.RenderSuccess
		}, converter.PriorityStandard,
	)

	return nil
}
