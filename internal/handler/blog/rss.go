package blog

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"
)

const rssLimit = 20

type rssFeed struct {
	XMLName xml.Name    `xml:"rss"`
	Version string      `xml:"version,attr"`
	Channel rssChannel  `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Items         []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Category    string `xml:"category,omitempty"`
}

func (h *SSRHandler) RSS(w http.ResponseWriter, r *http.Request) {
	posts, err := h.svc.GetLatestPosts(r.Context(), rssLimit)
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	now := time.Now().UTC().Format(time.RFC1123Z)

	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host

	items := make([]rssItem, 0, len(posts))
	for _, p := range posts {
		date, parseErr := time.Parse("2006-01-02", p.Date)
		if parseErr != nil {
			date = time.Now()
		}
		pubDate := date.UTC().Format(time.RFC1123Z)

		desc, renderErr := h.svc.RenderMarkdown(p.Markdown)
		if renderErr != nil {
			desc = "<p>" + xmlText(p.Markdown) + "</p>"
		}

		category := ""
		if p.Category != nil {
			category = p.Category.Name
		}

		items = append(items, rssItem{
			Title:       p.Title,
			Link:        baseURL + "/posts/" + p.Slug,
			Guid:        baseURL + "/posts/" + p.Slug,
			PubDate:     pubDate,
			Description: desc,
			Category:    category,
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:         "My Blog",
			Link:          baseURL + "/",
			Description:   "Thoughts, stories, and ideas.",
			Language:      "en-us",
			LastBuildDate: now,
			Items:         items,
		},
	}

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		log.Printf("rss marshal error: %v", err)
		h.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(xml.Header + string(output)))
}

func xmlText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
