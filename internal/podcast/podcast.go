package podcast

import (
	"html/template"
	"strings"
	"time"

	"github.com/flopp/lauf-podcasts/internal/utils"
	"github.com/mmcdole/gofeed"
)

type Podcast struct {
	Feed             *gofeed.Feed
	Slug             string
	Title            string
	Description      template.HTML
	Episodes         []*Episode
	FeedUrl          string
	WebsiteUrl       string
	WebsiteUrl2      string
	LatestPublish    time.Time
	LatestPublishStr string
	LatestEpisode    *Episode
}

func CreateFromFeed(slug string, feedUrl string, websiteUrl string, feed *gofeed.Feed) *Podcast {
	episodes := make([]*Episode, 0, len(feed.Items))
	for _, item := range feed.Items {
		episodes = append(episodes, CreateFromItem(item))
	}
	latestPublishStr := feed.Published
	latestPublish := time.Time{}
	latestEpisode := (*Episode)(nil)
	if len(episodes) > 0 {
		latestEpisode = episodes[0]
		latestPublish = episodes[0].Published
		latestPublishStr = episodes[0].PublishedStr
	}

	websiteUrl2 := ""
	if websiteUrl == "" {
		websiteUrl = feed.Link
	} else if feed.Link != "" {
		if !strings.HasPrefix(websiteUrl, feed.Link) && !strings.HasPrefix(feed.Link, websiteUrl) {
			websiteUrl2 = feed.Link
		}
	}

	podcast := &Podcast{
		feed,
		slug,
		feed.Title,
		utils.CreateHTML(CleanDescription(feed.Description)),
		episodes,
		feedUrl,
		websiteUrl,
		websiteUrl2,
		latestPublish,
		latestPublishStr,
		latestEpisode,
	}
	return podcast
}
