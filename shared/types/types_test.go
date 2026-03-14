package types

import (
	"testing"
	"time"
)

func TestIsValidGUID_Valid(t *testing.T) {
	if !IsValidGUID("550e8400-e29b-41d4-a716-446655440000") {
		t.Error("should be valid")
	}
}

func TestIsValidGUID_Invalid(t *testing.T) {
	if IsValidGUID("") {
		t.Error("empty should be invalid")
	}
	if IsValidGUID("invalid") {
		t.Error("invalid should be invalid")
	}
}

func TestNewGUID(t *testing.T) {
	g1 := NewGUID()
	g2 := NewGUID()
	if g1 == "" || g1 == g2 {
		t.Error("GUIDs invalid")
	}
	if !IsValidGUID(g1) {
		t.Error("not valid GUID")
	}
}

func TestRemoveHyphens(t *testing.T) {
	if got := RemoveHyphens("550e8400-e29b-41d4-a716-446655440000"); got != "550e8400e29b41d4a716446655440000" {
		t.Error("remove")
	}
}

func TestInsertHyphens(t *testing.T) {
	if got := InsertHyphens("550e8400e29b41d4a716446655440000"); got != "550e8400-e29b-41d4-a716-446655440000" {
		t.Error("insert")
	}
}

func TestTicksToDuration(t *testing.T) {
	if got := TicksToDuration(10000000); got.Seconds() != 1 {
		t.Error("1 sec ticks")
	}
}

func TestDurationToTicks(t *testing.T) {
	dur := time.Second
	if got := DurationToTicks(dur); got != 10000000 {
		t.Error("sec to ticks")
	}
}
