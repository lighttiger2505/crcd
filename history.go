package main

import (
	"errors"
	"fmt"
	"io"
	"os"
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

	readFilePath, err := copyHisotryDB(dbPath)
	if err != nil {
		return err
	}
	defer os.Remove(readFilePath)

	lastdate := ""
	if c.String("range") != "" {
		year, month, day, err := parseDate(c.String("range"))
		if err != nil {
			return err
		}
		d := time.Now().AddDate(-1*year, -1*month, -1*day)
		lastdate = d.Format("2006-01-02 15:04:05")
	}

	// GoogleChromeのブラウザ履歴を取得
	histories, err := selectHistory(readFilePath, lastdate)
	if err != nil {
		return err
	}

	// Frecencyアルゴリズムでソート
	histories = sortByFrecency(histories)

	// FZFで表示するための文字列を作成
	lines := []string{}
	for _, b := range histories {
		title := color.YellowString(b.Title)
		lastVisitTime := b.LastVisitTime.Format(time.RFC3339)
		url := color.HiBlackString(b.URL)
		line := fmt.Sprintf("%s (%s)\n%s", title, lastVisitTime, url)
		lines = append(lines, line)
	}

	// FZFを実行して選択
	// Frequencyアルゴリズムのソートを維持するため、FZFのソートはしない
	selectedURL, err := fzfOpen(lines, "--no-sort")
	if err != nil {
		return err
	}

	// 取得したURLをブラウザで開く
	if selectedURL != "" {
		if err := openbrowser(selectedURL); err != nil {
			return err
		}
	}
	return nil
}

func copyHisotryDB(dbPath string) (string, error) {
	dbFile, err := os.Open(dbPath)
	if err != nil {
		return "", err
	}
	tmpFile, err := os.CreateTemp("/tmp", "bhb")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(tmpFile, dbFile); err != nil {
		return "", err
	}
	readFilePath := tmpFile.Name()
	dbFile.Close()
	tmpFile.Close()

	return readFilePath, nil
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
	// 4時間以内
	{AgeHours: 4, Score: 100},
	// 1日以内
	{AgeHours: 24, Score: 80},
	// 3日以内
	{AgeHours: 24 * 3, Score: 60},
	// 1週間以内
	{AgeHours: 24 * 7, Score: 40},
	// 1ヶ月以内
	{AgeHours: 24 * 30, Score: 20},
	// 90日以内
	{AgeHours: 24 * 90, Score: 10},
}

// Recencyの重みを決定する関数
func getRecencyWeight(lastVisit time.Time) int {
	now := time.Now()
	duration := now.Sub(lastVisit)
	for _, recentModifyer := range recencyModifiers {
		if duration.Hours() < recentModifyer.AgeHours {
			return recentModifyer.Score
		}
	}
	return 5
}

// Frecencyスコアを計算する関数
func calculateFrecency(history *History) int {
	recencyWeight := getRecencyWeight(history.LastVisitTime)
	return history.VisitCount * recencyWeight
}

// Frecencyに基づいてソートする関数
func sortByFrecency(histories []*History) []*History {
	sort.Slice(histories, func(i, j int) bool {
		return calculateFrecency(histories[i]) > calculateFrecency(histories[j])
	})
	return histories
}
