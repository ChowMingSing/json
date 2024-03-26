package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	s := `{
	"k1": "abc",
		"k2": 123,
		"k3": -45.67,
		"k4": [333, 444],
		"k5": {
			"k8": "zxcv"
		},
		"k6": null,
		"k7": true
	}`

	var m = map[string]any{}
	//var m []any
	e := json.Unmarshal([]byte(s), &m)
	if e != nil {
		fmt.Println("err:", e)
		return
	}
	fmt.Println(m)
	return

}
