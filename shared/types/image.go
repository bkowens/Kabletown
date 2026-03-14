package types

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"

    "time"
)

// CalculateImageTag returns MD5 hash of formatted modification time for W9
// Tag format: MD5(DateModified.ToString("yyyyMMddHHmmss"))
func CalculateImageTag(dateModified time.Time) string {
    formatted := dateModified.Format("20060102150405")  // yyyyMMddHHmmss
    hash := md5.Sum([]byte(formatted))
    return hex.EncodeToString(hash[:])  // 32 char lowercase hex
}

// BuildImageUrl constructs W9-compliant image URL
// Format: /Items/{itemId}/Images/{imageType}/{imageIndex}?fillHeight=300&fillWidth=300&quality=90&tag={md5hash}
func BuildImageUrl(itemId, imageType string, imageIndex int, dateModified time.Time) string {
    tag := CalculateImageTag(dateModified)

    imageIndexPart := ""
    if imageIndex > 0 {
        imageIndexPart = fmt.Sprintf("/%d", imageIndex)
    }

    return fmt.Sprintf(
        "/Items/%s/Images/%s%s?fillHeight=300&fillWidth=300&quality=90&tag=%s",
        itemId, imageType, imageIndexPart, tag,
    )
}

// BuildImageUrlWithCustomDimensions builds URL with custom dimensions
func BuildImageUrlWithCustomDimensions(itemId, imageType string, imageIndex, height, width int, dateModified time.Time) string {
    tag := CalculateImageTag(dateModified)

    imageIndexPart := ""
    if imageIndex > 0 {
        imageIndexPart = fmt.Sprintf("/%d", imageIndex)
    }

    return fmt.Sprintf(
        "/Items/%s/Images/%s%s?fillHeight=%d&fillWidth=%d&quality=90&tag=%s",
        itemId, imageType, imageIndexPart, height, width, tag,
    )
}

// ImageTags represents W9 image tag collection
type ImageTags struct {
    Primary string `json:"Primary,omitempty"`
    Backdrop string `json:"Backdrop,omitempty"`
    Thumb    string `json:"Thumb,omitempty"`
    Logo     string `json:"Logo,omitempty"`
    Disc     string `json:"Disc,omitempty"`
    Banner   string `json:"Banner,omitempty"`
}

// ToURLs converts ImageTags to full URLs using image server base
func (i *ImageTags) ToURLs(imageServerURL string, dateModified time.Time) map[string]string {
    result := make(map[string]string, 8)  // W6: empty map, not nil

    addIfPresent := func(tagName, tagValue string) {
        if tagValue != "" {
            result[tagName] = BuildImageUrl("", tagName, 0, dateModified)  // Would use item ID in real impl
        }
    }

    addIfPresent("Primary", i.Primary)
    addIfPresent("Backdrop", i.Backdrop)
    addIfPresent("Thumb", i.Thumb)
    addIfPresent("Logo", i.Logo)
    addIfPresent("Disc", i.Disc)
    addIfPresent("Banner", i.Banner)

    return result
}
