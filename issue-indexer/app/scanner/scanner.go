package scanner

import (
	"issue-indexer/pckg/runtimeinfo"
	"time"
)

type signal int

const (
	run  signal = 1
	stop signal = 0
)

type DatabaseUpdates struct {
	ScanAfterSecond time.Duration
	signalsChannel  chan signal
	actionChannel chan signal
}

func NewDatabaseUpdates(scanAfterSecond time.Duration) *DatabaseUpdates {
	scanner := &DatabaseUpdates{
		ScanAfterSecond: scanAfterSecond,
		signalsChannel:  make(chan signal),
		actionChannel:  make(chan signal),
	}
	go scanner.scanSignals()
	return scanner
}

func (scanner *DatabaseUpdates) do() {
	for {
		select {
		case _ = <- scanner.actionChannel:
			now := time.Now()
			runtimeinfo.LogInfo("SCANNER IS STOPPED! [", now, "]")
			break
		case now := <-time.Tick(scanner.ScanAfterSecond):
			runtimeinfo.LogInfo("SCANNER NOW! [", now, "]")
			break
		}
	}
}

func (scanner *DatabaseUpdates) scanSignals() {
	for signal := range scanner.signalsChannel {
		switch signal {
		case stop:
			scanner.actionChannel <- stop
		case run:
			go scanner.do()
		}
	}
}

func (scanner *DatabaseUpdates) Stop() {
	go func() {
		scanner.signalsChannel <- stop
	}()
}

func (scanner *DatabaseUpdates) Run() {
	go func() {
		scanner.signalsChannel <- run
	}()
}
