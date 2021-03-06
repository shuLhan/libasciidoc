package html5

import (
	"bytes"
	"html"
	"strconv"
	texttemplate "text/template"

	"github.com/bytesparadise/libasciidoc/pkg/renderer"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var fencedBlockTmpl texttemplate.Template
var listingBlockTmpl texttemplate.Template
var sourceBlockTmpl texttemplate.Template
var exampleBlockTmpl texttemplate.Template
var admonitionBlockTmpl texttemplate.Template
var quoteBlockTmpl texttemplate.Template
var verseBlockTmpl texttemplate.Template
var sidebarBlockTmpl texttemplate.Template

// initializes the templates
func init() {
	fencedBlockTmpl = newTextTemplate("listing block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="listingblock">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
<div class="content">
<pre class="highlight"><code>{{ range $index, $element := .Elements }}{{ renderPlainString $ctx $element | printf "%s" }}{{ end }}</code></pre>
</div>
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderPlainString": renderPlainString,
			"escape":            html.EscapeString,
		})

	listingBlockTmpl = newTextTemplate("listing block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="listingblock">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
<div class="content">
<pre>{{ range $index, $element := .Elements }}{{ renderPlainString $ctx $element | printf "%s" | escape }}{{ end }}</pre>
</div>
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderPlainString": renderPlainString,
			"escape":            html.EscapeString,
		})

	sourceBlockTmpl = newTextTemplate("source block",
		`{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="listingblock">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
<div class="content">
<pre class="highlight"><code{{ if .Language}} class="language-{{ .Language}}" data-lang="{{ .Language}}"{{ end }}>{{ range $index, $element := .Elements }}{{ renderPlainString $ctx $element | printf "%s" | escape }}{{ end }}</code></pre>
</div>
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderPlainString": renderPlainString,
			"escape":            html.EscapeString,
		})

	exampleBlockTmpl = newTextTemplate("example block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="exampleblock">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
<div class="content">
{{ $elements := .Elements }}{{ renderElements $ctx $elements | printf "%s" }}
</div>
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderElements": renderElements,
			"escape":         html.EscapeString,
		})

	quoteBlockTmpl = newTextTemplate("quote block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="quoteblock">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
<blockquote>
{{ renderElements $ctx .Elements | printf "%s" }}
</blockquote>{{ if .Attribution.First }}
<div class="attribution">
&#8212; {{ .Attribution.First }}{{ if .Attribution.Second }}<br>
<cite>{{ .Attribution.Second }}</cite>{{ end }}
</div>{{ end }}
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderElements": renderElements,
			"escape":         html.EscapeString,
		})

	verseBlockTmpl = newTextTemplate("verse block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="verseblock">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
<pre class="content">{{ renderElements $ctx .Elements | printf "%s" }}</pre>{{ if .Attribution.First }}
<div class="attribution">
&#8212; {{ .Attribution.First }}{{ if .Attribution.Second }}<br>
<cite>{{ .Attribution.Second }}</cite>{{ end }}
</div>{{ end }}
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderElements": renderElements,
			"escape":         html.EscapeString,
		})

	admonitionBlockTmpl = newTextTemplate("admonition block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID}}" {{ end }}class="admonitionblock {{ .Class }}">
<table>
<tr>
<td class="icon">
{{ if .IconClass }}<i class="fa icon-{{ .IconClass }}" title="{{ .IconTitle }}"></i>{{ else }}<div class="title">{{ .IconTitle }}</div>{{ end }}
</td>
<td class="content">
{{ if .Title }}<div class="title">{{ escape .Title }}</div>
{{ end }}{{ renderElements $ctx .Elements | printf "%s" }}
</td>
</tr>
</table>
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderElements": renderElements,
			"escape":         html.EscapeString,
		})

	sidebarBlockTmpl = newTextTemplate("sidebar block", `{{ $ctx := .Context }}{{ with .Data }}<div {{ if .ID }}id="{{ .ID }}" {{ end }}class="sidebarblock">
<div class="content">{{ if .Title }}
<div class="title">{{ escape .Title }}</div>{{ end }}
{{ renderElements $ctx .Elements | printf "%s" }}
</div>
</div>{{ end }}`,
		texttemplate.FuncMap{
			"renderElements": renderElements,
			"escape":         html.EscapeString,
		})
}

func renderDelimitedBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	log.Debugf("rendering delimited block of kind '%v'", b.Attributes[types.AttrKind])
	var err error
	kind := b.Kind
	switch kind {
	case types.Fenced:
		return renderFencedBlock(ctx, b)
	case types.Listing:
		return renderListingBlock(ctx, b)
	case types.Source:
		return renderSourceBlock(ctx, b)
	case types.Example:
		return renderExampleBlock(ctx, b)
	case types.Quote:
		return renderQuoteBlock(ctx, b)
	case types.Verse:
		return renderVerseBlock(ctx, b)
	case types.Sidebar:
		return renderSidebarBlock(ctx, b)
	case types.Comment:
		return renderCommentBlock(ctx, b)
	default:
		return nil, errors.Wrapf(err, "unable to render delimited block")
	}
}

func renderFencedBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	previouslyWithin := ctx.SetWithinDelimitedBlock(true)
	previouslyInclude := ctx.SetIncludeBlankLine(true)
	defer func() {
		ctx.SetWithinDelimitedBlock(previouslyWithin)
		ctx.SetIncludeBlankLine(previouslyInclude)
	}()
	result := bytes.NewBuffer(nil)
	err := fencedBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID       string
			Title    string
			Elements []interface{}
		}{
			ID:       generateID(ctx, b.Attributes),
			Title:    getTitle(b.Attributes),
			Elements: discardTrailingBlankLines(b.Elements),
		},
	})
	return result.Bytes(), err
}

func renderListingBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	previouslyWithin := ctx.SetWithinDelimitedBlock(true)
	previouslyInclude := ctx.SetIncludeBlankLine(true)
	defer func() {
		ctx.SetWithinDelimitedBlock(previouslyWithin)
		ctx.SetIncludeBlankLine(previouslyInclude)
	}()
	result := bytes.NewBuffer(nil)
	err := listingBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID       string
			Title    string
			Elements []interface{}
		}{
			ID:       generateID(ctx, b.Attributes),
			Title:    getTitle(b.Attributes),
			Elements: discardTrailingBlankLines(b.Elements),
		},
	})
	return result.Bytes(), err
}

func renderSourceBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	previouslyWithin := ctx.SetWithinDelimitedBlock(true)
	previouslyInclude := ctx.SetIncludeBlankLine(true)
	defer func() {
		ctx.SetWithinDelimitedBlock(previouslyWithin)
		ctx.SetIncludeBlankLine(previouslyInclude)
	}()
	language := b.Attributes.GetAsString(types.AttrLanguage)
	result := bytes.NewBuffer(nil)
	err := sourceBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID       string
			Title    string
			Language string
			Elements []interface{}
		}{
			ID:       generateID(ctx, b.Attributes),
			Title:    getTitle(b.Attributes),
			Language: language,
			Elements: discardTrailingBlankLines(b.Elements),
		},
	})
	return result.Bytes(), err
}

func renderExampleBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	if k, ok := b.Attributes[types.AttrAdmonitionKind].(types.AdmonitionKind); ok {
		err := admonitionBlockTmpl.Execute(result, ContextualPipeline{
			Context: ctx,
			Data: struct {
				ID        string
				Class     string
				IconClass string
				IconTitle string
				Title     string
				Elements  []interface{}
			}{
				ID:        generateID(ctx, b.Attributes),
				Class:     getClass(k),
				IconClass: getIconClass(ctx, k),
				IconTitle: getIconTitle(k),
				Title:     getTitle(b.Attributes),
				Elements:  discardTrailingBlankLines(b.Elements),
			},
		})
		return result.Bytes(), err
	}
	// default, example block
	var title string
	if b.Attributes.Has(types.AttrTitle) {
		title = "Example " + strconv.Itoa(ctx.GetAndIncrementExampleBlockCounter()) + ". " + getTitle(b.Attributes)
	}
	err := exampleBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID       string
			Title    string
			Elements []interface{}
		}{
			ID:       generateID(ctx, b.Attributes),
			Title:    title,
			Elements: discardTrailingBlankLines(b.Elements),
		},
	})
	return result.Bytes(), err
}

func renderQuoteBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	err := quoteBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID          string
			Title       string
			Attribution Attribution
			Elements    []interface{}
		}{
			ID:          generateID(ctx, b.Attributes),
			Title:       getTitle(b.Attributes),
			Attribution: NewDelimitedBlockAttribution(b),
			Elements:    b.Elements,
		},
	})
	return result.Bytes(), err
}

func renderVerseBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	var elements = make([]interface{}, 0)
	if len(b.Elements) > 0 {
		for _, element := range b.Elements {
			switch e := element.(type) {
			case types.Paragraph:
				for _, l := range e.Lines {
					elements = append(elements, l)
				}
			case types.BlankLine:
				elements = append(elements, e)
			default:
				log.Warnf("unexpected type of element to include in verse block: %T", element)
			}
		}
	}
	before := ctx.SetIncludeBlankLine(true)
	defer ctx.SetIncludeBlankLine(before)
	result := bytes.NewBuffer(nil)
	err := verseBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID          string
			Title       string
			Attribution Attribution
			Elements    []interface{}
		}{
			ID:          generateID(ctx, b.Attributes),
			Title:       getTitle(b.Attributes),
			Attribution: NewDelimitedBlockAttribution(b),
			Elements:    elements,
		},
	})
	return result.Bytes(), err
}

func renderCommentBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) { //nolint: unparam
	// comments block are not preserved during rendering
	return []byte{}, nil
}

func renderSidebarBlock(ctx *renderer.Context, b types.DelimitedBlock) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	err := sidebarBlockTmpl.Execute(result, ContextualPipeline{
		Context: ctx,
		Data: struct {
			ID       string
			Title    string
			Elements []interface{}
		}{
			ID:       generateID(ctx, b.Attributes),
			Title:    getTitle(b.Attributes),
			Elements: discardTrailingBlankLines(b.Elements),
		},
	})
	return result.Bytes(), err
}

func discardTrailingBlankLines(elements []interface{}) []interface{} {
	// discard blank lines at the end
	filteredElements := make([]interface{}, len(elements))
	copy(filteredElements, elements)
	for {
		if len(filteredElements) == 0 {
			break
		}
		if _, ok := filteredElements[len(filteredElements)-1].(types.BlankLine); ok {
			log.Debugf("element of type '%T' at position %d is a blank line, discarding it", filteredElements[len(filteredElements)-1], len(filteredElements)-1)
			// remove last element of the slice since it's a blankline
			filteredElements = filteredElements[:len(filteredElements)-1]
		} else {
			break
		}
	}
	return filteredElements
}
