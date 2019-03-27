package main

import (
	"fmt"
	"strconv"
	"strings"

	runewidth "github.com/mattn/go-runewidth"
)

func Format(table [][]string, fieldNum int, width []int, delimiter string, ignoreIndices []int) []string {
	ignoreMap := map[int]int{}
	for _, ignoreIndex := range ignoreIndices {
		ignoreMap[ignoreIndex] = ignoreIndex
	}
	if len(width) > 0 {
		table = trancateProtrudeString(table, width, "…")
	}
	counts := countFields(table, fieldNum, ignoreMap)
	table = padFields(table, counts, ignoreMap)
	return createDrawLines(table, delimiter)
}

func FormatSimple(table [][]string, fieldNum int) []string {
	return Format(table, fieldNum, []int{}, "", []int{})
}

func splitToTable(str, delimiter string) ([][]string, int) {
	lines := strings.Split(str, "\n")

	var fieldNum int
	table := [][]string{}
	for _, v := range lines {
		if v == "" {
			continue
		}

		var fields []string
		if delimiter != "" {
			fields = strings.Split(v, delimiter)
			for i, column := range fields {
				fields[i] = strings.TrimSpace(column)
			}
		} else {
			fields = strings.Fields(v)
		}

		tmpSize := len(fields)
		if fieldNum < tmpSize {
			fieldNum = tmpSize
		}
		table = append(table, fields)
	}
	return table, fieldNum
}

func countFields(table [][]string, fieldNum int, ignoreMap map[int]int) []int {
	counts := make([]int, fieldNum)
	for _, fields := range table {
		for i, field := range fields {
			if _, ok := ignoreMap[i]; ok {
				continue
			}
			runeLen := runewidth.StringWidth(field)
			if counts[i] < runeLen {
				counts[i] = runeLen
			}
		}
	}
	return counts
}

func padFields(table [][]string, counts []int, ignoreMap map[int]int) [][]string {
	for i, fields := range table {
		for j, field := range fields {
			if counts[j] == 0 {
				continue
			}
			table[i][j] = padRight(field, counts[j])
		}
	}
	return table
}

func padRight(str string, length int) string {
	ws := fmt.Sprintf("%-*s", length-runewidth.StringWidth(str), "")
	return fmt.Sprint(str, ws)
}

func parceWidthFlag(str string) ([]int, error) {
	str = strings.Replace(str, " ", "", -1)
	sp := strings.Split(str, ",")

	limits := make([]int, len(sp))
	for i, limitstr := range sp {
		limit, err := strconv.Atoi(limitstr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit option")
		}
		limits[i] = limit
	}
	return limits, nil
}

func trancateProtrudeString(table [][]string, limits []int, tail string) [][]string {
	for _, fields := range table {
		for i, limit := range limits {
			if len(fields[i]) > limit {
				fields[i] = runewidth.Truncate(fields[i], limit, tail)
			}
		}
	}
	return table
}

func createDrawLines(table [][]string, delimiter string) []string {
	if delimiter == "" {
		delimiter = "\t"
	}

	res := make([]string, len(table))
	for i, fields := range table {
		res[i] = strings.Join(fields, delimiter)
	}
	return res
}
