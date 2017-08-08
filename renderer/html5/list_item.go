package html5

import (
	"bytes"
	"context"
	"html/template"

	"github.com/bytesparadise/libasciidoc/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var unorderedListTmpl *template.Template
var listItemTmpl *template.Template
var listItemContentTmpl *template.Template

// initializes the templates
func init() {
	unorderedListTmpl = newTemplate("list", `<div{{ if .ID }} id="{{.ID.Value}}"{{ end }} class="ulist">
<ul>
{{.Items}}
</ul>
</div>`)
	listItemTmpl = newTemplate("list item", `<li>
{{.Content}}{{ if .Children }}
{{.Children}}{{ end }}
</li>`)
	listItemContentTmpl = newTemplate("list item content", `<p>{{.}}</p>`)
}

func renderList(ctx context.Context, list types.List) ([]byte, error) {
	renderedElementsBuff := bytes.NewBuffer(make([]byte, 0))
	for i, item := range list.Items {
		renderedListItem, err := renderListItem(ctx, *item)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to render list of items")
		}
		renderedElementsBuff.Write(renderedListItem)
		if i < len(list.Items)-1 {
			renderedElementsBuff.WriteString("\n")
		}
	}

	result := bytes.NewBuffer(make([]byte, 0))
	// here we must preserve the HTML tags
	err := unorderedListTmpl.Execute(result, struct {
		ID    *types.ElementID
		Items template.HTML
	}{
		ID:    list.ID,
		Items: template.HTML(renderedElementsBuff.String()),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to render list of items")
	}
	log.Debugf("rendered list of items: %s", result.Bytes())
	return result.Bytes(), nil
}

func renderListItem(ctx context.Context, item types.ListItem) ([]byte, error) {
	renderedItemContent, err := renderListItemContent(ctx, *item.Content)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to render list item")
	}
	result := bytes.NewBuffer(make([]byte, 0))
	var renderedChildrenOutput *template.HTML
	if item.Children != nil {
		childrenOutput, err := renderList(ctx, *item.Children)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to render list item")
		}
		htmlChildrenOutput := template.HTML(string(childrenOutput))
		renderedChildrenOutput = &htmlChildrenOutput
	}
	err = listItemTmpl.Execute(result, struct {
		Content  template.HTML
		Children *template.HTML
	}{
		Content:  template.HTML(string(renderedItemContent)),
		Children: renderedChildrenOutput,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to render list item")
	}
	log.Debugf("rendered item: %s", result.Bytes())
	return result.Bytes(), nil
}

func renderListItemContent(ctx context.Context, content types.ListItemContent) ([]byte, error) {
	renderedLinesBuff := bytes.NewBuffer(make([]byte, 0))
	for _, line := range content.Lines {
		renderedLine, err := renderInlineContent(ctx, *line)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render list item content")
		}
		renderedLinesBuff.Write(renderedLine)
	}
	result := bytes.NewBuffer(make([]byte, 0))
	err := listItemContentTmpl.Execute(result, renderedLinesBuff.String())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to render list item")
	}
	log.Debugf("rendered item content: %s", result.Bytes())
	return result.Bytes(), nil
}
