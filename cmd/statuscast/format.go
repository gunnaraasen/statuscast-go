package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v3"

	statuscast "statuscast-go"
)

func useJSON(cmd *cli.Command) bool {
	return cmd.Bool("json")
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

func printTable(headers []string, rows [][]string) {
	w := newTabWriter()
	_, _ = fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		_, _ = fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	_ = w.Flush()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

func formatOptTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return formatTime(*t)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04",
	"2006-01-02",
}

func getPagination(cmd *cli.Command) statuscast.Pagination {
	return statuscast.Pagination{Page: int(cmd.Int("page")), PerPage: int(cmd.Int("per-page"))}
}

func toChannels(ss []string) []statuscast.NotificationChannel {
	ch := make([]statuscast.NotificationChannel, len(ss))
	for i, s := range ss {
		ch[i] = statuscast.NotificationChannel(s)
	}
	return ch
}

func channelStrings(channels []statuscast.NotificationChannel) []string {
	ss := make([]string, len(channels))
	for i, ch := range channels {
		ss[i] = string(ch)
	}
	return ss
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range timeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q; accepted formats: RFC3339, 2006-01-02T15:04, 2006-01-02", s)
}
