//Go中文分词
package sego // import "code.sajari.com/sego"

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"code.sajari.com/sego/data"
)

// Only read participles greater than or equal to this frequency from the dictionary file
const minTokenFrequency = 2

// 分词器结构体
type Segmenter struct {
	dict *Dictionary
}

// 该结构体用于记录Viterbi算法中某字元处的向前分词跳转信息
type jumper struct {
	minDistance float32
	token       *Token
}

// Dictionary returns the dictionary
func (seg *Segmenter) Dictionary() *Dictionary {
	return seg.dict
}

// LoadDictionary loads a dictionary from a file
//
// Multiple dictionary files can be loaded, with filenames separated by ",".
// "User Dictionary.txt, Common Dictionary.txt"
// When a participle appears in both the user dictionary and the general dictionary, the user dictionary is used preferentially.
//
// The format of the dictionary is (one line per participle):
// Word segmentation text Frequency Part of speech
func (seg *Segmenter) LoadDictionary(files ...string) error {
	seg.dict = NewDictionary()
	for _, file := range files {
		dictFile, err := os.Open(file)
		defer dictFile.Close()
		if err != nil {
			return fmt.Errorf("could not open %q: %v", file, err)
		}
		seg.tokenizeDictionary(dictFile)
	}
	seg.processDictionary()

	return nil
}

// LoadDictionaryFromReader loads a dictionary from an io.Reader
//
// The format of the dictionary is (one line per participle):
// Word segmentation text Frequency Part of speech
func (seg *Segmenter) LoadDictionaryFromReader(r io.Reader) {
	seg.dict = NewDictionary()
	seg.tokenizeDictionary(r)
	seg.processDictionary()
}

// LoadDefaultDictionary loads the default dictionary stored in data
func (seg *Segmenter) LoadDefaultDictionary() {
	seg.LoadDictionaryFromReader(data.MustDictionary())
}

func (seg *Segmenter) tokenizeDictionary(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		if len(line) < 2 {
			// invalid line
			continue
		}

		text := line[0]
		freqText := line[1]
		pos := "" // part of speech tag
		if len(line) > 2 {
			pos = line[2]
		}

		// Analyze word frequency
		frequency, err := strconv.Atoi(freqText)
		if err != nil {
			continue
		}

		// Filter words that are too small
		if frequency < minTokenFrequency {
			continue
		}

		// Add participles to the dictionary
		words := splitTextToWords([]byte(text))
		token := Token{text: words, frequency: frequency, pos: pos}
		seg.dict.addToken(token)
	}
}

func (seg *Segmenter) processDictionary() {
	// Calculate the path value of each participle.
	// For the meaning of the path value, see the annotation of the Token structure
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	// Make a careful division of each participle for the search engine pattern.
	// For usage of this pattern, see Token structure comments.
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		segments := seg.segmentWords(token.text, true)

		// Calculate the number of subparticiples that need to be added
		numTokensToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				numTokensToAdd++
			}
		}
		token.segments = make([]*Segment, numTokensToAdd)

		// Add child segmentation
		iSegmentsToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				token.segments[iSegmentsToAdd] = &segments[iToken]
				iSegmentsToAdd++
			}
		}
	}
}

// DefaultSegmenter creates a new Segmenter with the default dictionary loaded
func DefaultSegmenter() *Segmenter {
	var seg Segmenter
	seg.LoadDefaultDictionary()
	return &seg
}

// 对文本分词
//
// 输入参数：
//	bytes	UTF8文本的字节数组
//
// 输出：
//	[]Segment	划分的分词
func (seg *Segmenter) Segment(bytes []byte) []Segment {
	return seg.internalSegment(bytes, false)
}

func (seg *Segmenter) InternalSegment(bytes []byte, searchMode bool) []Segment {
	return seg.internalSegment(bytes, searchMode)
}

func (seg *Segmenter) internalSegment(bytes []byte, searchMode bool) []Segment {
	// 处理特殊情况
	if len(bytes) == 0 {
		return []Segment{}
	}

	// 划分字元
	text := splitTextToWords(bytes)

	return seg.segmentWords(text, searchMode)
}

func (seg *Segmenter) segmentWords(text []Text, searchMode bool) []Segment {
	// 搜索模式下该分词已无继续划分可能的情况
	if searchMode && len(text) == 1 {
		return []Segment{}
	}

	// jumpers定义了每个字元处的向前跳转信息，包括这个跳转对应的分词，
	// 以及从文本段开始到该字元的最短路径值
	jumpers := make([]jumper, len(text))

	tokens := make([]*Token, seg.dict.maxTokenLength)
	for current := 0; current < len(text); current++ {
		// 找到前一个字元处的最短路径，以便计算后续路径值
		var baseDistance float32
		if current == 0 {
			// 当本字元在文本首部时，基础距离应该是零
			baseDistance = 0
		} else {
			baseDistance = jumpers[current-1].minDistance
		}

		// 寻找所有以当前字元开头的分词
		numTokens := seg.dict.lookupTokens(
			text[current:minInt(current+seg.dict.maxTokenLength, len(text))], tokens)

		// 对所有可能的分词，更新分词结束字元处的跳转信息
		for iToken := 0; iToken < numTokens; iToken++ {
			location := current + len(tokens[iToken].text) - 1
			if !searchMode || current != 0 || location != len(text)-1 {
				updateJumper(&jumpers[location], baseDistance, tokens[iToken])
			}
		}

		// 当前字元没有对应分词时补加一个伪分词
		if numTokens == 0 || len(tokens[0].text) > 1 {
			updateJumper(&jumpers[current], baseDistance,
				&Token{text: []Text{text[current]}, frequency: 1, distance: 32, pos: "x"})
		}
	}

	// 从后向前扫描第一遍得到需要添加的分词数目
	numSeg := 0
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg++
		index = location - 1
	}

	// 从后向前扫描第二遍添加分词到最终结果
	outputSegments := make([]Segment, numSeg)
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg--
		outputSegments[numSeg].token = jumpers[index].token
		index = location - 1
	}

	// 计算各个分词的字节位置
	bytePosition := 0
	for iSeg := 0; iSeg < len(outputSegments); iSeg++ {
		outputSegments[iSeg].start = bytePosition
		bytePosition += textSliceByteLength(outputSegments[iSeg].token.text)
		outputSegments[iSeg].end = bytePosition
	}
	return outputSegments
}

// 更新跳转信息:
// 	1. 当该位置从未被访问过时(jumper.minDistance为零的情况)，或者
//	2. 当该位置的当前最短路径大于新的最短路径时
// 将当前位置的最短路径值更新为baseDistance加上新分词的概率
func updateJumper(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 || jumper.minDistance > newDistance {
		jumper.minDistance = newDistance
		jumper.token = token
	}
}

// 取两整数较小值
func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// 取两整数较大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 将文本划分成字元
func splitTextToWords(text Text) []Text {
	output := make([]Text, 0, len(text)/3)
	current := 0
	inAlphanumeric := true
	alphanumericStart := 0
	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
			// 当前是拉丁字母或数字（非中日韩文字）
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 {
					output = append(output, toLower(text[alphanumericStart:current]))
				}
			}
			output = append(output, text[current:current+size])
		}
		current += size
	}

	// 处理最后一个字元是英文的情况
	if inAlphanumeric {
		if current != 0 {
			output = append(output, toLower(text[alphanumericStart:current]))
		}
	}

	return output
}

// 将英文词转化为小写
func toLower(text []byte) []byte {
	output := make([]byte, len(text))
	for i, t := range text {
		if t >= 'A' && t <= 'Z' {
			output[i] = t - 'A' + 'a'
		} else {
			output[i] = t
		}
	}
	return output
}
