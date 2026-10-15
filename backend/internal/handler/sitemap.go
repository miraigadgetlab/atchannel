package handler

import (
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type SitemapHandler struct {
	sitemapService *service.SitemapService
}

func NewSitemapHandler(ss *service.SitemapService) *SitemapHandler {
	return &SitemapHandler{sitemapService: ss}
}

// Get serves the public sitemap.xml. It is unauthenticated: crawlers must be
// able to fetch it, and it only lists already-public URLs.
func (h *SitemapHandler) Get(c fiber.Ctx) error {
	body, err := h.sitemapService.Build(c.Context())
	if err != nil {
		return replyInternal(c, err, "Failed to build sitemap")
	}

	c.Set(fiber.HeaderContentType, "application/xml; charset=utf-8")
	return c.Send(body)
}
