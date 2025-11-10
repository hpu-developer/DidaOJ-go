package foundationrender

import (
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
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

	conv.Register.RendererFor("sup", converter.TagTypeInline, base.RenderAsHTML, converter.PriorityEarly)
	conv.Register.RendererFor("sub", converter.TagTypeInline, base.RenderAsHTML, converter.PriorityEarly)

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
