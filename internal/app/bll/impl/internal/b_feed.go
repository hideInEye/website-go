package internal

import (
	"context"
	"encoding/xml"
	"net/url"
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/config"
	"go-gin/internal/app/model"
)

const feedSiteName = "日落的足迹"
const feedSiteDescription = "把代码、生活和途中看到的光，留在这里。"

func NewFeed(store model.IFeed, cfg *config.Config) *Feed {
	return &Feed{
		store:  store,
		config: cfg,
	}
}

type Feed struct {
	store  model.IFeed
	config *config.Config
}

var _ bll.IFeed = (*Feed)(nil)

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type sitemapDocument struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"`
	Frequency    string `xml:"changefreq"`
	Priority     string `xml:"priority"`
}

func marshalFeedXML(value interface{}) ([]byte, error) {
	encoded, err := xml.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), encoded...), nil
}

func (a *Feed) RSS(ctx context.Context) ([]byte, error) {
	posts, err := a.store.ListPosts(ctx, 50)
	if err != nil {
		return nil, err
	}
	items := make([]rssItem, 0, len(posts))
	for _, post := range posts {
		link := a.config.Site.Origin + "/articles/" + url.PathEscape(post.Slug)
		published := post.CreatedAt
		if post.PublishedAt != nil {
			published = *post.PublishedAt
		}
		items = append(items, rssItem{
			Title:       post.Title,
			Link:        link,
			GUID:        link,
			Description: post.Excerpt,
			PubDate:     published.UTC().Format(time.RFC1123),
		})
	}
	return marshalFeedXML(rssDocument{
		Version: "2.0",
		Channel: rssChannel{
			Title:       feedSiteName,
			Link:        a.config.Site.Origin,
			Description: feedSiteDescription,
			Language:    "zh-CN",
			Items:       items,
		},
	})
}

func (a *Feed) Sitemap(ctx context.Context) ([]byte, error) {
	posts, err := a.store.ListPosts(ctx, 0)
	if err != nil {
		return nil, err
	}
	categories, err := a.store.ListCategorySlugs(ctx)
	if err != nil {
		return nil, err
	}
	tags, err := a.store.ListTagSlugs(ctx)
	if err != nil {
		return nil, err
	}
	base := a.config.Site.Origin
	urls := []sitemapURL{
		{
			Location:  base,
			Frequency: "weekly",
			Priority:  "1",
		},
		{
			Location:  base + "/about",
			Frequency: "monthly",
			Priority:  "0.5",
		},
		{
			Location:  base + "/search",
			Frequency: "weekly",
			Priority:  "0.4",
		},
		{
			Location:  base + "/categories",
			Frequency: "weekly",
			Priority:  "0.6",
		},
		{
			Location:  base + "/tags",
			Frequency: "weekly",
			Priority:  "0.5",
		},
	}
	for _, post := range posts {
		modified := ""
		if post.PublishedAt != nil {
			modified = post.PublishedAt.UTC().Format(time.RFC3339)
		}
		urls = append(urls, sitemapURL{
			Location:     base + "/articles/" + url.PathEscape(post.Slug),
			LastModified: modified,
			Frequency:    "monthly",
			Priority:     "0.8",
		})
	}
	for _, slug := range categories {
		urls = append(urls, sitemapURL{
			Location:  base + "/categories/" + url.PathEscape(slug),
			Frequency: "weekly",
			Priority:  "0.6",
		})
	}
	for _, slug := range tags {
		urls = append(urls, sitemapURL{
			Location:  base + "/tags/" + url.PathEscape(slug),
			Frequency: "weekly",
			Priority:  "0.5",
		})
	}
	return marshalFeedXML(sitemapDocument{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	})
}

func (a *Feed) Robots(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []byte("User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /account\nDisallow: /api\n\nSitemap: " + a.config.Site.Origin + "/sitemap.xml\n"), nil
}

func (a *Feed) LegacyRedirect(ctx context.Context, id string) (string, error) {
	if id != "" {
		slug, err := a.store.FindLegacySlug(ctx, id)
		if err == nil && slug != "" {
			return a.config.Site.Origin + "/articles/" + url.PathEscape(slug), nil
		}
	}
	return a.config.Site.Origin + "/", nil
}
