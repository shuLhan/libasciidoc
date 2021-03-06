package html5

import (
	"bytes"
	"strconv"
	"strings"
	texttemplate "text/template"

	"github.com/bytesparadise/libasciidoc/pkg/renderer"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/pkg/errors"
)

var footnoteTmpl texttemplate.Template
var footnoterefTmpl texttemplate.Template
var invalidFootnoteTmpl texttemplate.Template
var footnotesTmpl texttemplate.Template

// initializes the templates
func init() {
	footnoteTmpl = newTextTemplate("footnote", `<sup class="{{ .Class }}"{{ if .Ref }} id="_footnote_{{ .Ref }}"{{ end }}>[<a id="_footnoteref_{{ renderIndex .ID }}" class="footnote" href="#_footnotedef_{{ renderIndex .ID }}" title="View footnote.">{{ renderIndex .ID }}</a>]</sup>`,
		texttemplate.FuncMap{
			"renderIndex": renderFootnoteIndex,
		})
	footnoterefTmpl = newTextTemplate("footnote ref", `<sup class="{{ .Class }}">[<a class="footnote" href="#_footnotedef_{{ renderIndex .ID }}" title="View footnote.">{{ renderIndex .ID }}</a>]</sup>`,
		texttemplate.FuncMap{
			"renderIndex": renderFootnoteIndex,
		})

	invalidFootnoteTmpl = newTextTemplate("invalid footnote", `<sup class="{{ .Class }} red" title="Unresolved footnote reference.">[{{ .Ref }}]</sup>`)
	footnotesTmpl = newTextTemplate("footnotes", `
<div id="footnotes">
<hr>{{ $ctx := .Context }}{{ with .Data }}{{ $footnotes := .Footnotes }}{{ range $index, $footnote := $footnotes }}
<div class="footnote" id="_footnotedef_{{ renderIndex $index }}">
<a href="#_footnoteref_{{ renderIndex $index }}">{{ renderIndex $index }}</a>. {{ renderFootnoteContent $ctx $footnote.Elements }}
</div>{{ end }}{{ end }}
</div>`,
		texttemplate.FuncMap{
			"renderFootnoteContent": func(ctx *renderer.Context, element interface{}) (string, error) {
				result, err := renderElement(ctx, element)
				if err != nil {
					return "", errors.Wrapf(err, "unable to render foot note content")
				}
				return strings.TrimSpace(string(result)), nil
			},
			"renderIndex": renderFootnoteIndex,
		})
}

func renderFootnoteIndex(idx int) string {
	return strconv.Itoa(idx + 1)
}

func renderFootnote(ctx *renderer.Context, note types.Footnote) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	ref := ""
	noteRef, hasRef := ctx.Document.FootnoteReferences[note.Ref]
	if hasRef {
		ref = note.Ref
	}
	if id, ok := ctx.Document.Footnotes.IndexOf(note); ok {
		// valid case for a footnte with content, with our without an explicit reference
		err := footnoteTmpl.Execute(result, struct {
			ID    int
			Ref   string
			Class string
		}{
			ID:    id,
			Ref:   ref,
			Class: "footnote",
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to render footnote")
		}
	} else if hasRef {
		err := footnoterefTmpl.Execute(result, struct {
			ID    int
			Ref   string
			Class string
		}{
			ID:    noteRef.ID,
			Ref:   ref,
			Class: "footnoteref",
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to render footnote")
		}
	} else {
		// invalid footnote
		err := invalidFootnoteTmpl.Execute(result, struct {
			Ref   string
			Class string
		}{
			Ref:   note.Ref,
			Class: "footnoteref",
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to render missing footnote")
		}
	}

	return result.Bytes(), nil
}

func renderFootnotes(ctx *renderer.Context, notes types.Footnotes) ([]byte, error) {
	// skip if there's no foot note in the doc
	if len(notes) == 0 {
		return []byte{}, nil
	}
	result := bytes.NewBuffer(nil)
	err := footnotesTmpl.Execute(result,
		ContextualPipeline{
			Context: ctx,
			Data: struct {
				Footnotes types.Footnotes
			}{
				Footnotes: notes,
			},
		})
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to render footnotes")
	}
	return result.Bytes(), nil
}
