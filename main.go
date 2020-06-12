package main

import (
	"bufio"
	ui "github.com/gizak/termui/v3"
	_ "github.com/gizak/termui/v3/widgets"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// First check file directory if files present
	_, fileNames, err := fileListCurrentDirectory(10)
	if err != nil {
		waitForUserInput("Could not read directory.\nPress [ENTER] to end program.")
		os.Exit(1)
	} else if len(fileNames) == 0 {
		waitForUserInput("No compatible .dat/.txt files found.\nPress [ENTER] to end program.")
		os.Exit(1)
	}


	if err := ui.Init(); err != nil { // Inicializo el command line UI
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Creo primer menu para otorgarle espacio
	selly := NewSelector()
	selly.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{2, 3, 0})
	fileListWidth, _ := selly.GetDims()
	displayFileNames, fileNames, err := fileListCurrentDirectory(fileListWidth)
	selly.options = displayFileNames
	poller := NewPoller()
	poller.selector = &selly
	selly.title = "Select NASTRAN file. [q] to exit."
	selly.Init()
	poller.InitPoll()
	defer close(poller.askedToPoll)
	go poller.Poll2(poller.askedToPoll)
	poller.askedToPoll <- true
	var fileSelection int

	select { // File selector!
	case fileSelection = <-selly.selection:
		//poller.askedToPoll <- false
	}
	poller.askedToPoll <- false
	fileDir := fileNames[fileSelection]
	fileDir = fileDir[2:]
	err = writeNodosCSV("nodos.csv", fileDir)
	if err != nil {
		return
	}
	// MENU PARA ELEGIR NUMERACION DE ELEMENTOS
	elementNumberingSelector := NewSelector()
	elementNumberingSelector.options = []string{"ADINA", "NASTRAN (not recommended)"}
	elementNumberingSelector.title = "Select element numbering. ADINA recommended for MATLAB use."
	elementNumberingSelector.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{2, 3, 0}, [3]int{2, 3, 0})
	poller.selector = &elementNumberingSelector
	elementNumberingSelector.Init()
	elementNumberingSelector.Render()
	poller.askedToPoll <- true
	var elementNumbering int
	select {
	case elementNumbering = <-elementNumberingSelector.selection:

	}
	// MESH COLLECTOR MENU
	meshcol, err := readMeshCollectors(fileDir)
	if err != nil {
		return
	}
	collectorSelector := NewSelector()
	collectorSelector.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{2, 3, 0}, [3]int{3, 3, 0})
	collectorSelector.options = meshcol.KeySlice()
	collectorSelector.title = "Colector processing"
	poller.selector = &collectorSelector
	collectorSelector.Init()
	collectorSelector.Render()
	//poller.askedToPoll<-true
	var collectorSelection int
	var selectedEntity entity
	for {
		select {
		case collectorSelection = <-collectorSelector.selection:
			selectedEntity = meshcol[collectorSelector.options[collectorSelection]]
			collectorName := collectorSelector.options[collectorSelection]
			if selectedEntity.getNumber() == 0 { // Constraint Selection
				err = writeCollector(fileDir, collectorName, elementNumbering)
				if err != nil {
					collectorSelector.title = "Error reading constraint."
					collectorSelector.Render()
				} else {
					collectorSelector.title = "Completed constraint: " + collectorName + ". Press [q] to exit.q"
					collectorSelector.Render()
				}
			} else {
				err = writeCollector(fileDir, collectorName, elementNumbering)
				collectorSelector.title = "Completed " + strconv.Itoa(selectedEntity.getNumber()) + ". Press [q] to exit. Patricio Whittingslow 2020. Github: soypat"
				collectorSelector.Render()
			}
		}
		time.Sleep(time.Millisecond * 14)
	}
	// end of program.
}
func waitForUserInput(message string) string {
	if message == "" {
		message = "\nPress a [Enter] to continue...\n"
	}
	input := bufio.NewScanner(os.Stdin)
	_,_ = os.Stdout.WriteString(message)
	input.Scan()
	return input.Text()
}
