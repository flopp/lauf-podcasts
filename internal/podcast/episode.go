package podcast

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"time"

	"github.com/flopp/lauf-podcasts/internal/utils"
	"github.com/mmcdole/gofeed"
)

type Episode struct {
	Title        string
	Description  template.HTML
	Published    time.Time
	PublishedStr string
	Link         string
	Duration     string
}

var reSeconds = regexp.MustCompile(`^\s*(\d+)\s*$`)

func NormalizeDuration(s string) string {
	if match := reSeconds.FindStringSubmatch(s); match != nil {
		if totalSec, err := strconv.ParseInt(match[1], 10, 64); err == nil {
			s := totalSec % 60
			totalM := totalSec / 60
			m := totalM % 60
			h := totalM / 60
			return fmt.Sprintf("%0d:%02d:%02d", h, m, s)
		}
	}
	return s
}

func CreateFromItem(item *gofeed.Item) *Episode {
	published := time.Time{}
	publishedStr := item.Published
	if item.PublishedParsed != nil {
		published = *item.PublishedParsed
		publishedStr = published.Format("2006-01-02")
	}
	duration := "unbekannt"
	if item.ITunesExt != nil {
		duration = item.ITunesExt.Duration
	}
	episode := &Episode{
		item.Title,
		utils.CreateHTML(CleanDescription(item.Description)),
		published,
		publishedStr,
		item.Link,
		NormalizeDuration(duration),
	}
	return episode
}
