package main

import (
	"TelecomTask/internal/config"
	"TelecomTask/internal/process"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	cfg, err := config.New("./config/config.json")
	if err != nil {
		log.Fatal("Error loading config: ", err)
		return
	}
	events, err := process.LoadEvents("events")
	if err != nil {
		log.Fatal("Error loading events: ", err)
		return
	}
	competitors, outgoingEvents := process.Events(cfg, events)
	reports := process.GenerateReport(competitors, cfg)

	logFile, err := os.Create("output.log")
	if err != nil {
		log.Fatal("error creating log file: ", err)
		return
	}
	defer func(logFile *os.File) {
		err = logFile.Close()
		if err != nil {
			log.Fatal("error closing log file", err)
		}
	}(logFile)
	log.SetOutput(logFile)
	for _, event := range events {
		switch event.EventID {
		case 1:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) registered", event.CompetitorID))
		case 2:
			process.LogEvent(event, fmt.Sprintf("The start time for competitor(%d) was set by a draw to %s", event.CompetitorID, event.ExtraParams[0]))
		case 3:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) is on the start line", event.CompetitorID))
		case 4:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) has started", event.CompetitorID))
		case 5:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) is on the firing range(%s)", event.CompetitorID, event.ExtraParams[0]))
		case 6:
			process.LogEvent(event, fmt.Sprintf("The target(%s) has been hit by competitor(%d)", event.ExtraParams[0], event.CompetitorID))
		case 7:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) left the firing range", event.CompetitorID))
		case 8:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) entered the penalty laps", event.CompetitorID))
		case 9:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) left the penalty laps", event.CompetitorID))
		case 10:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) ended the main lap", event.CompetitorID))
		case 11:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) can't continue: %s", event.CompetitorID, strings.Join(event.ExtraParams, " ")))
		}
	}
	for _, event := range outgoingEvents {
		switch event.EventID {
		case 32:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) is disqualified", event.CompetitorID))
		case 33:
			process.LogEvent(event, fmt.Sprintf("The competitor(%d) has finished", event.CompetitorID))
		}
	}

	// Вывод итогового отчета
	fmt.Printf("\n")
	for _, r := range reports {
		fmt.Printf("[%s] %d %v %s %.3f %s\n",
			r.TotalTime, r.CompetitorID, r.LapDetails, r.PenaltyTime, r.PenaltySpeed, r.HitsShots)
	}
}
