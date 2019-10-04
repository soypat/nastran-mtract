package nastran_mtract

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"time"
)

var colorWheel = []ui.Color{ui.ColorGreen, ui.ColorBlue, ui.ColorCyan, ui.ColorYellow, ui.ColorRed, ui.ColorMagenta}
var selectedTheme = 0

func NewMenu() menu {
	return menu{border: true, color: ui.ColorYellow, selectedColor: ui.ColorClear, selection: -1}
}

type menu struct {
	options       []string
	title         string
	color         ui.Color
	border        bool
	selectedColor ui.Color
	*fitting
	associatedList *widgets.List
	// asociado a una accion/seleccion:
	selection int
	action    string
}

type thePoller struct {
	askedToPoll chan bool
	isPolling   bool
	*menu
	keyPress chan string
}

func NewPoller() thePoller {
	return thePoller{isPolling: true}
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

type fitting struct {
	// This stuff aint as stuffy as it looks
	widthStart  [3]int // 3 unit vector used to position each of the 4 window borders
	heightStart [3]int // first two units are used to indicate fraction of window. Last one is a constant character postion. basically y=m*x+b
	widthEnd    [3]int // [1, 2, 3] means the border is positioned 3 character lengths after halfway mark of terminal
	heightEnd   [3]int // [0, 5, 6] means the border is positioned 6 characters where leftwise terminal border begins. 5 is meaningless
}

func InitMenu(theMenu *menu) {
	menu := widgets.NewList()
	menu.Rows = theMenu.options
	//menu.TextStyle = ui.NewStyle(theMenu.color)
	//menu.SelectedRowStyle = ui.NewStyle(theMenu.selectedColor)
	menu.SetRect(theMenu.fitting.getRect())
	menu.Border = theMenu.border

	theMenu.associatedList = menu // LAST LINE
}

//func (theMenu *menu) Poller(askedToPoll <-chan bool) {
//	polling := false
//
//	for {
//		polling = getRequest(askedToPoll, polling)
//
//		if polling {
//			keyIdentifier := AskForKeyPress()
//			switch keyIdentifier {
//			case "":
//				time.Sleep(time.Millisecond * 30)
//			case "<Up>", "j":
//				theMenu.associatedList.ScrollUp()
//				//theMenu.associatedList.TextStyle = ui.NewStyle()
//			case "<Down>", "k":
//				theMenu.associatedList.ScrollDown()
//			case "<Enter>":
//				theMenu.selection = theMenu.associatedList.SelectedRow
//				theMenu.action = keyIdentifier
//			case "<End>":
//				theMenu.associatedList.ScrollBottom()
//			case "<Home>":
//				theMenu.associatedList.ScrollTop()
//			default:
//				theMenu.selection = theMenu.associatedList.SelectedRow
//				theMenu.action = keyIdentifier
//				continue
//			}
//			theMenu.associatedList.SelectedRowStyle = ui.NewStyle(theMenu.selectedColor)
//			ui.Render(theMenu.associatedList)
//		} else if askedToPoll == nil {
//			//time.Sleep(1000*time.Millisecond) // TODO make this even cleaner. a return would be fantastic
//			return
//		} else {
//			_, ok := <-askedToPoll // CHECK IF CHANNEL CLOSED
//			if !ok {
//				time.Sleep(20 * time.Millisecond)
//				return
//			}
//			time.Sleep(1000 * time.Millisecond)
//			theMenu.associatedList.SelectedRowStyle = ui.NewStyle(theMenu.color)
//		}
//	}
//}

func getRequest(response <-chan bool, currentUnderstanding bool) bool { // Basic concurrency function.
	select {
	case whatIHeard := <-response:
		if whatIHeard == true {
			return true
		} else {
			return false
		}
	default:
		return currentUnderstanding // If we do not recieve answer, continue doing what you were doing
	}
}

func (poller *thePoller) startPolling() { // Writes pressed key to channel
	uiEvents := ui.PollEvents()
	poller.isPolling = true
	for {
		keepGoing := <-poller.askedToPoll
		switch keepGoing {
		case false:
			poller.stopPolling()
		}
		e := <-uiEvents
		switch e.ID {
		case "q", "Q": // funciones intrinsicas, siempre disponibles
			ui.Clear()
			ui.Close()
			//close(askToRenderMain)
			panic("User Exited. [q] Press.")
		case "<C-w>", "<C-W>":
			selectedTheme++
			if selectedTheme > len(colorWheel)-1 {
				selectedTheme = 0
			}
		// Menu functions
		case "":
			time.Sleep(time.Millisecond * 20)
		case "<Up>", "j":
			poller.associatedList.ScrollUp()
			//poller.associatedList.TextStyle = ui.NewStyle()
		case "<Down>", "k":
			poller.associatedList.ScrollDown()
		case "<Enter>":
			if poller.isPolling {
				poller.selection = poller.associatedList.SelectedRow
				poller.action = e.ID
			}
		case "<End>":
			poller.associatedList.ScrollBottom()
		case "<Home>":
			poller.associatedList.ScrollTop()
		default:
			poller.selection = poller.associatedList.SelectedRow
			poller.action = e.ID
			poller.keyPress <- e.ID
			time.Sleep(time.Millisecond * 15)
		}
	}
}

func (poller *thePoller) stopPolling() {
	poller.isPolling = false
	close(poller.keyPress)
	close(poller.askedToPoll)
}

func (theMenu menu) GetDims() (X int, Y int) {
	x1, y1, x2, y2 := theMenu.getRect()
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

func (P fitting) getRect() (int, int, int, int) {
	width, height := ui.TerminalDimensions()
	return width*P.widthStart[0]/P.widthStart[1] + P.widthStart[2], height*P.heightStart[0]/P.heightStart[1] + P.heightStart[2], width*P.widthEnd[0]/P.widthEnd[1] + P.widthEnd[2], height*P.heightEnd[0]/P.heightEnd[1] + P.heightEnd[2]
}
