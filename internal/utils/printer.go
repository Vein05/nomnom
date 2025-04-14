package nomnom

import (
	"fmt"
	"sync"
)

// Printer struct to manage logging
type Printer struct {
	logCh chan string
	wg    sync.WaitGroup
}

func NewPrinter() *Printer {
	Printer := &Printer{
		logCh: make(chan string, 100), // Buffered channel to avoid blocking
	}
	// Start a goroutine to process logs
	Printer.wg.Add(1)
	go func() {
		defer Printer.wg.Done()
		for msg := range Printer.logCh {
			fmt.Println(msg)
		}
	}()
	return Printer
}

func (l *Printer) Log(msg string) {
	l.logCh <- msg
}

func (l *Printer) Close() {
	close(l.logCh)
	l.wg.Wait()
}
