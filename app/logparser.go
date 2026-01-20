package app

import (
	"regexp"
	"strings"
)

// StepLog represents a parsed step with its log lines
type StepLog struct {
	Name      string   // Step name extracted from ##[group]
	Lines     []string // Log lines for this step
	StartLine int      // Starting line number in the original logs
	EndLine   int      // Ending line number in the original logs
}

// ParsedLogs represents the parsed structure of GitHub Actions logs
type ParsedLogs struct {
	Steps    []StepLog // Parsed steps
	RawLogs  string    // Original raw logs
	AllLines []string  // All lines split from raw logs
}

// groupStartRegex matches ##[group]<step name>
var groupStartRegex = regexp.MustCompile(`##\[group\](.+)$`)

// groupEndRegex matches ##[endgroup]
var groupEndRegex = regexp.MustCompile(`##\[endgroup\]`)

// timestampRegex matches ISO 8601 timestamps at the start of log lines
var timestampRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T(\d{2}:\d{2}:\d{2})\.\d+Z)\s*`)

// ParseLogs parses GitHub Actions log output and extracts steps
func ParseLogs(rawLogs string) *ParsedLogs {
	parsed := &ParsedLogs{
		RawLogs:  rawLogs,
		Steps:    []StepLog{},
		AllLines: []string{},
	}

	if rawLogs == "" {
		return parsed
	}

	lines := strings.Split(rawLogs, "\n")
	parsed.AllLines = lines

	var currentStep *StepLog
	var inGroup bool

	for i, line := range lines {
		// Check for group start
		if match := groupStartRegex.FindStringSubmatch(line); match != nil {
			// Close previous unclosed group if any
			if currentStep != nil && inGroup {
				currentStep.EndLine = i - 1
				parsed.Steps = append(parsed.Steps, *currentStep)
			}

			currentStep = &StepLog{
				Name:      match[1],
				Lines:     []string{line},
				StartLine: i,
			}
			inGroup = true
			continue
		}

		// Check for group end
		if groupEndRegex.MatchString(line) {
			if currentStep != nil && inGroup {
				currentStep.Lines = append(currentStep.Lines, line)
				currentStep.EndLine = i
				parsed.Steps = append(parsed.Steps, *currentStep)
				currentStep = nil
				inGroup = false
			}
			continue
		}

		// Add line to current step if in a group
		if currentStep != nil && inGroup {
			currentStep.Lines = append(currentStep.Lines, line)
		}
	}

	// Handle unclosed group (still running)
	if currentStep != nil && inGroup {
		currentStep.EndLine = len(lines) - 1
		parsed.Steps = append(parsed.Steps, *currentStep)
	}

	return parsed
}

// GetStepLogs returns the log content for a specific step
// stepIndex = -1 returns all logs, otherwise returns the specific step's logs
func (p *ParsedLogs) GetStepLogs(stepIndex int) string {
	if p == nil {
		return ""
	}

	// -1 means all logs
	if stepIndex == -1 {
		return p.RawLogs
	}

	// Check bounds
	if stepIndex < 0 || stepIndex >= len(p.Steps) {
		return ""
	}

	return strings.Join(p.Steps[stepIndex].Lines, "\n")
}

// FormatLogLine formats a single log line by simplifying the timestamp
// Input: "2024-01-15T10:00:00.000Z Some message"
// Output: "10:00:00 Some message"
func FormatLogLine(line string) string {
	if line == "" {
		return ""
	}

	match := timestampRegex.FindStringSubmatch(line)
	if match == nil {
		return line
	}

	// match[1] is the full timestamp, match[2] is HH:MM:SS
	timeOnly := match[2]
	rest := strings.TrimPrefix(line, match[0])
	return timeOnly + " " + rest
}

// FormatStepLogs formats all lines in the logs with simplified timestamps
func (p *ParsedLogs) FormatStepLogs(stepIndex int) string {
	logs := p.GetStepLogs(stepIndex)
	if logs == "" {
		return ""
	}

	lines := strings.Split(logs, "\n")
	formatted := make([]string, len(lines))
	for i, line := range lines {
		formatted[i] = FormatLogLine(line)
	}
	return strings.Join(formatted, "\n")
}

// GitHub Actions marker regexes
var (
	errorMarkerRegex   = regexp.MustCompile(`##\[error\]`)
	warningMarkerRegex = regexp.MustCompile(`##\[warning\]`)
	noticeMarkerRegex  = regexp.MustCompile(`##\[notice\]`)
	errorKeywordRegex  = regexp.MustCompile(`(?i)\b(error|failed|failure|panic)\b`)
	warnKeywordRegex   = regexp.MustCompile(`(?i)\b(warning|warn)\b`)
	successKeywordRegex = regexp.MustCompile(`(?i)\b(success|passed|ok)\b`)
)

// FormatLogLineWithColor applies syntax highlighting to a log line
// It colors timestamps, GitHub Actions markers, and error/warning keywords
func FormatLogLineWithColor(line string) string {
	if line == "" {
		return ""
	}

	// First, simplify and extract timestamp
	var timestamp, rest string
	match := timestampRegex.FindStringSubmatch(line)
	if match != nil {
		timestamp = match[2] // HH:MM:SS
		rest = strings.TrimPrefix(line, match[0])
	} else {
		rest = line
	}

	// Check for GitHub Actions markers (these color the entire line)
	if errorMarkerRegex.MatchString(rest) {
		if timestamp != "" {
			return LogTimestampStyle.Render(timestamp) + " " + LogErrorStyle.Render(rest)
		}
		return LogErrorStyle.Render(rest)
	}
	if warningMarkerRegex.MatchString(rest) {
		if timestamp != "" {
			return LogTimestampStyle.Render(timestamp) + " " + LogWarningStyle.Render(rest)
		}
		return LogWarningStyle.Render(rest)
	}
	if noticeMarkerRegex.MatchString(rest) {
		if timestamp != "" {
			return LogTimestampStyle.Render(timestamp) + " " + LogNoticeStyle.Render(rest)
		}
		return LogNoticeStyle.Render(rest)
	}
	if groupStartRegex.MatchString(rest) {
		if timestamp != "" {
			return LogTimestampStyle.Render(timestamp) + " " + LogGroupStyle.Render(rest)
		}
		return LogGroupStyle.Render(rest)
	}
	if groupEndRegex.MatchString(rest) {
		if timestamp != "" {
			return LogTimestampStyle.Render(timestamp) + " " + LogEndGroupStyle.Render(rest)
		}
		return LogEndGroupStyle.Render(rest)
	}

	// Apply keyword highlighting to the rest of the line
	rest = highlightKeywords(rest)

	// Combine timestamp and rest
	if timestamp != "" {
		return LogTimestampStyle.Render(timestamp) + " " + rest
	}
	return rest
}

// highlightKeywords applies color to error/warning/success keywords in text
func highlightKeywords(text string) string {
	// Apply error keywords (red)
	text = errorKeywordRegex.ReplaceAllStringFunc(text, func(match string) string {
		return LogErrorKeyword.Render(match)
	})

	// Apply warning keywords (orange)
	text = warnKeywordRegex.ReplaceAllStringFunc(text, func(match string) string {
		return LogWarningKeyword.Render(match)
	})

	// Apply success keywords (green)
	text = successKeywordRegex.ReplaceAllStringFunc(text, func(match string) string {
		return LogSuccessKeyword.Render(match)
	})

	return text
}

// FormatStepLogsWithColor formats all lines with syntax highlighting
func (p *ParsedLogs) FormatStepLogsWithColor(stepIndex int) string {
	logs := p.GetStepLogs(stepIndex)
	if logs == "" {
		return ""
	}

	lines := strings.Split(logs, "\n")
	formatted := make([]string, len(lines))
	for i, line := range lines {
		formatted[i] = FormatLogLineWithColor(line)
	}
	return strings.Join(formatted, "\n")
}
