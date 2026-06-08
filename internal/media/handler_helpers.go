package media

import (
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func multipartHeaders(c *fiber.Ctx) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	headers := form.File["files"]
	if len(headers) == 0 {
		headers = form.File["file"]
	}
	if len(headers) == 0 {
		return nil, ErrInvalidUpload
	}
	return headers, nil
}

func parseInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "si", "on":
		return true
	default:
		return false
	}
}
