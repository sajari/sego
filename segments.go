package sego

import (
	"fmt"
	"strings"
)

type Segments []*Segment

func (s Segments) Size() int {
	return countSegments(s)
}

func countSegments(s Segments) (count int) {
	count += len(s)
	for _, seg := range s {
		count += countSegments(seg.token.segments)
	}
	return count
}

func (s Segments) Slice(mode Mode) (output []string) {
	output = make([]string, 0, s.Size())

	switch mode {
	case ModeSearch:
		for _, seg := range s {
			output = append(output, tokenToSlice(seg.token)...)
		}
	default:
		for _, seg := range s {
			output = append(output, seg.token.Text())
		}
	}

	return output
}

func tokenToSlice(token *Token) (output []string) {
	if token.segments.Size() > 1 {
		output = make([]string, 0, len(token.segments))
		for _, s := range token.segments {
			output = append(output, tokenToSlice(s.token)...)
		}
		return output
	}

	output = append(output, token.text.String())
	return output
}

func (s Segments) String(mode Mode) string {
	var str strings.Builder
	str.Grow(len(s))

	switch mode {
	case ModeSearch:
		for _, seg := range s {
			str.WriteString(tokenToString(seg.token))
		}
	default:
		for _, seg := range s {
			str.WriteString(fmt.Sprintf(
				"%s/%s ", seg.token.text.String(), seg.token.pos))
		}

	}

	return str.String()
}

func tokenToString(token *Token) (output string) {
	if token.segments.Size() > 1 {
		var str strings.Builder
		str.Grow(len(token.segments))

		for _, s := range token.segments {
			if s != nil {
				str.WriteString(tokenToString(s.token))
			}
		}

		return str.String()
	}

	return fmt.Sprintf("%s/%s ", token.text.String(), token.pos)
}
