package types

import (
	"bytes"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func toInlineElements(elements []interface{}) ([]InlineElement, error) {
	mergedElements := merge(elements)
	result := make([]InlineElement, len(mergedElements))
	for i, element := range mergedElements {
		switch element := element.(type) {
		case InlineElement:
			result[i] = element
		default:
			return nil, errors.Errorf("unexpected element of type: %T (expected a InlineElement instead)", element)
		}
	}
	return result, nil
}

// filterUnrelevantElements excludes the unrelevant (empty) blocks
func filterUnrelevantElements(blocks []interface{}) []DocElement {
	log.Debugf("Filtering %d blocks...", len(blocks))
	elements := make([]DocElement, 0)
	for _, block := range blocks {
		log.Debugf(" converting block of type '%T' into a DocElement...", block)
		switch block := block.(type) {
		case *BlankLine:
			// exclude blank lines from here, we won't need them in the rendering anyways
		case *Preamble:
			// exclude empty preambles
			if len(block.Elements) > 0 {
				// exclude empty preamble
				elements = append(elements, block)
			}
		case []interface{}:
			result := filterUnrelevantElements(block)
			elements = append(elements, result...)
		default:
			if block != nil {
				elements = append(elements, block)
			}
		}
	}
	return elements
}

func merge(elements []interface{}, extraElements ...interface{}) []interface{} {
	result := make([]interface{}, 0)
	allElements := append(elements, extraElements...)
	// log.Debugf("Merging %d element(s):", len(allElements))
	buff := bytes.NewBuffer(nil)
	for _, element := range allElements {
		if element == nil {
			continue
		}
		switch element := element.(type) {
		case string:
			buff.WriteString(element)
		case *string:
			buff.WriteString(*element)
		case []byte:
			for _, b := range element {
				buff.WriteByte(b)
			}
		case *StringElement:
			content := element.Content
			buff.WriteString(content)
		case *InlineContent:
			inlineElements := make([]interface{}, len(element.Elements))
			for i, e := range element.Elements {
				inlineElements[i] = e
			}
			result = merge(result, inlineElements...)
		case []interface{}:
			if len(element) > 0 {
				f := merge(element)
				result, buff = appendBuffer(result, buff)
				result = merge(result, f...)
			}
		default:
			log.Debugf("Merging with 'default' case an element of type %[1]T", element)
			result, buff = appendBuffer(result, buff)
			result = append(result, element)
		}
	}
	// if buff was filled because some text was found
	result, buff = appendBuffer(result, buff)
	// if len(extraElements) > 0 {
	// 	log.Debugf("merged '%v' (len=%d) with '%v' (len=%d) -> '%v' (len=%d)", elements, len(elements), extraElements, len(extraElements), result, len(result))

	// } else {
	// 	log.Debugf("merged '%v' (len=%d) -> '%v' (len=%d)", elements, len(elements), result, len(result))
	// }
	return result
}

// appendBuffer appends the content of the given buffer to the given array of elements,
// and returns a new buffer, or returns the given arguments if the buffer was empty
func appendBuffer(elements []interface{}, buff *bytes.Buffer) ([]interface{}, *bytes.Buffer) {
	if buff.Len() > 0 {
		return append(elements, NewStringElement(buff.String())), bytes.NewBuffer(nil)
	}
	return elements, buff
}

// stringifyOption a function to apply on the result of the `stringify` function below, before returning
type stringifyOption func(s string) (string, error)

// stringify convert the given elements into a string, then applies the optional `funcs` to convert the string before returning it.
// These StringifyFuncs can be used to trim the content, for example
func stringify(elements []interface{}, options ...stringifyOption) (*string, error) {
	mergedElements := merge(elements)
	b := make([]byte, 0)
	buff := bytes.NewBuffer(b)
	for _, element := range mergedElements {
		switch element := element.(type) {
		case *StringElement:
			buff.WriteString(element.Content)
		case []interface{}:
			stringifiedElement, err := stringify(element)
			if err != nil {
				// no need to wrap the error again in the same function
				return nil, err
			}
			buff.WriteString(*stringifiedElement)
		default:
			return nil, errors.Errorf("cannot convert element of type '%T' to string content", element)
		}

	}
	result := buff.String()
	for _, f := range options {
		var err error
		result, err = f(result)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to postprocess the stringified content")
		}
	}
	// log.Debugf("stringified %v -> '%s' (%v characters)", elements, result, len(result))
	return &result, nil
}