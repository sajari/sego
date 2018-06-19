package sego

import (
	"fmt"
	"testing"

	"github.com/issue9/assert"
)

var (
	strs = []Text{
		Text("one"),
		Text("two"),
		Text("three"),
		Text("four"),
		Text("five"),
		Text("six"),
		Text("seven"),
		Text("eight"),
		Text("nine"),
		Text("ten"),
	}
)

func BenchmarkStringsJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Join(strs)
	}
}

func Test_Token_TextEquals(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("one"),
			[]byte("two"),
		},
	}
	assert.True(t, token.TextEquals("onetwo"))
}

func Test_Token_TextEquals_CN(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("中国"),
			[]byte("文字"),
		},
	}
	assert.True(t, token.TextEquals("中国文字"))
}

func Test_Token_TextNotEquals(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("one"),
			[]byte("two"),
		},
	}
	assert.False(t, token.TextEquals("one-two"))
}

func Test_Token_TextNotEquals_CN(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("中国"),
			[]byte("文字"),
		},
	}
	assert.False(t, token.TextEquals("中国文字1"))
}

func Test_Token_TextNotEquals_CN_B(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("中国"),
			[]byte("文字"),
		},
	}
	assert.False(t, token.TextEquals("中国文"))
}

func Test_Token_Split(t *testing.T) {
	probMap := map[string]string{
		"衣门襟":    "拉链",
		"品牌":     "天奕",
		"图案":     "纯色 字母",
		"颜色分类":   "牛奶白 水粉色 湖水蓝 浅军绿 雅致灰",
		"尺码":     "大码XL 大码XXL 大码XXXL 大码XXXXL",
		"组合形式":   "单件",
		"面料":     "聚酯",
		"领型":     "连帽",
		"服饰工艺":   "立体裁剪",
		"货号":     "YZL-1806052",
		"厚薄":     "超薄",
		"年份季节":   "2018年夏季",
		"通勤":     "韩版",
		"服装款式细节": "不对称",
		"成分含量":   "81%(含)-90%(含)",
		"袖型":     "常规",
		"风格":     "通勤",
		"适用年龄":   "18-24周岁",
		"服装版型":   "宽松",
		"大码女装分类": "其它特大款式",
		"衣长":     "中长款",
		"袖长":     "长袖",
		"穿着方式":   "开衫",
	}
	word := "卫衣女宽松拉链外套开衫韩版"
	var segmenter Segmenter
	segmenter.LoadDictionary("dictionary.txt")
	segments := segmenter.InternalSegment([]byte(word), true)
	// for _, s := range segments {
	// 	fmt.Println(s.token.Text())
	// }
	for _, value := range probMap {
		for _, s := range segments {
			if s.Token().Text() == value {
				fmt.Println("=", value)
			}
		}
	}
}
