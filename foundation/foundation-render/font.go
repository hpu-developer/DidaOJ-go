package foundationrender

import (
	"fmt"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
)

type fontColorPlugin struct{}

func NewFontColorPlugin() converter.Plugin {
	return &fontColorPlugin{}
}

func (b *fontColorPlugin) Name() string {
	return "fontColor"
}

func (b *fontColorPlugin) Init(conv *converter.Converter) error {
	conv.Register.TagType("font", converter.TagTypeInline, converter.PriorityStandard)

	conv.Register.Renderer(
		func(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
			if n.Type != html.ElementNode || n.Data != "font" {
				return converter.RenderTryNext
			}

			// 获取 color 属性
			color := ""
			for _, attr := range n.Attr {
				if attr.Key == "color" {
					color = attr.Val
					break
				}
			}

			// 递归收集文本内容
			content := extractText(n)

			// 输出自定义 Markdown
			if color != "" {
				_, err := w.WriteString(fmt.Sprintf("<span style=\"color:%s\">%s</span>", color, content))
				if err != nil {
					return converter.RenderTryNext
				}
			} else {
				_, err := w.WriteString(content)
				if err != nil {
					return converter.RenderTryNext
				}
			}

			return converter.RenderSuccess
		}, converter.PriorityStandard,
	)

	return nil
}

// 提取节点及子节点的文本
func extractText(n *html.Node) string {
	text := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode {
			text += extractText(c)
		}
	}
	return text
}
