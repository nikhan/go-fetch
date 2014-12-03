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
		exp: `.[""]`,
	},
	FailTest{
		exp: `.[x]`,
	},
	FailTest{
		exp: `.[?]`,
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

	q := `.['arrayObj'][2]['nested'][0]['id']`
	l, err := Parse(q)
	if err != nil {
		t.Fail()
	}
	if l.String() != q {
		t.Fail()
		fmt.Println("Expected Value", q)
	}
	fmt.Println("\x1b[32;1m✓\x1b[0m", "String()")

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
		_, err := Run(l, umsg)
		if err != nil {
			fmt.Println(".")
		}
	}
}

func BenchmarkNoFetch(b *testing.B) {
	var umsg interface{}
	testFile, _ := ioutil.ReadFile("test.json")
	json.Unmarshal(testFile, &umsg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o, ok := umsg.(map[string]interface{})
		if !ok {
			fmt.Println(".")
		}
		f, ok := o["arrayObj"]
		if !ok {
			fmt.Println(".")
		}
		d, ok := f.([]interface{})
		if !ok {
			fmt.Println(".")
		}
		if 2 > len(d) {
			fmt.Println(".")
		}
		s := d[2]
		a, ok := s.(map[string]interface{})
		if !ok {
			fmt.Println(".")
		}
		z, ok := a["nested"]
		if !ok {
			fmt.Println(".")
		}
		x, ok := z.([]interface{})
		if !ok {
			fmt.Println(".")
		}
		if 0 > len(x) {
			fmt.Println(".")
		}
		c := x[0]
		v, ok := c.(map[string]interface{})
		if !ok {
			fmt.Println(".")
		}
		_, ok = v["id"]
		if !ok {
			fmt.Println(".")
		}
	}
}

func BenchmarkNoFetchNoCheck(b *testing.B) {
	testFile, _ := ioutil.ReadFile("test.json")

	var umsg struct {
		A struct {
			B struct {
				C []struct {
					D struct {
						E int64 `json:"e"`
					} `json:"d"`
				} `json:"c"`
			} `json:"b"`
		} `json:"a"`
		ArrayFloat []interface{} `json:"arrayFloat"`
		ArrayInt   []int64       `json:"arrayInt"`
		ArrayObj   []struct {
			Array  []int64 `json:"array"`
			Bool   bool    `json:"bool"`
			HasKey bool    `json:"hasKey"`
			Name   string  `json:"name"`
			Nested []struct {
				Id string `json:"id"`
				No string `json:"no"`
			} `json:"nested"`
			Nil     interface{} `json:"nil"`
			SameNum int64       `json:"sameNum"`
			SameStr string      `json:"sameStr"`
			Val     interface{} `json:"val"`
		} `json:"arrayObj"`
		ArrayString []string      `json:"arrayString"`
		Bool        bool          `json:"bool"`
		Empty       []interface{} `json:"empty"`
		Escapekey   struct {
			Nested struct {
				Foobar string `json:"foo.bar"`
			} `json:"nested"`
		} `json:"escape.key"`
		Float    float64 `json:"float"`
		FloatStr string  `json:"float_str"`
		Int      int64   `json:"int"`
		K        int64   `json:"#_k__"`
		Nested   struct {
			Baz []int64 `json:"baz"`
			Foo struct {
				Zip string `json:"zip"`
			} `json:"foo"`
		} `json:"nested"`
		Nil    interface{} `json:"nil"`
		String string      `json:"string"`
		Two    int64       `json:"two"`
	}

	json.Unmarshal(testFile, &umsg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o := umsg.ArrayObj[2].Nested[0].Id
		if o != "" {
		}
	}
}
