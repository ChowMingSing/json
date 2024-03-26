package main

import (
	"testing"
)

func Test1(t *testing.T) {
	s := `  {  }  `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test2(t *testing.T) {
	s := ` [  ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test3(t *testing.T) {
	s := `  {
	"k1": "abc",
	"k2": 123,
	"k3": -45.67,
	"k4": [333, [], -444.5, {}, "432", "das", true, {}, null, false],
	"k5": {
		"k8": "zxcv",
		"k9": ["nulllll"]
	},
	"k6": null,
	"k7": true
} `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test4(t *testing.T) {
	s := ` [ {  } , true,false, "  aa" ,null ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test5True(t *testing.T) {
	s := ` [ true ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test6False(t *testing.T) {
	s := ` [ false ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test7Null(t *testing.T) {
	s := ` [ null ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test8String(t *testing.T) {
	s := ` [ " das  true dqc " ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func Test9Number(t *testing.T) {
	s := ` [ -123.45 , 0 ,45.7  ] `
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}

func TestTemp(t *testing.T) {
	s := ` 

{
    "status": "1",
    "count": "1",
    "info": "OK",
    "infocode": "10000",
    "lives": [
        {
            "province": "北京",
            "city": "北京市",
            "adcode": "110000",
            "weather": "阴",
            "temperature": "11",
            "winddirection": "东",
            "windpower": "≤3",
            "humidity": "51",
            "reporttime": "2024-03-24 02:03:13",
            "temperature_float": "11.0",
            "humidity_float": "51.0"
        }
    ]
}

`
	err := CheckValid(s)
	if err != nil {
		t.Error(err)
	}
}
