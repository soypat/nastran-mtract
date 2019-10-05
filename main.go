package main

import (
	"fmt"
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
	poller := NewPoller()
	selly := NewSelector()
	rows := []string{"Sup", "nuck", "kuts"}
	selly.options = rows
	selly.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{2, 3, 0})
	poller.selector = &selly
	selly.Init()
	poller.InitPoll()
	defer close(poller.askedToPoll)
	go poller.Poll2(poller.askedToPoll)
	poller.askedToPoll <- true
	var fileSelection int
	select {
	case fileSelection = <-selly.selection:
		panic(fmt.Sprintf("you selected file %i", fileSelection))
	}
}

func dbslp() { // Debug sleep
	time.Sleep(time.Millisecond * 1000)
}
