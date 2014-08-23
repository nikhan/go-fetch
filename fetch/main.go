package main

import (
	"encoding/json"
	"fmt"
	"github.com/nikhan/go-fetch"
	"io/ioutil"
	"os"
)

var info = `fetch 0.1.0`

func main() {
	var obj interface{}

	if len(os.Args) != 2 {
		fmt.Println(info)
		os.Exit(1)
	}

	j, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = json.Unmarshal(j, &obj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	result, err := fetch.Fetch(string(os.Args[1]), obj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(r))
}
