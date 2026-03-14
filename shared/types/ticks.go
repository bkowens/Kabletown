package types

import (
	"fmt"
	"time"
)

// Ticks conversion constants
// Ticks in .NET/Jellyfin are 100-nanosecond intervals
const (
	// TicksPerMillisecond is 10,000 100-ns intervals per ms
	TicksPerMillisecond = 10000

	// TicksPerSecond is 10,000,000 100-ns intervals per second
	TicksPerSecond = 10000000

	// TicksPerMinute is 60 * 10,000,000
	TicksPerMinute = TicksPerSecond * 60

	// TicksPerHour is 60 * 60 * 10,000,000
	TicksPerHour = TicksPerSecond * 3600

	// TicksPerDay is 24 * 60 * 60 * 10,000,000
	TicksPerDay = TicksPerSecond * 86400

	// TicksEpochStart is the number of ticks from DateTime.MinValue (year 0001, month 1, day 1)
	// to Unix epoch (1970-01-01 00:00:00 UTC)
	// .NET DateTime ticks start from 0001-01-01, Unix epoch is 1970-01-01
	// That's 62,135,596,800,000,000 ticks
	EpochTicks = 62135596800000000
)

// TicksToDuration converts Jellyfin ticks to time.Duration
// Ticks are 100-ns intervals
func TicksToDuration(ticks int64) time.Duration {
	return time.Duration(ticks * 100) // Convert 100-ns to nanoseconds
}

// DurationToTicks converts time.Duration to Jellyfin ticks
func DurationToTicks(duration time.Duration) int64 {
	return duration.Nanoseconds() / 100 // Convert nanoseconds to 100-ns ticks
}

// DurationToTicksFloat converts time.Duration to Jellyfin ticks as float64
func DurationToTicksFloat(duration time.Duration) float64 {
	return float64(duration.Nanoseconds()) / 100.0
}

// TicksToDurationFloat converts Jellyfin ticks to time.Duration as float64
func TicksToDurationFloat(ticks float64) time.Duration {
	return time.Duration(ticks * 100) // Convert 100-ns to nanoseconds
}

// MillisecondsToTicks converts milliseconds to Jellyfin ticks
func MillisecondsToTicks(ms int64) int64 {
	return ms * TicksPerMillisecond
}

// TicksToMilliseconds converts Jellyfin ticks to milliseconds
func TicksToMilliseconds(ticks int64) int64 {
	return ticks / TicksPerMillisecond
}

// SecondsToTicks converts seconds to Jellyfin ticks
func SecondsToTicks(seconds int64) int64 {
	return seconds * TicksPerSecond
}

// TicksToSeconds converts Jellyfin ticks to seconds
func TicksToSeconds(ticks int64) int64 {
	return ticks / TicksPerSecond
}

// MinutesToTicks converts minutes to Jellyfin ticks
func MinutesToTicks(minutes int64) int64 {
	return minutes * TicksPerMinute
}

// TicksToMinutes converts Jellyfin ticks to minutes
func TicksToMinutes(ticks int64) int64 {
	return ticks / TicksPerMinute
}

// HoursToTicks converts hours to Jellyfin ticks
func HoursToTicks(hours int64) int64 {
	return hours * TicksPerHour
}

// TicksToHours converts Jellyfin ticks to hours
func TicksToHours(ticks int64) int64 {
	return ticks / TicksPerHour
}

// DaysToTicks converts days to Jellyfin ticks
func DaysToTicks(days int64) int64 {
	return days * TicksPerDay
}

// TicksToDays converts Jellyfin ticks to days
func TicksToDays(ticks int64) int64 {
	return ticks / TicksPerDay
}

// SecondsToMilliseconds converts seconds to milliseconds
func SecondsToMilliseconds(sec int64) int64 {
	return sec * 1000
}

// MillisecondsToSeconds converts milliseconds to seconds
func MillisecondsToSeconds(ms int64) int64 {
	return ms / 1000
}

// TimeToTicks converts a time.Time to Jellyfin ticks
// Jellyfin ticks are 100-ns intervals since year 0001-01-01
func TimeToTicks(t time.Time) int64 {
	// Get Unix nanoseconds and convert to ticks
	unixNanos := t.UnixNano()
	// Add epoch offset and convert to ticks (100ns per tick)
	return (unixNanos/100) + EpochTicks
}

// TicksToTime converts Jellyfin ticks to time.Time
func TicksToTime(ticks int64) time.Time {
	// Remove epoch offset and convert ticks to nanoseconds
	unixNanos := (ticks - EpochTicks) * 100
	return time.Unix(0, unixNanos).UTC()
}

// TimeToDuration calculates the duration between two times in ticks
func TimeToDuration(start, end time.Time) int64 {
	return DurationToTicks(end.Sub(start))
}

// DurationFromTicks creates a time.Duration from Jellyfin ticks
func DurationFromTicks(ticks int64) time.Duration {
	return TicksToDuration(ticks)
}

// FormatDuration formats a duration as a human-readable string
func FormatDuration(duration time.Duration) string {
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	switch {
	case hours > 0:
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// FormatDurationTicks formats ticks as milliseconds
func FormatDurationTicks(ticks int64) string {
	ms := TicksToMilliseconds(ticks)
	return fmt.Sprintf("%d ms", ms)
}

// DurationFormat formats a duration for display
func DurationFormat(d time.Duration) string {
	h := d / time.Hour
	d = d % time.Hour
	m := d / time.Minute
	d = d % time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// DurationFormatDetailed formats a duration more descriptively
func DurationFormatDetailed(d time.Duration) string {
	h := d / time.Hour
	d = d % time.Hour
	m := d / time.Minute
	d = d % time.Minute
	s := d / time.Second

	switch {
	case h > 0:
		return fmt.Sprintf("%d hour(s), %d minute(s), %d second(s)", h, m, s)
	case m > 0:
		return fmt.Sprintf("%d minute(s), %d second(s)", m, s)
	case s > 0:
		return fmt.Sprintf("%d second(s)", s)
	default:
		return "0 second(s)"
	}
}
