package sego

import (
	"fmt"
)

// 输出分词结果为字符串
//
// 有两种输出模式，以"中华人民共和国"为例
//
//  普通模式（searchMode=false）输出一个分词"中华人民共和国/ns "
//  搜索模式（searchMode=true） 输出普通模式的再细致切分：
//      "中华/nz 人民/n 共和/nz 共和国/ns 人民共和国/nt 中华人民共和国/ns "
//
// 搜索模式主要用于给搜索引擎提供尽可能多的关键字，详情请见Token结构体的注释。
func SegmentsToString(segs []Segment, searchMode bool) (output string) {
	if searchMode {
		for _, seg := range segs {
			output += tokenToString(seg.token)
		}
	} else {
		for _, seg := range segs {
			output += fmt.Sprintf(
				"%s/%s ", seg.token.text.String(), seg.token.pos)
		}
	}
	return
}

// func tokenToString(token *Token) (output string) {
// 	hasOnlyTerminalToken := true
// 	for _, s := range token.segments {
// 		if len(s.token.segments) > 1 {
// 			hasOnlyTerminalToken = false
// 		}
// 	}

// 	if !hasOnlyTerminalToken {
// 		for _, s := range token.segments {
// 			if s != nil {
// 				output += tokenToString(s.token)
// 			}
// 		}
// 	}
// 	output += fmt.Sprintf("%s/%s ", token.text.String(), token.pos)
// 	return
// }

// 输出分词结果到一个字符串slice
//
// 有两种输出模式，以"中华人民共和国"为例
//
//  普通模式（searchMode=false）输出一个分词"[中华人民共和国]"
//  搜索模式（searchMode=true） 输出普通模式的再细致切分：
//      "[中华 人民 共和 共和国 人民共和国 中华人民共和国]"
//
// 搜索模式主要用于给搜索引擎提供尽可能多的关键字，详情请见Token结构体的注释。

func SegmentsToSlice(segs []Segment, searchMode bool) (output []string) {
	if searchMode {
		for _, seg := range segs {
			output = append(output, tokenToSlice(seg.token)...)
		}
	} else {
		for _, seg := range segs {
			output = append(output, seg.token.Text())
		}
	}
	return
}

// func tokenToSlice(token *Token) (output []string) {
// 	hasOnlyTerminalToken := true
// 	for _, s := range token.segments {
// 		if len(s.token.segments) > 1 {
// 			hasOnlyTerminalToken = false
// 		}
// 	}
// 	if !hasOnlyTerminalToken {
// 		output = make([]string, 0, len(token.segments))
// 		for _, s := range token.segments {
// 			output = append(output, tokenToSlice(s.token)...)
// 		}
// 	}
// 	output = append(output, token.text.String())
// 	return output
// }
