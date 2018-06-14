package main

import (
	"flag"
	"fmt"

	"github.com/benhinchley/sego"
)

var (
	text = flag.String("text", "中国互联网历史上最大的一笔并购案", "The text to be segmented")
)

func main() {
	flag.Parse()

	var seg sego.Segmenter
	seg.LoadDictionary("../../data/dictionary.txt")

	segments := seg.Segment([]byte(*text))
	fmt.Println(sego.SegmentsToString(segments, true))
}
