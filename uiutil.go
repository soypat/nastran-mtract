package main

import (
	ui "github.com/gizak/termui/v3"
	//"github.com/gizak/termui/v3/widgets"
)

type fitting struct {
	// This stuff aint as stuffy as it looks
	widthStart  [3]int // 3 unit vector used to position each of the 4 window borders
	heightStart [3]int // first two units are used to indicate fraction of window. Last one is a constant character postion. basically y=m*x+b
	widthEnd    [3]int // [1, 2, 3] means the border is positioned 3 character lengths after halfway mark of terminal
	heightEnd   [3]int // [0, 5, 6] means the border is positioned 6 characters where leftwise terminal border begins. 5 is meaningless
}

func CreateFitting(wS [3]int, hS [3]int, wE [3]int, hE [3]int) *fitting {
	var P fitting
	if wS[1] == 0 {
		wS[1] = 1
	}
	if wE[1] == 0 {
		wE[1] = 1
	}
	if hS[1] == 0 {
		hS[1] = 1
	}
	if hE[1] == 0 {
		hE[1] = 1
	}
	P.widthStart = wS
	P.heightStart = hS
	P.widthEnd = wE
	P.heightEnd = hE

	return &P
}

func (P fitting) getRect() (int, int, int, int) {
	width, height := ui.TerminalDimensions()
	return width*P.widthStart[0]/P.widthStart[1] + P.widthStart[2], height*P.heightStart[0]/P.heightStart[1] + P.heightStart[2], width*P.widthEnd[0]/P.widthEnd[1] + P.widthEnd[2], height*P.heightEnd[0]/P.heightEnd[1] + P.heightEnd[2]
}
