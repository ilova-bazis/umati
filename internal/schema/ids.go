package schema

import (
	"fmt"
	"regexp"
	"strconv"
)

var taskIDPattern = regexp.MustCompile(`^([A-Z][A-Z0-9]*)-(\d+)$`)

type ParsedTaskID struct {
	Prefix string
	Number int
}

func ParseTaskID(id string) (ParsedTaskID, error) {
	matches := taskIDPattern.FindStringSubmatch(id)
	if len(matches) != 3 {
		return ParsedTaskID{}, fmt.Errorf("invalid task id: %q", id)
	}
	number, err := strconv.Atoi(matches[2])
	if err != nil || number <= 0 {
		return ParsedTaskID{}, fmt.Errorf("invalid task id: %q", id)
	}
	return ParsedTaskID{Prefix: matches[1], Number: number}, nil
}

func CompareTaskIDs(a, b string) (int, error) {
	parsedA, err := ParseTaskID(a)
	if err != nil {
		return 0, err
	}
	parsedB, err := ParseTaskID(b)
	if err != nil {
		return 0, err
	}
	if parsedA.Prefix != parsedB.Prefix {
		if parsedA.Prefix < parsedB.Prefix {
			return -1, nil
		}
		return 1, nil
	}
	if parsedA.Number < parsedB.Number {
		return -1, nil
	}
	if parsedA.Number > parsedB.Number {
		return 1, nil
	}
	return 0, nil
}
