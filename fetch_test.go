package fetch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

type Test struct {
	exp    string
	result string
}

type DotTest struct {
	input  string
	exp    string
	result string
}

type FailTest struct {
	exp string
}

var Tests = []Test{
	Test{
		exp:    `.["#_k__"]`,
		result: `1`,
	},
	Test{
		exp:    `.two`,
		result: `2`,
	},
	Test{
		exp:    `.arrayObj[0].name`,
		result: `"foo"`,
	},
	Test{
		exp:    `.a.b.c[1].d.e`,
		result: `1`,
	},
	Test{
		exp:    `.["a"]["b"]["c"][1]["d"]["e"]`,
		result: `1`,
	},
	Test{
		exp:    `.a["b"].c[1].d["e"]`,
		result: `1`,
	},
	Test{
		exp:    `.arrayInt`,
		result: `[1,2,3,4,5,6,7,8,9,10]`,
	},
	Test{
		exp:    `.arrayInt[2]`,
		result: `3`,
	},
	Test{
		exp:    `.arrayString`,
		result: `["yellow", "purple","red", "green"]`,
	},
	Test{
		exp:    `.['escape.key']`,
		result: `{"nested":{"foo.bar":"baz"}}`,
	},
	Test{
		exp:    `.['escape.key']['nested']`,
		result: `{"foo.bar":"baz"}`,
	},
	Test{
		exp:    `.['escape.key']['nested']['foo.bar']`,
		result: `"baz"`,
	},
	Test{
		exp:    `.['escape.key'].nested["foo.bar"]`,
		result: `"baz"`,
	},
}

var DotTests = []DotTest{
	DotTest{
		input:  `true`,
		exp:    `.`,
		result: `true`,
	},
	DotTest{
		input:  `false`,
		exp:    `.`,
		result: `false`,
	},
	DotTest{
		input:  `null`,
		exp:    `.`,
		result: `null`,
	},
	DotTest{
		input:  `100`,
		exp:    `.`,
		result: `100`,
	},
	DotTest{
		input:  `"hello world"`,
		exp:    `.`,
		result: `"hello world"`,
	},
	DotTest{
		input:  `[1,2,3,4,5]`,
		exp:    `.`,
		result: `[1,2,3,4,5]`,
	},
	DotTest{
		input:  `[1,2,3,4,5]`,
		exp:    `.[0]`,
		result: `1`,
	},
	DotTest{
		input:  `[0,null,true,"hello"]`,
		exp:    `.[0]`,
		result: `0`,
	},
	DotTest{
		input:  `[0,null,true,"hello"]`,
		exp:    `.[1]`,
		result: `null`,
	},
	DotTest{
		input:  `[0,null,true,"hello"]`,
		exp:    `.[2]`,
		result: `true`,
	},
	DotTest{
		input:  `[0,null,true,"hello"]`,
		exp:    `.[3]`,
		result: `"hello"`,
	},
}

var FailTests = []FailTest{
	FailTest{
		exp: ` .`,
	},
	FailTest{
		exp: `. `,
	},
	FailTest{
		exp: `.missingkey`,
	},
	FailTest{
		exp: `.#_k__`,
	},
	FailTest{
		exp: `.[0]`,
	},
	FailTest{
		exp: `.["arrayObj`,
	},
	FailTest{
		exp: `.arrayObj"]`,
	},
	FailTest{
		exp: `.arrayObj]`,
	},
	FailTest{
		exp: `.[]`,
	},
	FailTest{
		exp: `.[[][]]`,
	},
	FailTest{
		exp: `"jdsjdskdjsjs`,
	},
	FailTest{
		exp: `?!.foo`,
	},
	FailTest{
		exp: `.['escape.key']['nested']["foo.bar']`,
	},
	FailTest{
		exp: `.['escape.key']['nested'].["foo.bar"]`,
	},
	FailTest{
		exp: `...`,
	},
	FailTest{
		exp: `.['escape.key'].['nested'].["foo.bar"]`,
	},
	FailTest{
		exp: `.['escape.key']['nested'].["foo.bar"].`,
	},
	FailTest{
		exp: `.['arrayString' arrayString].`,
	},
	FailTest{
		exp: `.['arrayString' 'arrayString'].`,
	},
	FailTest{
		exp: `.['arrayString''arrayString'].`,
	},
	FailTest{
		exp: `.['arrayString'2].`,
	},
	FailTest{
		exp: `.[22!2].`,
	},
	FailTest{
		exp: `.[['arrayString']].`,
	},
}

func TestAll(t *testing.T) {
	var m interface{}
	var r interface{}

	testFile, _ := ioutil.ReadFile("test.json")
	json.Unmarshal(testFile, &m)

	for _, test := range Tests {
		err := json.Unmarshal([]byte(test.result), &r)
		if err != nil {
			t.Error(err, "bad test")
		}

		result, err := Fetch(test.exp, m)
		if err != nil {
			t.Error("failed Fetch")
		}

		if reflect.DeepEqual(r, result) {
			fmt.Println("\x1b[32;1m✓\x1b[0m", test.exp)
		} else {
			t.Fail()
			fmt.Println("\x1b[31;1m", "✕", "\x1b[0m", r, "\t", result)
			fmt.Println("Expected Value", r, "\tResult Value:", result)
		}
	}

	for _, test := range DotTests {
		var inputJSON interface{}
		var expected interface{}
		json.Unmarshal([]byte(test.input), &inputJSON)
		json.Unmarshal([]byte(test.result), &expected)
		result, err := Fetch(test.exp, inputJSON)
		if err != nil {
			t.Error("failed Fetch")
		}
		if reflect.DeepEqual(expected, result) {
			fmt.Println("\x1b[32;1m✓\x1b[0m", test.exp)
		} else {
			t.Fail()
			fmt.Println("\x1b[31;1m", "✕", "\x1b[0m", expected, "\t", result)
			fmt.Println("Expected Value", expected, "\tResult Value:", result)
		}
	}

	for _, test := range FailTests {
		_, err := Fetch(test.exp, m)
		if err != nil {
			fmt.Println("\x1b[32;1m✓\x1b[0m", test.exp)
			fmt.Println(err.Error())
		} else {
			t.Fail()
			fmt.Println("\x1b[31;1m", "✕", "\x1b[0m\t", test.exp)
			fmt.Println("\tExpression Value:", test.exp)
		}
	}

}

func BenchmarkFetch(b *testing.B) {
	var umsg interface{}
	testFile, _ := ioutil.ReadFile("test.json")
	json.Unmarshal(testFile, &umsg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Fetch(`.['arrayObj'][2]['nested'][0]['id']`, umsg)
	}
}

func BenchmarkFetchParseOnce(b *testing.B) {
	var umsg interface{}
	testFile, _ := ioutil.ReadFile("test.json")
	json.Unmarshal(testFile, &umsg)
	l, _ := Parse(`.['arrayObj'][2]['nested'][0]['id']`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Run(l, umsg)
	}
}

func BenchmarkNoFetch(b *testing.B) {
	var umsg interface{}
	testFile, _ := ioutil.ReadFile("test.json")
	json.Unmarshal(testFile, &umsg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o, _ := umsg.(map[string]interface{})
		f, _ := o["arrayObj"]
		d, _ := f.([]interface{})
		s := d[2]
		a, _ := s.(map[string]interface{})
		z, _ := a["nested"]
		x, _ := z.([]interface{})
		c := x[0]
		v, _ := c.(map[string]interface{})
		_, ok := v["id"]
		if !ok {
		}
	}
}
