package sego

import (
	"bufio"
	"os"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		in, expected string
	}{
		{in: "中国有十三亿人口", expected: "中/国/有/十/三/亿/人/口/"},
		{in: "GitHub is a web-based hosting service, for software development projects.", expected: "github/ /is/ /a/ /web/-/based/ /hosting/ /service/,/ /for/ /software/ /development/ /projects/./"},
		{in: "中国雅虎Yahoo! China致力于，领先的公益民生门户网站。", expected: "中/国/雅/虎/yahoo/!/ /china/致/力/于/，/领/先/的/公/益/民/生/门/户/网/站/。/"},
		{in: "こんにちは", expected: "こ/ん/に/ち/は/"},
		{in: "안녕하세요", expected: "안/녕/하/세/요/"},
		{in: "Я тоже рада Вас видеть", expected: "Я/ /тоже/ /рада/ /Вас/ /видеть/"},
		{in: "¿Cómo van las cosas", expected: "¿/cómo/ /van/ /las/ /cosas/"},
		{in: "Wie geht es Ihnen", expected: "wie/ /geht/ /es/ /ihnen/"},
		{in: "Je suis enchanté de cette pièce", expected: "je/ /suis/ /enchanté/ /de/ /cette/ /pièce/"},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			actual := bytesToString(splitTextToWords([]byte(
				test.in)))

			if actual != test.expected {
				t.Errorf("expected %q, actual %q", test.expected, actual)
			}
		})
	}
}

func TestSegment(t *testing.T) {
	seg := NewSegmenter(ModeNormal)
	if err := seg.LoadDictionary("testdata/test_dict1.txt,testdata/test_dict2.txt"); err != nil {
		t.Fatal(err)
	}

	expect(t, "12", seg.dict.NumTokens())
	segments := seg.Segment([]byte("中国有十三亿人口"))
	expect(t, "中国/ 有/p3 十三亿/ 人口/p12 ", SegmentsToString(segments, false))
	expect(t, "4", len(segments))
	expect(t, "0", segments[0].start)
	expect(t, "6", segments[0].end)
	expect(t, "6", segments[1].start)
	expect(t, "9", segments[1].end)
	expect(t, "9", segments[2].start)
	expect(t, "18", segments[2].end)
	expect(t, "18", segments[3].start)
	expect(t, "24", segments[3].end)
}

var segments []Segment

func BenchmarkSegment(b *testing.B) {
	seg := NewSegmenter(ModeNormal)
	if err := seg.LoadDictionary("data/dictionary.txt"); err != nil {
		b.Fatal(err)
	}

	file, err := os.Open("testdata/bailuyuan.txt")
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	size := 0
	lines := [][]byte{}
	for scanner.Scan() {
		text := []byte(scanner.Text())
		size += len(text)
		lines = append(lines, text)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			segments = seg.Segment(l)
		}
	}
}

// func TestLargeDictionary(t *testing.T) {
// 	seg := NewSegmenter(ModeNormal)
// 	if err := seg.LoadDictionary("data/dictionary.txt"); err != nil {
// 		t.Fatal(err)
// 	}

// 	expect(t, "中国/ns 人口/n ", SegmentsToString(seg.Segment(
// 		[]byte("中国人口")), false))

// 	expect(t, "中国/ns 人口/n ", SegmentsToString(seg.internalSegment(
// 		[]byte("中国人口"), false), false))

// 	expect(t, "中国/ns 人口/n ", SegmentsToString(seg.internalSegment(
// 		[]byte("中国人口"), true), false))

// 	expect(t, "中华人民共和国/ns 中央人民政府/nt ", SegmentsToString(seg.internalSegment(
// 		[]byte("中华人民共和国中央人民政府"), true), false))

// 	expect(t, "中华人民共和国中央人民政府/nt ", SegmentsToString(seg.internalSegment(
// 		[]byte("中华人民共和国中央人民政府"), false), false))

// 	expect(t, "中华/nz 人民/n 共和/nz 国/n 共和国/ns 人民共和国/nt 中华人民共和国/ns 中央/n 人民/n 政府/n 人民政府/nt 中央人民政府/nt 中华人民共和国中央人民政府/nt ", SegmentsToString(seg.Segment(
// 		[]byte("中华人民共和国中央人民政府")), true))
// }
