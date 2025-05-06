package process

import (
	"TelecomTask/internal/config"
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

type Event struct {
	Time         string
	EventID      int
	CompetitorID int
	ExtraParams  []string
}

type Competitor struct {
	ID              int
	Registered      bool
	StartTime       time.Time
	ActualStart     time.Time
	LapTimes        []time.Duration
	PenaltyTimes    []time.Duration
	Hits            map[int][]int
	Shots           map[int]int
	Status          string
	CurrentLap      int
	PenaltyLaps     int
	FiringRange     int
	LastPenaltyTime time.Time
	LastLapTime     time.Time
}

type LapDetail struct {
	Time  string
	Speed float64
}

type Report struct {
	CompetitorID int
	TotalTime    string
	LapDetails   []LapDetail
	PenaltyTime  string
	PenaltySpeed float64
	HitsShots    string
}

// parseEvent parses events from file into Event struct
func parseEvent(line string) (Event, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return Event{}, fmt.Errorf("invalid event format: %s", line)
	}
	var eventID int
	var competitorID int
	if _, err := fmt.Sscanf(parts[1], "%d", &eventID); err != nil {
		return Event{}, fmt.Errorf("invalid event id: %s", parts[1])
	}
	if _, err := fmt.Sscanf(parts[2], "%d", &competitorID); err != nil {
		return Event{}, fmt.Errorf("invalid competitor id: %s", parts[2])
	}

	timeStr := strings.Trim(parts[0], "[]")
	extraParams := parts[3:]
	return Event{
		Time:         timeStr,
		EventID:      eventID,
		CompetitorID: competitorID,
		ExtraParams:  extraParams,
	}, nil
}

// LoadEvents loads events from file and converts them into Event slice
func LoadEvents(filename string) ([]Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("LoadEvents: error opening the file: %w", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Printf("LoadEvents: %s\n", err.Error())
		}
	}(file)

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		event, err := parseEvent(line)
		if err != nil {
			return nil, fmt.Errorf("LoadEvents: error parsing the file: %v", err)
		}
		events = append(events, event)
	}
	return events, nil
}

// LogEvent logs input event with message
func LogEvent(event Event, message string) {
	log.Printf("[%s] %s\n", event.Time, message)
}

// formatDuration formats input time duration into correct format
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	millis := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}

// GenerateReport generates report by map of competitors
func GenerateReport(competitors map[int]*Competitor, config *config.Config) []Report {
	var reports []Report
	for _, comp := range competitors {
		totalTime := time.Duration(0)
		for _, lt := range comp.LapTimes {
			totalTime += lt
		}
		for _, pt := range comp.PenaltyTimes {
			totalTime += pt
		}

		var lapDetails []LapDetail
		for _, lt := range comp.LapTimes {
			speed := 0.0
			if lt.Seconds() > 0 {
				speed = float64(config.LapLen) / lt.Seconds()
			}
			lapDetails = append(lapDetails, LapDetail{
				Time:  formatDuration(lt),
				Speed: speed,
			})
		}
		for len(lapDetails) < config.Laps {
			lapDetails = append(lapDetails, LapDetail{Time: "", Speed: 0.0})
		}

		penaltyTime := time.Duration(0)
		for _, pt := range comp.PenaltyTimes {
			penaltyTime += pt
		}
		penaltySpeed := 0.0
		if penaltyTime > 0 {
			penaltySpeed = float64(config.PenaltyLen*len(comp.PenaltyTimes)) / penaltyTime.Seconds()
		}

		totalHits, totalShots := 0, 0
		for _, hits := range comp.Hits {
			totalHits += len(hits)
		}
		for _, shots := range comp.Shots {
			totalShots += shots
		}

		report := Report{
			CompetitorID: comp.ID,
			TotalTime:    comp.Status,
			LapDetails:   lapDetails,
			PenaltyTime:  formatDuration(penaltyTime),
			PenaltySpeed: penaltySpeed,
			HitsShots:    fmt.Sprintf("%d/%d", totalHits, totalShots),
		}
		if comp.Status == "Finished" {
			report.TotalTime = formatDuration(totalTime)
		}
		reports = append(reports, report)
	}

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].TotalTime == "NotStarted" || reports[i].TotalTime == "NotFinished" {
			return false
		}
		if reports[j].TotalTime == "NotStarted" || reports[j].TotalTime == "NotFinished" {
			return true
		}
		ti, _ := time.ParseDuration(strings.Replace(reports[i].TotalTime, ":", "h", 1) + "m" + "s")
		tj, _ := time.ParseDuration(strings.Replace(reports[j].TotalTime, ":", "h", 1) + "m" + "s")
		return ti < tj
	})

	return reports
}

// Events generate map of competitors and slice of outgoing events
func Events(config *config.Config, events []Event) (map[int]*Competitor, []Event) {
	competitors := make(map[int]*Competitor)
	var outgoingEvents []Event

	for _, event := range events {
		comp, exists := competitors[event.CompetitorID]
		if !exists {
			comp = &Competitor{
				ID:          event.CompetitorID,
				Hits:        make(map[int][]int),
				Shots:       make(map[int]int),
				Status:      "NotStarted",
				CurrentLap:  -1,
				LastLapTime: time.Time{},
			}
			competitors[event.CompetitorID] = comp
		}

		eventTime, err := time.Parse("15:04:05.000", event.Time)
		if err != nil {
			log.Printf("Events: error in extraParams string format: %v", err)
			continue
		}

		switch event.EventID {
		case 1:
			comp.Registered = true
			comp.Status = "Registered"
			LogEvent(event, "The competitor registered")

		case 2:
			startTime, err := time.Parse("15:04:05.000", event.ExtraParams[0])
			if err != nil {
				log.Printf("Events: error in extraParams string format: %v", err)
				continue
			}
			comp.StartTime = startTime
			LogEvent(event, fmt.Sprintf("The start time was set by a draw to %s", event.ExtraParams[0]))

		case 3:
			LogEvent(event, "The competitor is on the start line")

		case 4:
			comp.ActualStart = eventTime
			comp.Status = "Started"
			comp.LastLapTime = eventTime
			startDelta, _ := time.ParseDuration(strings.Replace(config.StartDelta, ":", "h", 1) + "m" + "s")
			if eventTime.Sub(comp.StartTime) > startDelta {
				comp.Status = "NotStarted"
				outgoingEvents = append(outgoingEvents, Event{
					Time:         event.Time,
					EventID:      32,
					CompetitorID: comp.ID,
				})
				LogEvent(Event{Time: event.Time, EventID: 32, CompetitorID: comp.ID}, "The competitor is disqualified")
			} else {
				LogEvent(event, "The competitor has started")
			}

		case 5:
			var rangeID int
			_, err := fmt.Sscanf(event.ExtraParams[0], "%d", &rangeID)
			if err != nil {
				log.Printf("Events: error in extraParams string format: %v", err)
				continue
			}
			comp.FiringRange = rangeID
			comp.Shots[comp.FiringRange] = 5
			LogEvent(event, fmt.Sprintf("The competitor is on the firing range(%d)", rangeID))

		case 6:
			var target int
			_, err := fmt.Sscanf(event.ExtraParams[0], "%d", &target)
			if err != nil {
				log.Printf("Events: error in extraParams string format: %v", err)
				continue
			}
			comp.Hits[comp.FiringRange] = append(comp.Hits[comp.FiringRange], target)
			LogEvent(event, fmt.Sprintf("The target(%d) has been hit by competitor", target))

		case 7:
			misses := comp.Shots[comp.FiringRange] - len(comp.Hits[comp.FiringRange])
			comp.PenaltyLaps += misses
			LogEvent(event, "The competitor left the firing range")

		case 8:
			comp.LastPenaltyTime = eventTime
			LogEvent(event, "The competitor entered the penalty laps")

		case 9:
			penaltyTime := eventTime.Sub(comp.LastPenaltyTime)
			comp.PenaltyTimes = append(comp.PenaltyTimes, penaltyTime)
			comp.PenaltyLaps--
			LogEvent(event, "The competitor left the penalty laps")

		case 10:
			comp.CurrentLap++
			var lapTime time.Duration
			if comp.CurrentLap == 0 {
				lapTime = eventTime.Sub(comp.ActualStart)
			} else {
				lapTime = eventTime.Sub(comp.LastLapTime)
			}
			comp.LapTimes = append(comp.LapTimes, lapTime)
			comp.LastLapTime = eventTime
			LogEvent(event, fmt.Sprintf("The competitor ended the main lap %d", comp.CurrentLap+1))
			if comp.CurrentLap+1 == config.Laps && comp.PenaltyLaps == 0 {
				comp.Status = "Finished"
				outgoingEvents = append(outgoingEvents, Event{
					Time:         event.Time,
					EventID:      33,
					CompetitorID: comp.ID,
				})
				LogEvent(Event{Time: event.Time, EventID: 33, CompetitorID: comp.ID}, "The competitor has finished")
			}

		case 11:
			comp.Status = "NotFinished"
			event.Time = eventTime.Format("15:04:05.000")
			LogEvent(event, fmt.Sprintf("The competitor can't continue: %s", strings.Join(event.ExtraParams, " ")))
		}
	}
	return competitors, outgoingEvents
}
