package srt

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// SRT字幕条目
type SrtEntry struct {
	ID    int
	Start int64 // 微秒
	End   int64 // 微秒
	Text  string
}

// 解析SRT时间格式为微秒
func ParseSrtTime(timeStr string) (int64, error) {
	timeParts := strings.Split(timeStr, ",")
	hms := strings.Split(timeParts[0], ":")
	ms, _ := strconv.Atoi(timeParts[1])

	hour, _ := strconv.Atoi(hms[0])
	minute, _ := strconv.Atoi(hms[1])
	second, _ := strconv.Atoi(hms[2])

	totalMicros := int64(hour*3600+minute*60+second)*1000000 + int64(ms)*1000
	return totalMicros, nil
}

// 解析SRT文件
func ParseSrtFile(filePath string) ([]SrtEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []SrtEntry
	scanner := bufio.NewScanner(file)

	currentEntry := SrtEntry{}
	state := 0 // 0: ID, 1: Time, 2: Text

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			if currentEntry.ID != 0 {
				entries = append(entries, currentEntry)
				currentEntry = SrtEntry{}
				state = 0
			}
			continue
		}

		switch state {
		case 0: // ID
			id, err := strconv.Atoi(line)
			if err != nil {
				continue
			}
			currentEntry.ID = id
			state = 1
		case 1: // Time
			timeParts := strings.Split(line, " --> ")
			if len(timeParts) == 2 {
				start, err1 := ParseSrtTime(timeParts[0])
				end, err2 := ParseSrtTime(timeParts[1])
				if err1 == nil && err2 == nil {
					currentEntry.Start = start
					currentEntry.End = end
				}
			}
			state = 2
		case 2: // Text
			if currentEntry.Text == "" {
				currentEntry.Text = line
			} else {
				currentEntry.Text += "\n" + line
			}
		}
	}

	// 添加最后一个条目
	if currentEntry.ID != 0 {
		entries = append(entries, currentEntry)
	}

	return entries, nil
}