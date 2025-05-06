package process

import (
	"TelecomTask/internal/config"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

// Helper function to compare slices of strings
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to compare maps of int to int
func mapsEqualInt(a, b map[int]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || va != vb {
			return false
		}
	}
	return true
}

// Helper function to compare maps of int to []int
func mapsEqualIntSlice(a, b map[int][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || !reflect.DeepEqual(va, vb) {
			return false
		}
	}
	return true
}

// TestParseEvent tests the parseEvent function with various inputs
func TestParseEvent(t *testing.T) {
	tests := []struct {
		input    string
		expected Event
		err      bool
	}{
		{"[09:05:59.867] 1 1", Event{"09:05:59.867", 1, 1, []string{}}, false},
		{"[09:15:00.841] 2 1 09:30:00.000", Event{"09:15:00.841", 2, 1, []string{"09:30:00.000"}}, false},
		{"invalid line", Event{}, true},
		{"[12:34:56.789] 3 2 param1 param2", Event{"12:34:56.789", 3, 2, []string{"param1", "param2"}}, false},
	}
	for _, test := range tests {
		event, err := parseEvent(test.input)
		if test.err && err == nil {
			t.Errorf("Expected error for input %s", test.input)
		} else if !test.err && err != nil {
			t.Errorf("Unexpected error for input %s: %v", test.input, err)
		} else if !test.err && (event.Time != test.expected.Time || event.EventID != test.expected.EventID || event.CompetitorID != test.expected.CompetitorID || !stringSlicesEqual(event.ExtraParams, test.expected.ExtraParams)) {
			t.Errorf("For input %s, expected %v, got %v", test.input, test.expected, event)
		}
	}
}

// TestLoadEvents tests loading events from a file
func TestLoadEvents(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err = os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpfile.Name())

	testEvents := "[09:05:59.867] 1 1\n[09:15:00.841] 2 1 09:30:00.000\n[12:34:56.789] 3 2 param1 param2"
	if _, err := tmpfile.Write([]byte(testEvents)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	events, err := LoadEvents(tmpfile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	if events[0].Time != "09:05:59.867" || events[0].EventID != 1 || events[0].CompetitorID != 1 {
		t.Errorf("First event mismatch: %v", events[0])
	}

	if events[1].Time != "09:15:00.841" || events[1].EventID != 2 || events[1].CompetitorID != 1 || len(events[1].ExtraParams) != 1 || events[1].ExtraParams[0] != "09:30:00.000" {
		t.Errorf("Second event mismatch: %v", events[1])
	}

	if events[2].Time != "12:34:56.789" || events[2].EventID != 3 || events[2].CompetitorID != 2 || len(events[2].ExtraParams) != 2 || events[2].ExtraParams[0] != "param1" || events[2].ExtraParams[1] != "param2" {
		t.Errorf("Third event mismatch: %v", events[2])
	}
}

// TestFormatDuration tests formatting of durations
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "00:00:00.000"},
		{1*time.Hour + 2*time.Minute + 3*time.Second + 4*time.Millisecond, "01:02:03.004"},
		{9*time.Minute + 48*time.Second, "00:09:48.000"},
		{1 * time.Minute, "00:01:00.000"},
	}
	for _, test := range tests {
		result := formatDuration(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %s, got %s", test.input, test.expected, result)
		}
	}
}

// TestEvents tests processing of events
func TestEvents(t *testing.T) {
	cfg := &config.Config{
		Laps:        2,
		LapLen:      1000,
		PenaltyLen:  100,
		FiringLines: 1,
		Start:       "10:00:00",
		StartDelta:  "00:00:10",
	}

	events := []Event{
		{"10:00:00.000", 1, 1, []string{}},
		{"10:00:05.000", 2, 1, []string{"10:00:10.000"}},
		{"10:00:08.000", 3, 1, []string{}},
		{"10:00:12.000", 4, 1, []string{}},
		{"10:05:00.000", 5, 1, []string{"1"}},
		{"10:05:10.000", 6, 1, []string{"1"}},
		{"10:05:20.000", 6, 1, []string{"2"}},
		{"10:05:30.000", 6, 1, []string{"3"}},
		{"10:05:40.000", 6, 1, []string{"4"}},
		{"10:05:50.000", 7, 1, []string{}},
		{"10:06:00.000", 8, 1, []string{}},
		{"10:07:00.000", 9, 1, []string{}},
		{"10:10:00.000", 10, 1, []string{}},
		{"10:30:05.000", 11, 1, []string{"Injury"}},
	}

	competitors, _ := Events(cfg, events)

	if len(competitors) != 1 {
		t.Errorf("Expected 1 competitor, got %d", len(competitors))
	}

	comp, ok := competitors[1]
	if !ok {
		t.Fatal("Competitor 1 not found")
	}

	if comp.Status != "NotFinished" {
		t.Errorf("Expected status NotFinished, got %s", comp.Status)
	}

	if len(comp.LapTimes) != 1 {
		t.Errorf("Expected 1 lap time, got %d", len(comp.LapTimes))
	}

	tStart, _ := time.Parse("15:04:05.000", "10:00:12.000")
	tEndLap1, _ := time.Parse("15:04:05.000", "10:10:00.000")
	expectedLapTime := tEndLap1.Sub(tStart)
	if comp.LapTimes[0] != expectedLapTime {
		t.Errorf("Expected lap time %v, got %v", expectedLapTime, comp.LapTimes[0])
	}

	tPenaltyStart, _ := time.Parse("15:04:05.000", "10:06:00.000")
	tPenaltyEnd, _ := time.Parse("15:04:05.000", "10:07:00.000")
	expectedPenaltyTime := tPenaltyEnd.Sub(tPenaltyStart)
	if len(comp.PenaltyTimes) != 1 || comp.PenaltyTimes[0] != expectedPenaltyTime {
		t.Errorf("Expected penalty time %v, got %v", expectedPenaltyTime, comp.PenaltyTimes)
	}

	expectedHits := map[int][]int{1: {1, 2, 3, 4}}
	if !mapsEqualIntSlice(comp.Hits, expectedHits) {
		t.Errorf("Expected hits %v, got %v", expectedHits, comp.Hits)
	}

	expectedShots := map[int]int{1: 5}
	if !mapsEqualInt(comp.Shots, expectedShots) {
		t.Errorf("Expected shots %v, got %v", expectedShots, comp.Shots)
	}

}

// TestGenerateReport tests generating the final report
func TestGenerateReport(t *testing.T) {
	cfg := &config.Config{
		Laps:        2,
		LapLen:      1000,
		PenaltyLen:  100,
		FiringLines: 1,
		Start:       "10:00:00",
		StartDelta:  "00:00:10",
	}

	competitors := map[int]*Competitor{
		1: {
			ID:           1,
			Status:       "NotFinished",
			LapTimes:     []time.Duration{9*time.Minute + 48*time.Second},
			PenaltyTimes: []time.Duration{1 * time.Minute},
			Hits:         map[int][]int{1: {1, 2, 3, 4}},
			Shots:        map[int]int{1: 5},
		},
	}

	reports := GenerateReport(competitors, cfg)

	if len(reports) != 1 {
		t.Errorf("Expected 1 report, got %d", len(reports))
	}

	report := reports[0]

	if report.CompetitorID != 1 {
		t.Errorf("Expected competitor ID 1, got %d", report.CompetitorID)
	}

	if report.TotalTime != "NotFinished" {
		t.Errorf("Expected TotalTime NotFinished, got %s", report.TotalTime)
	}

	expectedLapDetails := []LapDetail{
		{"00:09:48.000", float64(1000) / (9*60 + 48)},
		{"", 0.0},
	}

	if len(report.LapDetails) != 2 {
		t.Errorf("Expected 2 lap details, got %d", len(report.LapDetails))
	} else {
		for i := range expectedLapDetails {
			if report.LapDetails[i].Time != expectedLapDetails[i].Time {
				t.Errorf("Lap %d time mismatch: expected %s, got %s", i+1, expectedLapDetails[i].Time, report.LapDetails[i].Time)
			}
			if math.Abs(report.LapDetails[i].Speed-expectedLapDetails[i].Speed) > 0.001 {
				t.Errorf("Lap %d speed mismatch: expected %.3f, got %.3f", i+1, expectedLapDetails[i].Speed, report.LapDetails[i].Speed)
			}
		}
	}

	if report.PenaltyTime != "00:01:00.000" {
		t.Errorf("Expected PenaltyTime 00:01:00.000, got %s", report.PenaltyTime)
	}

	expectedPenaltySpeed := float64(100) / 60.0
	if math.Abs(report.PenaltySpeed-expectedPenaltySpeed) > 0.001 {
		t.Errorf("Expected PenaltySpeed %.3f, got %.3f", expectedPenaltySpeed, report.PenaltySpeed)
	}

	if report.HitsShots != "4/5" {
		t.Errorf("Expected HitsShots 4/5, got %s", report.HitsShots)
	}
}
