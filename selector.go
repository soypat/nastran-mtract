package main

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"time"
)

var colorWheel = []ui.Color{ui.ColorGreen, ui.ColorBlue, ui.ColorCyan, ui.ColorYellow, ui.ColorRed, ui.ColorMagenta}
var selectedTheme = 0

type selector struct {
	options       []string
	title         string
	color         ui.Color
	border        bool
	selectedColor ui.Color
	*fitting
	associatedList *widgets.List
	// asociado a una accion/seleccion:
	selection chan int
	action    string
}

func NewSelector() selector {
	return selector{border: true, color: ui.ColorYellow, selectedColor: ui.ColorClear, selection: make(chan int)}
}
func (seltr *selector) Init() {
	seltr.associatedList = widgets.NewList() // Crea la lista asociada nueva
	seltr.associateList()
}
func (seltr *selector) associateList() {
	seltr.associatedList.Rows = seltr.options
	seltr.associatedList.TextStyle = ui.NewStyle(seltr.color)
	seltr.associatedList.SelectedRowStyle = ui.NewStyle(seltr.selectedColor)
	seltr.associatedList.SetRect(seltr.fitting.getRect())
	seltr.associatedList.Border = seltr.border
	seltr.associatedList.Title = seltr.title
	//seltr.associatedList.SelectedRow = seltr.selection // DO NOT ADD THIS! RISK OF TOTAL UTTER BREAKAGE
}

func (seltr *selector) Render() {
	seltr.associateList()
	ui.Render(seltr.associatedList)
}

func (seltr *selector) renderAction(ID *string) {
	if *ID == "" {
		return
	}
	switch *ID {
	case "<C-w>", "<C-W>":
		selectedTheme++
		if selectedTheme > len(colorWheel)-1 {
			selectedTheme = 0
		}
	case "<Up>", "j":
		seltr.associatedList.ScrollUp()
	case "<Down>", "k":
		seltr.associatedList.ScrollDown()
	case "<Enter>":
		seltr.selection <- seltr.associatedList.SelectedRow
		seltr.action = *ID
	case "<End>":
		seltr.associatedList.ScrollBottom()
	case "<Home>":
		seltr.associatedList.ScrollTop()
	default:
		seltr.action = *ID
		time.Sleep(time.Millisecond * 15)
	}
	seltr.Render()
}

func (seltr *selector) GetDims() (X int, Y int) {
	x1, y1, x2, y2 := seltr.getRect()
	if x1 > x2 {
		X = x1 - x2
	} else {
		X = x2 - x1
	}
	if y1 > y2 {
		Y = y1 - y2
	} else {
		Y = y2 - y1
	}
	return X, Y
}
