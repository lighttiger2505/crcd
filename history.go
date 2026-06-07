package main

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
)

func history(c *cli.Context) error {
	dbPath, err := getHistoryPath(runtime.GOOS)
	if err != nil {
		return err
	}

	var minTime *time.Time
	if c.String("range") != "" {
		year, month, day, err := parseDate(c.String("range"))
		if err != nil {
			return err
		}
		d := time.Now().AddDate(-1*year, -1*month, -1*day)
		minTime = &d
	}

	// Retrieve Chrome browser history.
	histories, err := selectHistory(dbPath, minTime)
	if err != nil {
		return err
	}

	// Sort by frecency and get precomputed scores in one pass.
	histories, scores := sortByFrecency(histories)

	// Determine the column width for the left-aligned score column.
	scoreWidth := 0
	for _, s := range scores {
		if w := len(fmt.Sprintf("%d", s)); w > scoreWidth {
			scoreWidth = w
		}
	}

	// Build display lines for FZF.
	lines := []string{}
	for i, b := range histories {
		score := color.WhiteString(fmt.Sprintf("%-*d", scoreWidth, scores[i]))
		title := color.YellowString(b.Title)
		url := color.HiBlackString(b.URL)
		line := fmt.Sprintf("%s  %s  %s", score, title, url)
		lines = append(lines, line)
	}

	// Run FZF. Keep frecency order (--no-sort) and exclude the score column from search (--nth=2..).
	selectedURL, err := fzfOpen(lines, "--no-sort", "--nth=2..")
	if err != nil {
		return err
	}

	if selectedURL != "" {
		if err := openbrowser(selectedURL); err != nil {
			return err
		}
	}
	return nil
}

var unitMap = map[string]string{
	"D": "Day",
	"M": "Month",
	"Y": "Year",
	"d": "Day",
	"m": "Month",
	"y": "Year",
}

func parseDate(s string) (int, int, int, error) {
	// ([0-9]+[a-z]+)+
	orig := s
	var timeDeltaMap = map[string]int{}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, 0, 0, nil
	}
	if s == "" {
		return 0, 0, 0, errors.New("invalid date " + orig)
	}
	for s != "" {
		var v int64
		var err error

		// The next character must be [0-9]
		if !('0' <= s[0] && s[0] <= '9') {
			return 0, 0, 0, errors.New("invalid date " + orig)
		}

		// Consume [0-9]*
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, 0, 0, errors.New("invalid duration " + orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, 0, 0, errors.New("missing unit in date " + orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, 0, 0, errors.New("unknown unit " + u + " in duration " + orig)
		}
		timeDeltaMap[unit] = int(v)
	}

	return timeDeltaMap["Year"], timeDeltaMap["Month"], timeDeltaMap["Day"], nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

type RecencyModifier struct {
	AgeHours float64
	Score    int
}

var recencyModifiers = []*RecencyModifier{
	// within 4 hours
	{AgeHours: 4, Score: 100},
	// within 1 day
	{AgeHours: 24, Score: 80},
	// within 3 days
	{AgeHours: 24 * 3, Score: 60},
	// within 1 week
	{AgeHours: 24 * 7, Score: 40},
	// within 1 month
	{AgeHours: 24 * 30, Score: 20},
	// within 90 days
	{AgeHours: 24 * 90, Score: 10},
}

// getRecencyWeight returns a recency score based on how long ago lastVisit occurred.
func getRecencyWeight(lastVisit time.Time) int {
	now := time.Now()
	duration := now.Sub(lastVisit)
	for _, m := range recencyModifiers {
		if duration.Hours() < m.AgeHours {
			return m.Score
		}
	}
	return 5
}

// calculateFrecency computes frecency as visit_count multiplied by the average
// recency weight across sampled individual visits. This gives more accurate
// results than weighting only the last visit time.
func calculateFrecency(history *History) int {
	if len(history.Visits) == 0 {
		return 0
	}
	sum := 0
	for _, v := range history.Visits {
		sum += getRecencyWeight(v)
	}
	return history.VisitCount * sum / len(history.Visits)
}

// sortByFrecency sorts histories by frecency score (highest first) in a single
// pass and returns the sorted slice together with the precomputed scores so
// callers do not need to recalculate.
func sortByFrecency(histories []*History) ([]*History, []int) {
	type scored struct {
		h     *History
		score int
	}
	pairs := make([]scored, len(histories))
	for i, h := range histories {
		pairs[i] = scored{h: h, score: calculateFrecency(h)}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})
	result := make([]*History, len(pairs))
	scores := make([]int, len(pairs))
	for i, p := range pairs {
		result[i] = p.h
		scores[i] = p.score
	}
	return result, scores
}
