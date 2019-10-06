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
	elementNumberingSelector.options = []string{"ADINA", "NASTRAN"}
	elementNumberingSelector.title = "Elegir tipo de numeración"
	elementNumberingSelector.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{2, 3, 0})
	poller.selector = &elementNumberingSelector
	elementNumberingSelector.Init()
	elementNumberingSelector.Render()
	poller.askedToPoll <- true
	var elementNumbering int
	select {
	case elementNumbering = <-elementNumberingSelector.selection:

	}
	elementNumbering++
	// MESH COLLECTOR MENU
	meshcol, err := readMeshCollectors(fileDir)
	if err != nil {
		return
	}
	collectorSelector := NewSelector()
	collectorSelector.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{2, 3, 0})
	collectorSelector.options = meshcol.KeySlice()
	collectorSelector.title = "Procesamiento de colector"
	poller.selector = &collectorSelector
	collectorSelector.Init()
	collectorSelector.Render()
	//poller.askedToPoll<-true
	var collectorSelection int
	select {
	case collectorSelection = <-collectorSelector.selection:
	}

	err = writeMeshCollector(fileDir, collectorSelector.options[collectorSelection], elementNumbering)
	if err != nil {
		panic("Could not write mesh collector")
	}

}

func dbslp() { // Debug sleep
	time.Sleep(time.Millisecond * 1000)
}
