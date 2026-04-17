package jobs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ScheduleSpec calculates the next time a recurring job should run.
type ScheduleSpec interface {
	Next(after time.Time) time.Time
}

type intervalSchedule struct {
	every time.Duration
}

func (s intervalSchedule) Next(after time.Time) time.Time {
	if s.every <= 0 {
		return time.Time{}
	}
	return after.Add(s.every)
}

type cronSchedule struct {
	minute cronField
	hour   cronField
	dom    cronField
	month  cronField
	dow    cronField
}

type cronField struct {
	allowed map[int]struct{}
	any     bool
}

// ParseSchedule parses a five-field cron expression or aliases such as @every.
func ParseSchedule(spec string) (ScheduleSpec, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("%w: empty schedule", ErrInvalidSchedule)
	}

	if delay, err := time.ParseDuration(spec); err == nil && delay > 0 {
		return intervalSchedule{every: delay}, nil
	}

	switch {
	case strings.HasPrefix(spec, "@every "):
		delay, err := time.ParseDuration(strings.TrimSpace(strings.TrimPrefix(spec, "@every ")))
		if err != nil || delay <= 0 {
			return nil, fmt.Errorf("%w: @every requires a positive duration", ErrInvalidSchedule)
		}
		return intervalSchedule{every: delay}, nil
	case spec == "@hourly":
		spec = "0 * * * *"
	case spec == "@daily" || spec == "@midnight":
		spec = "0 0 * * *"
	case spec == "@weekly":
		spec = "0 0 * * 0"
	}

	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, fmt.Errorf("%w: expected 5 cron fields", ErrInvalidSchedule)
	}

	minute, err := parseCronField(fields[0], 0, 59, nil, 60)
	if err != nil {
		return nil, fmt.Errorf("%w: minute: %v", ErrInvalidSchedule, err)
	}
	hour, err := parseCronField(fields[1], 0, 23, nil, 24)
	if err != nil {
		return nil, fmt.Errorf("%w: hour: %v", ErrInvalidSchedule, err)
	}
	dom, err := parseCronField(fields[2], 1, 31, nil, 31)
	if err != nil {
		return nil, fmt.Errorf("%w: day-of-month: %v", ErrInvalidSchedule, err)
	}
	month, err := parseCronField(fields[3], 1, 12, monthAliases(), 12)
	if err != nil {
		return nil, fmt.Errorf("%w: month: %v", ErrInvalidSchedule, err)
	}
	dow, err := parseCronField(fields[4], 0, 7, weekdayAliases(), 7)
	if err != nil {
		return nil, fmt.Errorf("%w: day-of-week: %v", ErrInvalidSchedule, err)
	}

	return cronSchedule{
		minute: minute,
		hour:   hour,
		dom:    dom,
		month:  month,
		dow:    dow,
	}, nil
}

func (s cronSchedule) Next(after time.Time) time.Time {
	next := after.Truncate(time.Minute).Add(time.Minute)
	limit := next.AddDate(5, 0, 0)

	for !next.After(limit) {
		if s.matches(next) {
			return next
		}
		next = next.Add(time.Minute)
	}
	return time.Time{}
}

func (s cronSchedule) matches(t time.Time) bool {
	if !s.minute.allows(t.Minute()) || !s.hour.allows(t.Hour()) || !s.month.allows(int(t.Month())) {
		return false
	}

	domMatch := s.dom.allows(t.Day())
	dowMatch := s.dow.allows(int(t.Weekday()))
	if !s.dom.any && !s.dow.any {
		return domMatch || dowMatch
	}
	return domMatch && dowMatch
}

func (f cronField) allows(value int) bool {
	_, ok := f.allowed[value]
	return ok
}

func parseCronField(expr string, min, max int, aliases map[string]int, distinct int) (cronField, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return cronField{}, errors.New("empty field")
	}

	field := cronField{allowed: map[int]struct{}{}}
	for _, part := range strings.Split(expr, ",") {
		if err := fillCronPart(field.allowed, strings.TrimSpace(part), min, max, aliases); err != nil {
			return cronField{}, err
		}
	}

	field.any = len(field.allowed) == distinct
	return field, nil
}

func fillCronPart(allowed map[int]struct{}, part string, min, max int, aliases map[string]int) error {
	if part == "" {
		return errors.New("empty list item")
	}

	base := part
	step := 1
	if rawBase, rawStep, ok := strings.Cut(part, "/"); ok {
		base = rawBase
		value, err := strconv.Atoi(rawStep)
		if err != nil || value <= 0 {
			return fmt.Errorf("invalid step %q", rawStep)
		}
		step = value
	}

	start, end, err := cronRange(base, min, max, aliases)
	if err != nil {
		return err
	}
	if start > end {
		return fmt.Errorf("invalid range %q", base)
	}

	for value := start; value <= end; value += step {
		allowed[normalizeCronValue(value, max)] = struct{}{}
	}
	return nil
}

func cronRange(base string, min, max int, aliases map[string]int) (int, int, error) {
	base = strings.TrimSpace(base)
	if base == "*" {
		return min, max, nil
	}

	if start, end, ok := strings.Cut(base, "-"); ok {
		startValue, err := parseCronValue(start, min, max, aliases)
		if err != nil {
			return 0, 0, err
		}
		endValue, err := parseCronValue(end, min, max, aliases)
		if err != nil {
			return 0, 0, err
		}
		return startValue, endValue, nil
	}

	value, err := parseCronValue(base, min, max, aliases)
	if err != nil {
		return 0, 0, err
	}
	return value, value, nil
}

func parseCronValue(raw string, min, max int, aliases map[string]int) (int, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return 0, errors.New("empty value")
	}
	if aliases != nil {
		if value, ok := aliases[raw]; ok {
			return value, nil
		}
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q", raw)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("value %d outside %d-%d", value, min, max)
	}
	return value, nil
}

func normalizeCronValue(value, max int) int {
	if max == 7 {
		return normalizeWeekday(value)
	}
	return value
}

func normalizeWeekday(value int) int {
	if value == 7 {
		return 0
	}
	return value
}

func monthAliases() map[string]int {
	return map[string]int{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}
}

func weekdayAliases() map[string]int {
	return map[string]int{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}
}
