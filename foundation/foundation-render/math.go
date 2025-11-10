package foundationrender

import (
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
)

type mathPlugin struct{}

func NewMathPlugin() converter.Plugin {
	return &mathPlugin{}
}

func (b *mathPlugin) Name() string {
	return "math"
}

func (b *mathPlugin) Init(conv *converter.Converter) error {

	conv.Register.Renderer(
		func(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
			if n.Type == html.ElementNode && strings.ToLower(n.Data) == "sup" {
				_, _ = w.WriteString("<sup>")
				_, _ = w.WriteString(extractRawText(n))
				_, _ = w.WriteString("</sup>")
				return converter.RenderSuccess
			}
			if n.Type == html.ElementNode && strings.ToLower(n.Data) == "sub" {
				_, _ = w.WriteString("<sub>")
				_, _ = w.WriteString(extractRawText(n))
				_, _ = w.WriteString("</sub>")
				return converter.RenderSuccess
			}
			return converter.RenderTryNext
		},
		converter.PriorityLate,
	)

	conv.Register.Renderer(
		func(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
			if n.Type != html.ElementNode || strings.ToLower(n.Data) != "var" {
				return converter.RenderTryNext
			}

			latex := convertMathNode(n)

			// 可改成 $$...$$ 变成块级数学
			_, err := w.WriteString("$" + latex + "$")
			if err != nil {
				return converter.RenderTryNext
			}
			return converter.RenderSuccess
		},
		converter.PriorityStandard,
	)

	return nil
}

func extractRawText(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		} else {
			sb.WriteString(extractRawText(c))
		}
	}
	return sb.String()
}

// 递归转换 <sub>/<sup> → _{} / ^{}
func convertMathNode(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tag := strings.ToLower(c.Data)
		switch tag {
		case "sub":
			sb.WriteString("_{")
			sb.WriteString(convertMathNode(c))
			sb.WriteString("}")
		case "sup":
			sb.WriteString("^{")
			sb.WriteString(convertMathNode(c))
			sb.WriteString("}")
		default:
			sb.WriteString(convertMathNode(c))
		}
	}
	return sb.String()
}
