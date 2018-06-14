package sego

import (
	"fmt"
	"testing"
)

func expect(t *testing.T, expect string, actual interface{}) {
	actualString := fmt.Sprint(actual)
	if expect != actualString {
		t.Errorf("Expected value = \"%s\", Actual =\"%s\"", expect, actualString)
	}
}

func printTokens(tokens []*Token, numTokens int) (output string) {
	for iToken := 0; iToken < numTokens; iToken++ {
		for _, word := range tokens[iToken].text {
			output += fmt.Sprint(string(word))
		}
		output += " "
	}
	return
}

func bytesToString(bytes Text) (output string) {
	for _, b := range bytes {
		output += (string(b) + "/")
	}
	return
}
