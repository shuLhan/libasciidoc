package types

import (
	"fmt"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ------------------------------------------
// DocElement (and other interfaces)
// ------------------------------------------

// DocElement the interface for all document elements
type DocElement interface {
	// Visitable
}

// InlineElement the interface for inline elements
type InlineElement interface {
	DocElement
}

// Visitable the interface for visitable elements
type Visitable interface {
	Accept(Visitor) error
}

// Visitor a visitor that can visit/traverse the DocElement and its children (if applicable)
type Visitor interface {
	BeforeVisit(Visitable) error
	Visit(Visitable) error
	AfterVisit(Visitable) error
}

// ------------------------------------------
// Document
// ------------------------------------------

// Document the top-level structure for a document
type Document struct {
	Attributes        DocumentAttributes
	Elements          []DocElement
	ElementReferences ElementReferences
}

// NewDocument initializes a new `Document` from the given lines
func NewDocument(frontmatter, header interface{}, blocks []interface{}) (*Document, error) {
	log.Debugf("Initializing a new Document with %d blocks(s)", len(blocks))
	for i, block := range blocks {
		log.Debugf("Line #%d: %T", i, block)
	}
	// elements := convertBlocksToDocElements(blocks)
	elements := filterUnrelevantElements(blocks)
	attributes := make(map[string]interface{})

	if frontmatter != nil {
		for attrName, attrValue := range frontmatter.(*FrontMatter).Content {
			attributes[attrName] = attrValue
		}
	}
	if header != nil {
		for attrName, attrValue := range header.(*DocumentHeader).Content {
			attributes[attrName] = attrValue
			if attrName == "toc" {
				// insert a TableOfContentsMacro element if `toc` value is:
				// - "auto" (or empty)
				// - "preamble"
				switch attrValue {
				case "", "auto":
					// insert TableOfContentsMacro at first position
					elements = append([]DocElement{&TableOfContentsMacro{}}, elements...)
				case "preamble":
					// lookup preamble in elements (should be first)
					preambleIndex := 0
					for i, e := range elements {
						if _, ok := e.(*Preamble); ok {
							preambleIndex = i
							break
						}
					}
					// insert TableOfContentsMacro just after preamble
					remainingElements := make([]DocElement, len(elements)-(preambleIndex+1))
					copy(remainingElements, elements[preambleIndex+1:])
					elements = append(elements[0:preambleIndex+1], &TableOfContentsMacro{})
					elements = append(elements, remainingElements...)
				case "macro":
				default:
					log.Warnf("invalid value for 'toc' attribute: '%s'", attrValue)

				}
			}
		}
	}

	c := NewElementReferencesCollector()
	for _, e := range elements {
		if v, ok := e.(Visitable); ok {
			v.Accept(c)
		}
	}
	document := &Document{
		Attributes:        attributes,
		Elements:          elements,
		ElementReferences: c.ElementReferences,
	}

	// visit all elements in the `AST` to retrieve their reference (ie, their ElementID if they have any)
	return document, nil
}

// ------------------------------------------
// Document Header
// ------------------------------------------

// DocumentHeader the document header
type DocumentHeader struct {
	Content DocumentAttributes
}

// NewDocumentHeader initializes a new DocumentHeader
func NewDocumentHeader(header, authors, revision interface{}, otherAttributes []interface{}) (*DocumentHeader, error) {
	content := DocumentAttributes{}
	if header != nil {
		content["doctitle"] = header.(*SectionTitle)
	}
	log.Debugf("Initializing a new DocumentHeader with content '%v', authors '%+v' and revision '%+v'", content, authors, revision)
	if authors != nil {
		for i, author := range authors.([]*DocumentAuthor) {
			if i == 0 {
				content.Add("firstname", author.FirstName)
				content.Add("middlename", author.MiddleName)
				content.Add("lastname", author.LastName)
				content.Add("author", author.FullName)
				content.Add("authorinitials", author.Initials)
				content.Add("email", author.Email)
			} else {
				content.Add(fmt.Sprintf("firstname_%d", i+1), author.FirstName)
				content.Add(fmt.Sprintf("middlename_%d", i+1), author.MiddleName)
				content.Add(fmt.Sprintf("lastname_%d", i+1), author.LastName)
				content.Add(fmt.Sprintf("author_%d", i+1), author.FullName)
				content.Add(fmt.Sprintf("authorinitials_%d", i+1), author.Initials)
				content.Add(fmt.Sprintf("email_%d", i+1), author.Email)
			}
		}
	}
	if revision != nil {
		rev := revision.(*DocumentRevision)
		content.Add("revnumber", rev.Revnumber)
		content.Add("revdate", rev.Revdate)
		content.Add("revremark", rev.Revremark)
	}
	for _, attr := range otherAttributes {
		if attr, ok := attr.(*DocumentAttributeDeclaration); ok {
			content.AddAttribute(attr)
		}
	}
	return &DocumentHeader{
		Content: content,
	}, nil
}

// ------------------------------------------
// Document Author
// ------------------------------------------

// DocumentAuthor a document author
type DocumentAuthor struct {
	FullName   string
	Initials   string
	FirstName  *string
	MiddleName *string
	LastName   *string
	Email      *string
}

// NewDocumentAuthors converts the given authors into an array of `DocumentAuthor`
func NewDocumentAuthors(authors []interface{}) ([]*DocumentAuthor, error) {
	log.Debugf("Initializing a new array of document authors from `%+v`", authors)
	result := make([]*DocumentAuthor, len(authors))
	for i, author := range authors {
		switch author.(type) {
		case *DocumentAuthor:
			result[i] = author.(*DocumentAuthor)
		default:
			return nil, errors.Errorf("unexpected type of author: %T", author)
		}
	}
	return result, nil
}

//NewDocumentAuthor initializes a new DocumentAuthor
func NewDocumentAuthor(namePart1, namePart2, namePart3, emailAddress interface{}) (*DocumentAuthor, error) {
	var part1, part2, part3, email *string
	var err error
	if namePart1 != nil {
		part1, err = stringify(namePart1.([]interface{}),
			func(s string) (string, error) {
				return strings.TrimSpace(s), nil
			},
			func(s string) (string, error) {
				return strings.Replace(s, "_", " ", -1), nil
			},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentAuthor")
		}
	}
	if namePart2 != nil {
		part2, err = stringify(namePart2.([]interface{}),
			func(s string) (string, error) {
				return strings.TrimSpace(s), nil
			},
			func(s string) (string, error) {
				return strings.Replace(s, "_", " ", -1), nil
			},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentAuthor")
		}
	}
	if namePart3 != nil {
		part3, err = stringify(namePart3.([]interface{}),
			func(s string) (string, error) {
				return strings.TrimSpace(s), nil
			},
			func(s string) (string, error) {
				return strings.Replace(s, "_", " ", -1), nil
			},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentAuthor")
		}
	}
	if emailAddress != nil {
		email, err = stringify(emailAddress.([]interface{}),
			func(s string) (string, error) {
				return strings.TrimPrefix(s, "<"), nil
			}, func(s string) (string, error) {
				return strings.TrimSuffix(s, ">"), nil
			}, func(s string) (string, error) {
				return strings.TrimSpace(s), nil
			})
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentAuthor")
		}
	}
	result := new(DocumentAuthor)
	if part2 != nil && part3 != nil {
		result.FirstName = part1
		result.MiddleName = part2
		result.LastName = part3
		result.FullName = fmt.Sprintf("%s %s %s", *part1, *part2, *part3)
		result.Initials = initials(*result.FirstName, *result.MiddleName, *result.LastName)
	} else if part2 != nil {
		result.FirstName = part1
		result.LastName = part2
		result.FullName = fmt.Sprintf("%s %s", *part1, *part2)
		result.Initials = initials(*result.FirstName, *result.LastName)
	} else {
		result.FirstName = part1
		result.FullName = *part1
		result.Initials = initials(*result.FirstName)
	}
	result.Email = email
	log.Debugf("Initialized a new document author: `%v`", result.String())
	return result, nil
}

func initials(firstPart string, otherParts ...string) string {
	result := fmt.Sprintf("%s", firstPart[0:1])
	if otherParts != nil {
		for _, otherPart := range otherParts {
			result = result + otherPart[0:1]
		}
	}
	return result
}

func (a *DocumentAuthor) String() string {
	email := ""
	if a.Email != nil {
		email = *a.Email
	}
	return fmt.Sprintf("%s (%s)", a.FullName, email)
}

// ------------------------------------------
// Document Revision
// ------------------------------------------

// DocumentRevision a document revision
type DocumentRevision struct {
	Revnumber *string
	Revdate   *string
	Revremark *string
}

// NewDocumentRevision intializes a new DocumentRevision
func NewDocumentRevision(revnumber, revdate, revremark interface{}) (*DocumentRevision, error) {
	// log.Debugf("Initializing document revision with revnumber=%v, revdate=%v, revremark=%v", revnumber, revdate, revremark)
	// stringify, then remove the "v" prefix and trim spaces
	var number, date, remark *string
	var err error
	if revnumber != nil {
		number, err = stringify(revnumber.([]interface{}),
			func(s string) (string, error) {
				return strings.TrimPrefix(s, "v"), nil
			}, func(s string) (string, error) {
				return strings.TrimPrefix(s, "V"), nil
			}, func(s string) (string, error) {
				return strings.TrimSpace(s), nil
			})
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentRevision")
		}
	}
	if revdate != nil {
		// stringify, then remove the "," prefix and trim spaces
		date, err = stringify(revdate.([]interface{}), func(s string) (string, error) {
			return strings.TrimSpace(s), nil
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentRevision")
		}
		// do not keep empty values
		if *date == "" {
			date = nil
		}
	}
	if revremark != nil {
		// then we need to strip the heading "," and spaces
		remark, err = stringify(revremark.([]interface{}),
			func(s string) (string, error) {
				return strings.TrimPrefix(s, ":"), nil
			}, func(s string) (string, error) {
				return strings.TrimSpace(s), nil
			})
		if err != nil {
			return nil, errors.Wrapf(err, "error while initializing a DocumentRevision")
		}
		// do not keep empty values
		if *remark == "" {
			remark = nil
		}
	}
	// log.Debugf("Initializing a new DocumentRevision with revnumber='%v', revdate='%v' and revremark='%v'", *n, *d, *r)
	result := DocumentRevision{
		Revnumber: number,
		Revdate:   date,
		Revremark: remark,
	}
	log.Debugf("Initialized a new document revision: `%s`", result.String())
	return &result, nil
}

func (r *DocumentRevision) String() string {
	number := ""
	if r.Revnumber != nil {
		number = *r.Revnumber
	}
	date := ""
	if r.Revdate != nil {
		date = *r.Revdate
	}
	remark := ""
	if r.Revremark != nil {
		remark = *r.Revremark
	}
	return fmt.Sprintf("%v, %v: %v", number, date, remark)
}

// ------------------------------------------
// Document Attributes
// ------------------------------------------

// DocumentAttributeDeclaration the type for Document Attribute Declarations
type DocumentAttributeDeclaration struct {
	Name  string
	Value string
}

// NewDocumentAttributeDeclaration initializes a new DocumentAttributeDeclaration
func NewDocumentAttributeDeclaration(name []interface{}, value []interface{}) (*DocumentAttributeDeclaration, error) {
	attrName, err := stringify(name,
		func(s string) (string, error) {
			return strings.TrimSpace(s), nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "error while initializing a DocumentAttributeDeclaration")
	}
	attrValue, err := stringify(value,
		func(s string) (string, error) {
			return strings.TrimSpace(s), nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "error while initializing a DocumentAttributeDeclaration")
	}
	log.Debugf("Initialized a new DocumentAttributeDeclaration: '%s' -> '%s'", *attrName, *attrValue)
	return &DocumentAttributeDeclaration{
		Name:  *attrName,
		Value: *attrValue,
	}, nil
}

// DocumentAttributeReset the type for DocumentAttributeReset
type DocumentAttributeReset struct {
	Name string
}

// NewDocumentAttributeReset initializes a new Document Attribute Resets.
func NewDocumentAttributeReset(name []interface{}) (*DocumentAttributeReset, error) {
	attrName, err := stringify(name)
	if err != nil {
		return nil, errors.Wrapf(err, "error while initializing a DocumentAttributeReset")
	}
	log.Debugf("Initialized a new DocumentAttributeReset: '%s'", *attrName)
	return &DocumentAttributeReset{Name: *attrName}, nil
}

// DocumentAttributeSubstitution the type for DocumentAttributeSubstitution
type DocumentAttributeSubstitution struct {
	Name string
}

// NewDocumentAttributeSubstitution initializes a new Document Attribute Substitutions
func NewDocumentAttributeSubstitution(name []interface{}) (*DocumentAttributeSubstitution, error) {
	attrName, err := stringify(name)
	if err != nil {
		return nil, errors.Wrapf(err, "error while initializing a DocumentAttributeSubstitution")
	}
	log.Debugf("Initialized a new DocumentAttributeSubstitution: '%s'", *attrName)
	return &DocumentAttributeSubstitution{Name: *attrName}, nil
}

// ------------------------------------------
// Table of Contents
// ------------------------------------------

// TableOfContentsMacro the structure for Table of Contents
type TableOfContentsMacro struct {
}

// ------------------------------------------
// Preamble
// ------------------------------------------

// Preamble the structure for document Preamble
type Preamble struct {
	Elements []DocElement
}

// NewPreamble initializes a new Preamble from the given elements
func NewPreamble(elements []interface{}) (*Preamble, error) {
	log.Debugf("Initialiazing new Preamble with %d elements", len(elements))
	return &Preamble{Elements: filterUnrelevantElements(elements)}, nil
}

// ------------------------------------------
// Front Matter
// ------------------------------------------

// FrontMatter the structure for document front-matter
type FrontMatter struct {
	Content map[string]interface{}
}

// NewYamlFrontMatter initializes a new FrontMatter from the given `content`
func NewYamlFrontMatter(content string) (*FrontMatter, error) {
	attributes := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(content), &attributes)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse yaml content in front-matter of document")
	}
	log.Debugf("Initialized a new FrontMatter with attributes: %+v", attributes)
	return &FrontMatter{Content: attributes}, nil
}

// ------------------------------------------
// Sections
// ------------------------------------------

// Section the structure for a section
type Section struct {
	Level        int
	SectionTitle SectionTitle
	Elements     []DocElement
}

// NewSection initializes a new `Section` from the given section title and elements
func NewSection(level int, sectionTitle *SectionTitle, blocks []interface{}) (*Section, error) {
	// log.Debugf("Initializing a new Section with %d block(s)", len(blocks))
	elements := filterUnrelevantElements(blocks)
	log.Debugf("Initialized a new Section of level %d with %d block(s)", level, len(blocks))
	return &Section{
		Level:        level,
		SectionTitle: *sectionTitle,
		Elements:     elements,
	}, nil
}

// Accept implements Visitable#Accept(Visitor)
func (s *Section) Accept(v Visitor) error {
	err := v.BeforeVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting section")
	}
	err = v.Visit(s)
	if err != nil {
		return errors.Wrapf(err, "error while visiting section")
	}
	for _, element := range s.Elements {
		if visitable, ok := element.(Visitable); ok {
			err = visitable.Accept(v)
			if err != nil {
				return errors.Wrapf(err, "error while visiting section element")
			}
		}

	}
	err = v.AfterVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting section")
	}
	return nil
}

// ------------------------------------------
// SectionTitle
// ------------------------------------------

// SectionTitle the structure for the section titles
type SectionTitle struct {
	ID      *ElementID
	Content *InlineContent
}

// NewSectionTitle initializes a new `SectionTitle`` from the given level and content, with the optional attributes.
// In the attributes, only the ElementID is retained
func NewSectionTitle(inlineContent *InlineContent, attributes []interface{}) (*SectionTitle, error) {
	// counting the lenght of the 'level' value (ie, the number of `=` chars)
	id, _, _ := newElementAttributes(attributes)
	// make a default id from the sectionTitle's inline content
	if id == nil {
		replacement, err := ReplaceNonAlphanumerics(inlineContent, "_")
		if err != nil {
			return nil, errors.Wrapf(err, "unable to generate default ID while instanciating a new SectionTitle element")
		}
		id, _ = NewElementID(*replacement)
	}
	sectionTitle := SectionTitle{Content: inlineContent, ID: id}
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Initialized a new SectionTitle:")
		spew.Dump(sectionTitle)
	}
	return &sectionTitle, nil
}

// ------------------------------------------
// Lists
// ------------------------------------------

// List the structure for the lists
type List struct {
	ID    *ElementID
	Items []*ListItem
}

// NewList initializes a new `ListItem` from the given content
func NewList(elements []interface{}, attributes []interface{}) (*List, error) {
	id, _, _ := newElementAttributes(attributes)
	items := make([]*ListItem, 0)
	log.Debugf("Initializing a new List from %d elements", len(elements))
	currentLevel := 1
	lastItems := make([]*ListItem, 10)
	for _, element := range elements {
		// each "list item" can be a "list item" element followed by an optional blank line (ignored during the processing below)
		//  also, a list item may need to be divided when it contains lines starting with a caret or a group of stars...
		if itemElements, ok := element.([]interface{}); ok {
			if item, ok := itemElements[0].(*ListItem); ok {
				if item.Level == 1 {
					items = append(items, item)
				} else if item.Level > currentLevel {
					// force the current item level to (last seen level + 1)
					item.Level = currentLevel + 1
				}

				if item.Level > 1 {
					// now join the item to its parent
					parentItem := lastItems[item.Level-2]
					if parentItem.Children == nil {
						parentItem.Children = &List{}
					}
					parentItem.Children.Items = append(parentItem.Children.Items, item)
				}
				if item.Level < currentLevel { // remove some items
					lastItems = lastItems[:item.Level+1]
				}
				currentLevel = item.Level
				lastItems[item.Level-1] = item
			}
		}
	}
	log.Debugf("Initialized a new List with %d item(s)", len(items))
	return &List{
		ID:    id,
		Items: items,
	}, nil
}

// ListItem the structure for the list items
type ListItem struct {
	Level    int
	Content  *ListItemContent
	Children *List
}

// NewListItem initializes a new `ListItem` from the given content
func NewListItem(level interface{}, content *ListItemContent, children *List) (*ListItem, error) {
	switch vals := reflect.ValueOf(level); vals.Kind() {
	case reflect.Slice:
		log.Debugf("Initializing a new ListItem with content '%s' lines and input level '%d'", content, vals.Len())
		return &ListItem{
			Level:    vals.Len(),
			Content:  content,
			Children: children,
		}, nil
	default:
		return nil, errors.Errorf("Unable to initialize a ListItem with level '%v", level)
	}
}

// ListItemContent the structure for the list item content
type ListItemContent struct {
	Lines []*InlineContent
}

// NewListItemContent initializes a new `ListItemContent`
func NewListItemContent(lines []interface{}) (*ListItemContent, error) {
	log.Debugf("Initializing a new ListItemContent with %d line(s)", len(lines))
	typedLines := make([]*InlineContent, 0)
	for _, line := range lines {
		// here, `line` is an []interface{} in which we need to locate the relevant `*InlineContent` fragment
		if lineFragments, ok := line.([]interface{}); ok {
			for i := range lineFragments {
				if fragment, ok := lineFragments[i].(*InlineContent); ok {
					typedLines = append(typedLines, fragment)
				}
			}
		}
	}
	return &ListItemContent{Lines: typedLines}, nil
}

// ------------------------------------------
// Paragraph
// ------------------------------------------

// Paragraph the structure for the paragraph
type Paragraph struct {
	Lines []*InlineContent
	ID    *ElementID
	Title *ElementTitle
}

// NewParagraph initializes a new `Paragraph`
func NewParagraph(lines []interface{}, attributes []interface{}) (*Paragraph, error) {
	log.Debugf("Initializing a new Paragraph with %d line(s)", len(lines))
	id, title, _ := newElementAttributes(attributes)

	typedLines := make([]*InlineContent, 0)
	for _, line := range lines {
		// each `line` element is an array with the actual `InlineContent` + `EOF`
		typedLines = append(typedLines, line.([]interface{})[0].(*InlineContent))
	}
	return &Paragraph{
		Lines: typedLines,
		ID:    id,
		Title: title,
	}, nil
}

// ------------------------------------------
// InlineContent
// ------------------------------------------

// InlineContent the structure for the lines in paragraphs
type InlineContent struct {
	// Input    []byte
	Elements []InlineElement
}

// NewInlineContent initializes a new `InlineContent` from the given values
func NewInlineContent(elements []interface{}) (*InlineContent, error) {
	inlineElements, err := toInlineElements(elements)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize a new InlineContent")
	}
	return &InlineContent{Elements: inlineElements}, nil
}

// Accept implements Visitable#Accept(Visitor)
func (c *InlineContent) Accept(v Visitor) error {
	err := v.BeforeVisit(c)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting inline content")
	}
	err = v.Visit(c)
	if err != nil {
		return errors.Wrapf(err, "error while visiting inline content")
	}
	for _, element := range c.Elements {
		if visitable, ok := element.(Visitable); ok {
			err = visitable.Accept(v)
			if err != nil {
				return errors.Wrapf(err, "error while visiting inline content element")
			}
		}

	}
	err = v.AfterVisit(c)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting sectionTitle")
	}
	return nil
}

// ------------------------------------------
// Cross References
// ------------------------------------------

// CrossReference the struct for Cross References
type CrossReference struct {
	ID string
}

// NewCrossReference initializes a new `CrossReference` from the given ID
func NewCrossReference(id string) (*CrossReference, error) {
	log.Debugf("Initializing a new CrossReference with ID=%s", id)
	return &CrossReference{ID: id}, nil
}

// ------------------------------------------
// Images
// ------------------------------------------

// BlockImage the structure for the block images
type BlockImage struct {
	Macro ImageMacro
	ID    *ElementID
	Title *ElementTitle
	Link  *ElementLink
}

// NewBlockImage initializes a new `BlockImage`
func NewBlockImage(imageMacro ImageMacro, attributes []interface{}) (*BlockImage, error) {
	id, title, link := newElementAttributes(attributes)
	return &BlockImage{
		Macro: imageMacro,
		ID:    id,
		Title: title,
		Link:  link,
	}, nil
}

// InlineImage the structure for the inline image macros
type InlineImage struct {
	Macro ImageMacro
}

// NewInlineImage initializes a new `InlineImage` (similar to BlockImage, but without attributes)
func NewInlineImage(imageMacro ImageMacro) (*InlineImage, error) {
	return &InlineImage{
		Macro: imageMacro,
	}, nil
}

// ImageMacro the structure for the block image macros
type ImageMacro struct {
	Path   string
	Alt    string
	Width  *string
	Height *string
}

// NewImageMacro initializes a new `ImageMacro`
func NewImageMacro(path string, attributes interface{}) (*ImageMacro, error) {
	var alt string
	var width, height *string
	if attributes != nil {
		// optionally, the width and height can be specified in the alt text, using `,` as a separator
		// eg: `image::foo.png[a title,200,100]`
		splittedAttributes := strings.Split(attributes.(string), ",")
		// naively assume that if the splitted 'alt' contains more than 3 elements, the 2 last ones are for the width and height
		splitCount := len(splittedAttributes)
		alt = splittedAttributes[0]
		if splitCount > 1 {
			w := strings.Trim(splittedAttributes[1], " ")
			width = &w
		}
		if splitCount > 2 {
			h := strings.Trim(splittedAttributes[2], " ")
			height = &h
		}
	} else {
		dir := filepath.Dir(path)
		extension := filepath.Ext(path)
		var offset int
		if dir == "." {
			offset = 0
		} else {
			offset = len(dir) + 1
		}
		alt = path[offset : len(path)-len(extension)]
	}
	return &ImageMacro{
		Path:   path,
		Alt:    alt,
		Width:  width,
		Height: height}, nil
}

// ------------------------------------------
// Delimited blocks
// ------------------------------------------

// DelimitedBlockKind the type for delimited blocks
type DelimitedBlockKind int

const (
	// FencedBlock a fenced block
	FencedBlock DelimitedBlockKind = iota
	// ListingBlock a listing block
	ListingBlock
)

// DelimitedBlock the structure for the delimited blocks
type DelimitedBlock struct {
	Kind    DelimitedBlockKind
	Content string
}

// NewDelimitedBlock initializes a new `DelimitedBlock` of the given kind with the given content
func NewDelimitedBlock(kind DelimitedBlockKind, content []interface{}) (*DelimitedBlock, error) {
	blockContent, err := stringify(content,
		// remove "\n" or "\r\n", depending on the OS.
		func(s string) (string, error) {
			return strings.TrimSuffix(s, "\n"), nil
		}, func(s string) (string, error) {
			return strings.TrimSuffix(s, "\r"), nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize a new delimited block")
	}
	log.Debugf("Initialized a new DelimitedBlock with content=`%s`", *blockContent)
	return &DelimitedBlock{
		Kind:    kind,
		Content: *blockContent,
	}, nil
}

// ------------------------------------------
// Literal blocks
// ------------------------------------------

// LiteralBlock the structure for the literal blocks
type LiteralBlock struct {
	Content string
}

// NewLiteralBlock initializes a new `DelimitedBlock` of the given kind with the given content,
// along with the given sectionTitle spaces
func NewLiteralBlock(spaces, content []interface{}) (*LiteralBlock, error) {
	// concatenates the spaces with the actual content in a single 'stringified' value
	// log.Debugf("Initializing a new LiteralBlock with spaces='%v' and content=`%v`", spaces, content)
	c, err := stringify(append(spaces, content...))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize a new literal block")
	}
	// remove "\n" or "\r\n", depending on the OS.
	blockContent := strings.TrimSuffix(strings.TrimSuffix(*c, "\n"), "\r")
	log.Debugf("Initialized a new LiteralBlock with content=`%s`", blockContent)
	return &LiteralBlock{
		Content: blockContent,
	}, nil
}

// ------------------------------------------
// Meta Elements
// ------------------------------------------

func newElementAttributes(attributes []interface{}) (*ElementID, *ElementTitle, *ElementLink) {
	var id *ElementID
	var title *ElementTitle
	var link *ElementLink
	for _, item := range attributes {
		switch item := item.(type) {
		case *ElementID:
			id = item
		case *ElementLink:
			link = item
		case *ElementTitle:
			title = item
		default:
			log.Warnf("Unexpected attributes: %T", item)
		}
	}
	return id, title, link
}

// ElementLink the structure for element links
type ElementLink struct {
	Path string
}

// NewElementLink initializes a new `ElementLink` from the given path
func NewElementLink(path string) (*ElementLink, error) {
	log.Debugf("Initializing a new ElementLink with path=%s", path)
	return &ElementLink{Path: path}, nil
}

// ElementID the structure for element IDs
type ElementID struct {
	Value string
}

// NewElementID initializes a new `ElementID` from the given ID
func NewElementID(id string) (*ElementID, error) {
	log.Debugf("Initializing a new ElementID with ID=%s", id)
	return &ElementID{Value: id}, nil
}

// ElementTitle the structure for element Titles
type ElementTitle struct {
	Value string
}

// NewElementTitle initializes a new `ElementTitle` from the given value
func NewElementTitle(value []interface{}) (*ElementTitle, error) {
	v, err := stringify(value)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize a new ElementTitle")
	}
	log.Debugf("Initializing a new ElementTitle with content=%s", *v)
	return &ElementTitle{Value: *v}, nil
}

// InvalidElementAttribute the struct for invalid element attributes
type InvalidElementAttribute struct {
	Value string
}

// NewInvalidElementAttribute initializes a new `InvalidElementAttribute` from the given text
func NewInvalidElementAttribute(text []byte) (*InvalidElementAttribute, error) {
	value := string(text)
	log.Debugf("Initializing a new InvalidElementAttribute with text=%s", value)
	return &InvalidElementAttribute{Value: value}, nil
}

// ------------------------------------------
// StringElement
// ------------------------------------------

// StringElement the structure for strings
type StringElement struct {
	Content string
}

// NewStringElement initializes a new `StringElement` from the given content
func NewStringElement(content interface{}) *StringElement {
	return &StringElement{Content: content.(string)}
}

// Accept implements Visitable#Accept(Visitor)
func (s *StringElement) Accept(v Visitor) error {
	err := v.BeforeVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting string element")
	}
	err = v.Visit(s)
	if err != nil {
		return errors.Wrapf(err, "error while visiting string element")
	}
	err = v.AfterVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting string element")
	}
	return nil
}

// ------------------------------------------
// Quoted text
// ------------------------------------------

// QuotedText the structure for quoted text
type QuotedText struct {
	Kind     QuotedTextKind
	Elements []InlineElement
}

// QuotedTextKind the type for
type QuotedTextKind int

const (
	// Bold bold quoted text
	Bold QuotedTextKind = iota
	// Italic italic quoted text
	Italic
	// Monospace monospace quoted text
	Monospace
)

// NewQuotedText initializes a new `QuotedText` from the given kind and content
func NewQuotedText(kind QuotedTextKind, content []interface{}) (*QuotedText, error) {
	elements, err := toInlineElements(content)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize a new QuotedText")
	}
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Initialized a new QuotedText with %d elements:", len(elements))
		spew.Dump(elements)
	}
	return &QuotedText{Kind: kind, Elements: elements}, nil
}

// Accept implements Visitable#Accept(Visitor)
func (t *QuotedText) Accept(v Visitor) error {
	err := v.BeforeVisit(t)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting quoted text")
	}
	err = v.Visit(t)
	if err != nil {
		return errors.Wrapf(err, "error while visiting quoted text")
	}
	for _, element := range t.Elements {
		if visitable, ok := element.(Visitable); ok {
			err := visitable.Accept(v)
			if err != nil {
				return errors.Wrapf(err, "error while visiting quoted text element")
			}
		}
	}
	err = v.AfterVisit(t)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting quoted text")
	}
	return nil
}

// ------------------------------------------------------
// Escaped Quoted Text (i.e., with substitution prevention)
// ------------------------------------------------------

// NewEscapedQuotedText returns a new InlineContent where the nested elements are preserved (ie, substituted as expected)
func NewEscapedQuotedText(backslashes []interface{}, punctuation string, content []interface{}) (*InlineContent, error) {
	backslashesStr, err := stringify(backslashes,
		func(s string) (string, error) {
			// remove the number of back-slashes that match the length of the punctuation. Eg: `\*` or `\\**`, but keep extra back-slashes
			if len(s) > len(punctuation) {
				return s[len(punctuation):], nil
			}
			return "", nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "error while initializing quoted text with substitution prevention")
	}
	return NewInlineContent([]interface{}{backslashesStr, punctuation, content, punctuation})
}

// ------------------------------------------
// Passthrough
// ------------------------------------------

// Passthrough the structure for Passthroughs
type Passthrough struct {
	Kind     PassthroughKind
	Elements []InlineElement
}

// PassthroughKind the kind of passthrough
type PassthroughKind int

const (
	// SinglePlusPassthrough a passthrough with a single `+` punctuation
	SinglePlusPassthrough PassthroughKind = iota
	// TriplePlusPassthrough a passthrough with a triple `+++` punctuation
	TriplePlusPassthrough
	// PassthroughMacro a passthrough with the `pass:[]` macro
	PassthroughMacro
)

// NewPassthrough returns a new passthrough
func NewPassthrough(kind PassthroughKind, elements []interface{}) (*Passthrough, error) {
	passthroughElements, err := toInlineElements(elements)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize a new Passthrough")
	}
	log.Debugf("Initialized a new Passthrough with content: '%v'", passthroughElements)
	return &Passthrough{
		Kind:     kind,
		Elements: passthroughElements,
	}, nil

}

// ------------------------------------------
// BlankLine
// ------------------------------------------

// BlankLine the structure for the empty lines, which are used to separate logical blocks
type BlankLine struct {
}

// NewBlankLine initializes a new `BlankLine`
func NewBlankLine() (*BlankLine, error) {
	log.Debug("Initializing a new BlankLine")
	return &BlankLine{}, nil
}

// ------------------------------------------
// Links
// ------------------------------------------

// ExternalLink the structure for the external links
type ExternalLink struct {
	URL  string
	Text string
}

// NewExternalLink initializes a new `ExternalLink`
func NewExternalLink(url, text []interface{}) (*ExternalLink, error) {
	urlStr, err := stringify(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize a new ExternalLink element")
	}
	textStr, err := stringify(text, // remove "\n" or "\r\n", depending on the OS.
		// remove heading "[" and traingin "]"
		func(s string) (string, error) {
			return strings.TrimPrefix(s, "["), nil
		},
		func(s string) (string, error) {
			return strings.TrimSuffix(s, "]"), nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize a new ExternalLink element")
	}
	return &ExternalLink{URL: *urlStr, Text: *textStr}, nil
}