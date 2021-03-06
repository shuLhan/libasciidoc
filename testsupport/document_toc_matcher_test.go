package testsupport_test

import (
	"fmt"

	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/bytesparadise/libasciidoc/testsupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("document table of contents assertions", func() {

	toc := types.DocumentAttributeDeclaration{
		Name: "toc",
	}
	preamble := types.Preamble{
		Elements: []interface{}{
			types.BlankLine{},
			types.Paragraph{
				Attributes: types.ElementAttributes{},
				Lines: []types.InlineElements{
					{
						types.StringElement{Content: "A short preamble"},
					},
				},
			},
			types.BlankLine{},
		},
	}
	section := types.Section{
		Level:      1,
		Attributes: types.ElementAttributes{},
		Title: types.InlineElements{
			types.StringElement{Content: "section 1"},
		},
		Elements: []interface{}{},
	}
	tableOfContents := types.TableOfContentsMacro{}

	actual := types.Document{
		Attributes:         types.DocumentAttributes{},
		ElementReferences:  types.ElementReferences{}, // can leave empty for this test
		Footnotes:          types.Footnotes{},
		FootnoteReferences: types.FootnoteReferences{},
		Elements: []interface{}{
			toc,
			preamble,
			section,
		},
	}
	expected := types.Document{
		Attributes:         types.DocumentAttributes{},
		ElementReferences:  types.ElementReferences{},
		Footnotes:          types.Footnotes{},
		FootnoteReferences: types.FootnoteReferences{},
		Elements: []interface{}{
			tableOfContents,
			toc,
			preamble,
			section,
		},
	}

	It("should match", func() {
		// given
		matcher := testsupport.HaveTableOfContents(expected)
		// when
		result, err := matcher.Match(actual)
		// then
		GinkgoT().Log(matcher.FailureMessage(actual))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeTrue())
	})

	It("should not match", func() {
		// given
		matcher := testsupport.HaveTableOfContents(expected)
		// when
		result, err := matcher.Match(types.Document{})
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeFalse())
		Expect(matcher.FailureMessage(actual)).To(Equal(fmt.Sprintf("expected document to match:\n\texpected: '%v'\n\tactual:   '%v'", expected, types.Document{})))
		Expect(matcher.NegatedFailureMessage(actual)).To(Equal(fmt.Sprintf("expected document not to match:\n\texpected: '%v'\n\tactual:   '%v'", expected, types.Document{})))
	})

	It("should return error when invalid type is input", func() {
		// given
		matcher := testsupport.HaveTableOfContents(types.Document{})
		// when
		result, err := matcher.Match(1) // not a doc
		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("HaveTableOfContents matcher expects a Document (actual: int)"))
		Expect(result).To(BeFalse())
	})
})
