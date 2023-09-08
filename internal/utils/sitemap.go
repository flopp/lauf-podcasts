package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SitemapEntry struct {
	Url       string
	Timestamp time.Time
}

type Sitemap struct {
	Base         string
	MinTimestamp string
	Entries      []*SitemapEntry
}

func CreateSitemap(base string, minTimestamp string) *Sitemap {
	return &Sitemap{base, minTimestamp, make([]*SitemapEntry, 0)}
}

func (sitemap *Sitemap) Add(url string, timestamp time.Time) {
	sitemap.Entries = append(sitemap.Entries, &SitemapEntry{url, timestamp})
}

func (sitemap Sitemap) Render(fileName string) error {
	err := EnsureDir(filepath.Dir(fileName))
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	f.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	for _, entry := range sitemap.Entries {
		t := formatTimestamp(entry.Timestamp)
		if t < sitemap.MinTimestamp {
			t = sitemap.MinTimestamp
		}
		f.WriteString("    <url>\n")
		f.WriteString(fmt.Sprintf("        <loc>%s%s</loc>\n", sitemap.Base, entry.Url))
		f.WriteString(fmt.Sprintf("        <lastmod>%s</lastmod>\n", t))
		f.WriteString("    </url>\n")
	}
	f.WriteString("</urlset>\n")

	return nil
}

func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02")
}

func GetTimestamp(filePath string) string {
	stat, err := os.Stat(filePath)
	if err != nil {
		return ""
	}
	return formatTimestamp(stat.ModTime())
}
