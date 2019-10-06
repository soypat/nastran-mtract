package main

import (
	ui "github.com/gizak/termui/v3"
	_ "github.com/gizak/termui/v3/widgets"
	"log"
	"time"
)

func main() {
	if err := ui.Init(); err != nil { // Inicializo el command line UI
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	// Creo primer menu para otorgarle espacio
	selly := NewSelector()
	selly.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{2, 3, 0})
	fileListWidth, _ := selly.GetDims()

	displayFileNames, fileNames, err := fileListCurrentDirectory(fileListWidth)
	if err != nil {
		log.Fatalf("Failed file search: %v", err)
	} else if len(fileNames) == 0 {
		log.Fatalf("No se encontraron archivos compatibles")
	}
	selly.options = displayFileNames
	poller := NewPoller()
	poller.selector = &selly
	selly.title = "Seleccione su archivo NASTRAN"
	selly.Init()
	poller.InitPoll()
	defer close(poller.askedToPoll)
	go poller.Poll2(poller.askedToPoll)
	poller.askedToPoll <- true
	var fileSelection int

	select { // File selector!
	case fileSelection = <-selly.selection:
		poller.askedToPoll <- false
	}
	fileDir := fileNames[fileSelection]
	fileDir = fileDir[2:]
	err = writeNodosCSV("nodos.csv", fileDir)
	if err != nil {
		return
	}

}

func dbslp() { // Debug sleep
	time.Sleep(time.Millisecond * 1000)
}
