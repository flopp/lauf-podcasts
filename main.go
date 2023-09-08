package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flopp/lauf-podcasts/internal/podcast"
	"github.com/flopp/lauf-podcasts/internal/utils"
	"github.com/mmcdole/gofeed"
	"golang.org/x/image/draw"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type CommandLineOptions struct {
	configFile string
	cacheDir   string
	outDir     string
}

const (
	usage = `USAGE: %s [OPTIONS...]

OPTIONS:
`
)

func parseCommandLine() CommandLineOptions {
	configFile := flag.String("config", "", "select config file")
	cacheDir := flag.String("cache", ".cache", "cache directory")
	outDir := flag.String("out", ".out", "output directory")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *configFile == "" {
		panic("You have to specify a config file, e.g. -config myconfig.json")
	}

	return CommandLineOptions{
		*configFile,
		*cacheDir,
		*outDir,
	}
}

type ApiConfig struct {
	ApiKey  string `json:"api_key"`
	SheetId string `json:"sheet_id"`
}

func readApiConfig(fileName string) (ApiConfig, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return ApiConfig{}, err
	}

	var config ApiConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ApiConfig{}, err
	}

	return config, nil
}

type Config struct {
	api      ApiConfig
	options  CommandLineOptions
	minMtime time.Time
}

func fetchPodcasts(config Config, srv *sheets.Service) []*podcast.Podcast {
	resp, err := srv.Spreadsheets.Values.Get(config.api.SheetId, "LIST!A2:Z").Do()
	check(err)
	if len(resp.Values) == 0 {
		panic("No podcast data found.")
	}

	podcasts := make([]*podcast.Podcast, 0, len(resp.Values))

	for _, row := range resp.Values {
		var name, websiteUrl, feedUrl string
		ll := len(row)
		if ll > 0 {
			name = fmt.Sprintf("%v", row[0])
		}
		if ll > 1 {
			feedUrl = fmt.Sprintf("%v", row[1])
		}
		if ll > 2 {
			websiteUrl = fmt.Sprintf("%v", row[2])
		}

		if ll > 3 {
			log.Printf("podcast '%s': too many non-empty columns (%d)", name, ll)
		}

		sanitizedName := utils.SanitizeName(name)
		if sanitizedName == "" {
			log.Printf("podcast '%s': sanitized name is empty => skip", name)
			continue
		}
		if feedUrl == "" {
			log.Printf("podcast '%s': feed url is empty => skip", name)
			continue
		}

		fileName := fmt.Sprintf("%s/%s.feed", config.options.cacheDir, sanitizedName)
		err := utils.DownloadIfOutdated(feedUrl, fileName, config.minMtime)
		if err != nil {
			log.Printf("podcast '%s': failed to download '%s' to '%s': %v", name, feedUrl, fileName, err)
			continue
		}

		file, _ := os.Open(fileName)
		defer file.Close()
		fp := gofeed.NewParser()
		feed, err := fp.Parse(file)
		if feed == nil {
			log.Printf("podcast '%s': failed to parse feed '%s': %v", name, fileName, err)
			continue
		}

		podcasts = append(podcasts, podcast.CreateFromFeed(sanitizedName, feedUrl, websiteUrl, feed))
	}

	sort.Slice(podcasts, func(i, j int) bool {
		return podcasts[i].LatestPublish.After(podcasts[j].LatestPublish)
	})

	return podcasts
}

func createCoverImage(config Config, slug string, url string) error {
	fileName := fmt.Sprintf("%s/%s.cover", config.options.cacheDir, slug)
	err := utils.DownloadIfOutdated(url, fileName, config.minMtime)
	if err != nil {
		log.Printf("podcast '%s': failed to download '%s' to '%s': %v", slug, url, fileName, err)
		return err
	}

	input, _ := os.Open(fileName)
	defer input.Close()

	var src image.Image
	urlLower := strings.ToLower(url)
	if strings.Contains(urlLower, ".jpg") || strings.Contains(urlLower, ".jpeg") {
		src, err = jpeg.Decode(input)
		if err != nil {
			log.Printf("podcast '%s': failed to decode JPG '%s': %v", slug, url, err)
			return err
		}
	} else if strings.Contains(urlLower, ".png") {
		src, err = png.Decode(input)
		if err != nil {
			log.Printf("podcast '%s': failed to decode PNG '%s': %v", slug, url, err)
			return err
		}
	} else {
		log.Printf("podcast '%s': failed to detect format for '%s'", slug, url)
		return fmt.Errorf("podcast '%s': failed to detect format for '%s'", slug, url)
	}

	outFileName := fmt.Sprintf("%s/%s/cover.jpg", config.options.outDir, slug)
	err = utils.EnsureDir(filepath.Dir(outFileName))
	if err != nil {
		log.Printf("podcast '%s': failed to create folder for cover image '%s': %v", slug, outFileName, err)
		return err
	}
	out, err := os.Create(outFileName)
	if err != nil {
		log.Printf("podcast '%s': failed to create file for cover image '%s': %v", slug, outFileName, err)
		return err
	}
	defer out.Close()

	dst := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	err = jpeg.Encode(out, dst, &jpeg.Options{Quality: 90})
	if err != nil {
		log.Printf("podcast '%s': failed to encode cover image '%s': %v", slug, outFileName, err)
	}
	return err
}

func useDefaultCoverImage(config Config, slug string) error {
	outFileName := fmt.Sprintf("%s/%s/cover.jpg", config.options.outDir, slug)
	err := utils.EnsureDir(filepath.Dir(outFileName))
	if err != nil {
		log.Printf("podcast '%s': failed to create folder for cover image '%s': %v", slug, outFileName, err)
		return err
	}
	out, err := os.Create(outFileName)
	if err != nil {
		log.Printf("podcast '%s': failed to create file for cover image '%s': %v", slug, outFileName, err)
		return err
	}
	defer out.Close()

	source, err := os.Open("default-cover.jpg")
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(out, source)
	return err
}

var templates = make(map[string]*template.Template)

func loadTemplate(name string) *template.Template {
	t, ok := templates[name]
	if ok {
		return t
	}
	t, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", name), "templates/header.html", "templates/footer.html")
	check(err)
	templates[name] = t
	return t
}

type TemplateData struct {
	Title     string
	Canonical string
	Podcasts  []*podcast.Podcast
	Podcast   *podcast.Podcast
}

func executeTemplate(templateName string, fileName string, title string, canonical string, podcasts []*podcast.Podcast, podcast *podcast.Podcast) {
	err := utils.EnsureDir(filepath.Dir(fileName))
	check(err)
	out, err := os.Create(fileName)
	check(err)
	defer out.Close()
	err = loadTemplate(templateName).Execute(out, TemplateData{title, canonical, podcasts, podcast})
	check(err)
}

func main() {
	now := time.Now()
	// today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// timestamp := now.Format("2006-01-02")
	// timestampFull := now.Format("2006-01-02 15:04:05")
	// sheetUrl := ""
	options := parseCommandLine()

	apiConfig, err := readApiConfig(options.configFile)
	check(err)

	config := Config{
		apiConfig,
		options,
		now.AddDate(0, 0, -1),
	}

	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithAPIKey(config.api.ApiKey))
	check(err)

	podcasts := fetchPodcasts(config, srv)
	for _, podcast := range podcasts {
		if podcast.Feed.Image != nil {
			err := createCoverImage(config, podcast.Slug, podcast.Feed.Image.URL)
			if err != nil {
				useDefaultCoverImage(config, podcast.Slug)
			}
		} else {
			if err := useDefaultCoverImage(config, podcast.Slug); err != nil {
				check(err)
			}
		}
	}

	base := "https://lauf-podcasts.flopp.net"
	sitemap := utils.CreateSitemap(base, "2023-09-07")

	sitemap.Add("/", podcasts[0].LatestPublish)
	sitemap.Add("/info.html", utils.MustGetMtime("templates/info.html"))
	sitemap.Add("/impressum.html", utils.MustGetMtime("templates/impressum.html"))

	executeTemplate("index", fmt.Sprintf("%s/index.html", config.options.outDir), "Lauf Podcasts", fmt.Sprintf("%s/", base), podcasts, nil)
	executeTemplate("info", fmt.Sprintf("%s/info.html", config.options.outDir), "Info | Lauf Podcasts", fmt.Sprintf("%s/info.html", base), podcasts, nil)
	executeTemplate("impressum", fmt.Sprintf("%s/impressum.html", config.options.outDir), "Impressum | Lauf Podcasts", fmt.Sprintf("%s/impressum.html", base), podcasts, nil)
	for _, podcast := range podcasts {
		executeTemplate("podcast", fmt.Sprintf("%s/%s/index.html", config.options.outDir, podcast.Slug), fmt.Sprintf("%s | Lauf Podcasts", podcast.Title), fmt.Sprintf("%s/%s/", base, podcast.Slug), podcasts, podcast)
		sitemap.Add(fmt.Sprintf("/%s/", podcast.Slug), podcast.LatestPublish)
	}

	sitemap.Render(fmt.Sprintf("%s/sitemap.xml", config.options.outDir))
}
