package uid

import (
	"encoding/base64"

	"github.com/satori/go.uuid"
)

// NewID 创建ID
func NewID() string {
	id, _ := uuid.NewV4()
	b64 := base64.URLEncoding.EncodeToString(id.Bytes()[:12])
	return b64
}
