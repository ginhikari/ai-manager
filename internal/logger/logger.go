package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogManager struct {
	logDir string
}

func NewLogManager(logDir string) *LogManager {
	return &LogManager{logDir: logDir}
}

func (l *LogManager) GetLogPath(processID string) string {
	return filepath.Join(l.logDir, processID+".log")
}

func (l *LogManager) GetLogs(processID string) ([]string, error) {
	path := l.GetLogPath(processID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func (l *LogManager) Tail(processID string, n int) ([]string, error) {
	path := l.GetLogPath(processID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if n > len(allLines) {
		n = len(allLines)
	}

	start := len(allLines) - n
	if start < 0 {
		start = 0
	}
	return allLines[start:], nil
}

func (l *LogManager) Search(processID string, query string) ([]string, error) {
	entries, err := l.GetLogs(processID)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, line := range entries {
		if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
			matches = append(matches, line)
		}
	}
	return matches, nil
}

func (l *LogManager) Clear(processID string) error {
	path := l.GetLogPath(processID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *LogManager) ListLogs() ([]LogEntry, error) {
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var logs []LogEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		logs = append(logs, LogEntry{
			Name:      e.Name(),
			Size:      info.Size(),
			Modified:  info.ModTime(),
			ProcessID: strings.TrimSuffix(e.Name(), ".log"),
		})
	}
	return logs, nil
}

func (l *LogManager) WriteLog(processID string, message string) error {
	path := l.GetLogPath(processID)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(f, "[%s] %s\n", timestamp, message)
	return err
}

type LogEntry struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	ProcessID string    `json:"process_id"`
}
