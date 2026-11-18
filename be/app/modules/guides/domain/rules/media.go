// Package rules holds the guides module's pure rules.
package rules

import (
	"strings"

	"ielts-practice/app/modules/guides/domain/entity"
)

// KeyForFile maps a media file name (<sha256>.mp4 or <sha256>.jpg) to its object key in the bucket.
func KeyForFile(file string) string {
	if strings.HasSuffix(file, ".mp4") {
		return entity.VideoKeyPrefix + file
	}
	return entity.PosterKeyPrefix + file
}
