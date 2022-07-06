package main

import (
	"strconv"
	"testing"
)

func Test(t *testing.T) {
	testCases := []struct {
		small  string
		normal string
	}{
		{
			small:  "-2.918-4",
			normal: "-2.918e-4",
		},
	}
	for _, tC := range testCases {
		got, err := parseSmallFortranFloat(tC.small)
		if err != nil {
			t.Fatal(err)
		}
		expect, err := strconv.ParseFloat(tC.normal, 32)
		if got != expect {
			t.Errorf("got different expect,got, %g , %g", got, expect)
		}
	}
}
