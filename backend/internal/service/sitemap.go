package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

// sitemapEntry is one <url> in the sitemap. lastMod is optional (zero = omit).
type sitemapEntry struct {
	loc     string
	lastMod time.Time
}

// SitemapService renders the sitemap.xml that tells search engines which pages
// exist. Because the frontend is a SPA, discovery depends on this list (plus
// internal links), so every public, indexable URL is included: the static
// pages, every channel, every post, and every non-banned profile.
type SitemapService struct {
	db      *gorm.DB
	baseURL string
}

func NewSitemapService(db *gorm.DB, baseURL string) *SitemapService {
	return &SitemapService{db: db, baseURL: strings.TrimSuffix(baseURL, "/")}
}

func (s *SitemapService) Build(ctx context.Context) ([]byte, error) {
	entries := []sitemapEntry{
		{loc: s.baseURL + "/"},
		{loc: s.baseURL + "/channels"},
		{loc: s.baseURL + "/channelers"},
	}

	var channels []models.Channel
	if err := s.db.WithContext(ctx).Select("id", "updated_at").Find(&channels).Error; err != nil {
		return nil, err
	}
	for _, ch := range channels {
		entries = append(entries, sitemapEntry{
			loc:     fmt.Sprintf("%s/channels/%d", s.baseURL, ch.ID),
			lastMod: ch.UpdatedAt,
		})
	}

	var posts []models.Post
	if err := s.db.WithContext(ctx).Select("id", "updated_at").Find(&posts).Error; err != nil {
		return nil, err
	}
	for _, p := range posts {
		entries = append(entries, sitemapEntry{
			loc:     fmt.Sprintf("%s/posts/%d", s.baseURL, p.ID),
			lastMod: p.UpdatedAt,
		})
	}

	var users []models.User
	if err := s.db.WithContext(ctx).Where("banned_at IS NULL").Select("id", "updated_at").Find(&users).Error; err != nil {
		return nil, err
	}
	for _, u := range users {
		entries = append(entries, sitemapEntry{
			loc:     fmt.Sprintf("%s/users/%d", s.baseURL, u.ID),
			lastMod: u.UpdatedAt,
		})
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, e := range entries {
		b.WriteString("  <url>\n")
		b.WriteString("    <loc>" + xmlEscape(e.loc) + "</loc>\n")
		if !e.lastMod.IsZero() {
			b.WriteString("    <lastmod>" + e.lastMod.UTC().Format("2006-01-02") + "</lastmod>\n")
		}
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>\n")

	return []byte(b.String()), nil
}

// xmlEscape escapes the five characters that may not appear literally in XML
// text. Sitemap URLs are ASCII normally, but a 404-style id or handler typo
// must never produce malformed XML.
func xmlEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	).Replace(s)
}
