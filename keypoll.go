package main

import (
	ui "github.com/gizak/termui/v3"
	"log"
	"time"
)

type thePoller struct {
	askedToPoll chan bool
	isPolling   bool
	event       <-chan ui.Event
	*selector
	keyPress chan string
}

func NewPoller() thePoller {
	askToPoll := make(chan bool)
	return thePoller{isPolling: true, askedToPoll: askToPoll}
}

func (poller *thePoller) InitPoll() {
	poller.event = ui.PollEvents()
	poller.isPolling = true
}

func (poller *thePoller) resumePoll() {
	poller.askedToPoll <- true
}
func (poller *thePoller) pausePoll() {
	poller.askedToPoll <- false
}

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

func (poller *thePoller) Poll2(askToPoll <-chan bool) {
	var c ui.Event
	var ok = true
	for {
		//poller.isPolling = getRequest(poller.askedToPoll, poller.isPolling) // getRequest is a basic concurrency function!
		poller.isPolling = getRequest(askToPoll, poller.isPolling) // getRequest is a basic concurrency function!
		if poller.isPolling {
			select {
			case c, ok = <-poller.event:
			}
		}

		if !ok {
			log.Fatal("Unexpected key polling channel close.")
			return // Return on unexpected key polling close if fatal log doesent do job
		}
		if poller.isPolling {
			switch c.ID {
			case "q", "Q":
				ui.Clear()
				ui.Close()
				panic("Exiting upon user request. [q] press.")
			default:
				poller.selector.renderAction(&c.ID)
			}

		} else if poller.isPolling == false {
			request, ok := <-askToPoll
			if !ok {
				return // Return on channel closed
			}
			if request {
				poller.isPolling = true
			}
			time.Sleep(time.Millisecond * 30) // Si no estoy polleando espero un rato para relajar uso de procesador}
		} else {
			log.Fatal("Close of polling channel on poller!")
			return // RETURN ON NIL VAL
		}
	}
}

//func (poller *thePoller) Poll(askToPoll <-chan bool) {
//	select {
//	case poller.isPolling = <-askToPoll:
//
//	case c := <-poller.event:
//		if poller.isPolling {
//			switch c.ID {
//			case "q", "Q":
//				ui.Clear()
//				ui.Close()
//				panic("Exiting upon user request. [q] press.")
//			}
//		} else {
//			time.Sleep(time.Millisecond * 100) // Si no estoy polleando espero un rato para relajar uso de procesador}
//		}
//	}
//}
