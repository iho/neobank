package pagination

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// Cursor identifies the last item from a previous page (created_at + id).
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

func Encode(c Cursor) string {
	if c.ID == "" || c.CreatedAt.IsZero() {
		return ""
	}
	raw := fmt.Sprintf("%s|%s", c.CreatedAt.UTC().Format(time.RFC3339Nano), c.ID)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func Decode(s string) (Cursor, error) {
	if s == "" {
		return Cursor{}, nil
	}
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor")
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 || parts[1] == "" {
		return Cursor{}, fmt.Errorf("invalid cursor")
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor")
	}
	return Cursor{CreatedAt: at.UTC(), ID: parts[1]}, nil
}