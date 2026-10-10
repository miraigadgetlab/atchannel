package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

// UploadAvatar accepts a multipart "avatar" file, validates its size and
// image type, stores it on disk and points the account at it. Old avatars are
// left on disk: they are cheap and the URL swap is atomic from the reader's
// point of view.
func (h *ProfileHandler) UploadAvatar(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return replyErr(c, fiber.StatusUnauthorized, err.Error())
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return replyBadRequest(c, "avatar file is required (multipart field \"avatar\")")
	}

	if file.Size > h.maxAvatarBytes {
		return replyErr(c, fiber.StatusRequestEntityTooLarge,
			fmt.Sprintf("Avatar must be at most %d bytes", h.maxAvatarBytes))
	}

	src, err := file.Open()
	if err != nil {
		return replyInternal(c, err, "Failed to read upload")
	}
	defer src.Close()

	// Sniff the real content type from the file header; never trust the
	// client-declared one.
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return replyInternal(c, err, "Failed to read upload")
	}

	ext, ok := avatarExtension(http.DetectContentType(head[:n]))
	if !ok {
		return replyBadRequest(c, "Only PNG, JPEG, GIF and WebP images are allowed")
	}

	name := randomName() + ext
	dest := filepath.Join(h.uploadDir, name)

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return replyInternal(c, err, "Failed to process upload")
	}

	out, err := os.Create(dest)
	if err != nil {
		return replyInternal(c, err, "Failed to store avatar")
	}
	defer out.Close()

	// Enforce the cap again while copying, in case the multipart header lied
	// about the size.
	written, err := io.Copy(out, io.LimitReader(src, h.maxAvatarBytes+1))
	if err != nil {
		os.Remove(dest)
		return replyInternal(c, err, "Failed to store avatar")
	}
	if written > h.maxAvatarBytes {
		os.Remove(dest)
		return replyErr(c, fiber.StatusRequestEntityTooLarge,
			fmt.Sprintf("Avatar must be at most %d bytes", h.maxAvatarBytes))
	}

	user, err := h.userService.UpdateAvatar(c.Context(), userID, "/uploads/"+name)
	if err != nil {
		os.Remove(dest)
		return svcErr(c, err, "Failed to update profile")
	}

	return c.JSON(user)
}

// avatarExtension maps a sniffed content type to its filename extension.
func avatarExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/png":
		return ".png", true
	case "image/jpeg":
		return ".jpg", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

// randomName returns 16 random bytes as hex: 2^128 distinct avatar names, so
// guessing someone else's is not a realistic concern.
func randomName() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand only fails if the entropy source is broken; fall back
		// to a timestamped name rather than crash on an upload.
		return fmt.Sprintf("fallback-%d", os.Getpid())
	}
	return hex.EncodeToString(buf)
}
