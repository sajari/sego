package sego

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
)

// Only read participles greater than or equal to this frequency from the dictionary file
const minTokenFrequency = 2

// Mode represent the word segmentation mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
)

// Segmenter represents the wordbreaker structure
type Segmenter struct {
	mode Mode
	dict *Dictionary
}

func NewSegmenter(mode Mode) *Segmenter {
	return &Segmenter{
		dict: NewDictionary(),
		mode: mode,
	}
}

// This structure is used to record the forward segmentation jump
// information at a certain character in the Viterbi algorithm
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
func (seg *Segmenter) LoadDictionary(files string) error {
	for _, file := range strings.Split(files, ",") {
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
	seg.tokenizeDictionary(r)
	seg.processDictionary()
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
	if seg.mode == ModeSearch {
		for i := range seg.dict.tokens {
			token := &seg.dict.tokens[i]
			segments := seg.segmentWords(token.text)

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
}

// Segment returns the word segmentation
func (seg *Segmenter) Segment(bytes []byte) []Segment {
	if len(bytes) == 0 {
		return nil
	}

	return seg.segmentWords(splitTextToWords(bytes))
}

func (seg *Segmenter) segmentWords(text Text) []Segment {
	// The word segmentation has no further division in the search mode.
	if seg.mode == ModeSearch && len(text) == 1 {
		return nil
	}

	// The jumpers define the forward jump information at each character, including the participle of the jump.
	// and the shortest path value starting from the text segment to this character
	jumpers := make([]jumper, len(text))

	tokens := make([]*Token, seg.dict.maxTokenLength)
	for current := 0; current < len(text); current++ {
		// Find the shortest path at the previous character to calculate the subsequent path value
		var baseDistance float32
		if current != 0 {
			baseDistance = jumpers[current-1].minDistance
		}

		// Find all participles that start with the current character
		numTokens := seg.dict.lookupTokens(
			text[current:minInt(current+seg.dict.maxTokenLength, len(text))], tokens)

		// Update skip information at the end of the word segmentation for all possible participles
		for iToken := 0; iToken < numTokens; iToken++ {
			location := current + len(tokens[iToken].text) - 1
			if seg.mode != ModeSearch || current != 0 || location != len(text)-1 {
				updateJumper(&jumpers[location], baseDistance, tokens[iToken])
			}
		}

		// If the current character does not have a corresponding participle, add a pseudo-participle
		if numTokens == 0 || len(tokens[0].text) > 1 {
			updateJumper(&jumpers[current], baseDistance,
				&Token{text: Text{text[current]}, frequency: 1, distance: 32, pos: "x"})
		}
	}

	// Scan the first pass from back to front to get the number of participles to add
	numSeg := 0
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg++
		index = location - 1
	}

	// Scan the second pass from back to front to add participles to the final result
	outputSegments := make([]Segment, numSeg)
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg--
		outputSegments[numSeg].token = jumpers[index].token
		index = location - 1
	}

	// Calculate the byte position of each participle
	bytePosition := 0
	for iSeg := 0; iSeg < len(outputSegments); iSeg++ {
		outputSegments[iSeg].start = bytePosition
		bytePosition += outputSegments[iSeg].token.text.Size()
		outputSegments[iSeg].end = bytePosition
	}
	return outputSegments
}

// Update the jump information:
// 1. When the location has never been accessed (jumper.minDistance is zero), or
// 2. When the current shortest path at this location is greater than the new shortest path
// Update the shortest path value of the current position to the probability of baseDistance plus the new participle
func updateJumper(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 || jumper.minDistance > newDistance {
		jumper.minDistance = newDistance
		jumper.token = token
	}
}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// splitTextToWords divides the text into characters
func splitTextToWords(text []byte) Text {
	output := make(Text, 0, len(text)/3)

	var alphaStart int
	var inAlpha bool
	for current := 0; current < len(text); {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
			if !inAlpha {
				alphaStart = current
				inAlpha = true
			}

			current += size
			continue
		}

		if inAlpha {
			output = append(output, toLower(text[alphaStart:current]))
			inAlpha = false
		}

		output = append(output, text[current:current+size])
		current += size
	}

	if inAlpha {
		output = append(output, toLower(text[alphaStart:]))
	}

	return output
}

// toLower converts english words to lowercase
func toLower(text []byte) []byte {
	output := text[:0]
	for _, t := range text {
		if t >= 'A' && t <= 'Z' {
			output = append(output, t-'A'+'a')
		} else {
			output = append(output, t)
		}
	}
	return output
}
