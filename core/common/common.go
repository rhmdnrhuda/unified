package common

import (
	"encoding/base64"
	"fmt"
	_ "github.com/gin-gonic/gin"
	"github.com/rhmdnrhuda/unified/core/entity"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func PrepareMessage(req entity.MessageRequest, message, messageType string) entity.AdaRequest {
	if messageType == "" {
		messageType = "text"
	}

	return entity.AdaRequest{
		Platform:     req.Platform,
		From:         req.AccountNo,
		To:           req.FromNo,
		Type:         messageType,
		Text:         message,
		TemplateLang: "en",
	}
}

func PrepareStickerMessage(req entity.MessageRequest, message string) entity.AdaRequest {
	return entity.AdaRequest{
		Platform:  req.Platform,
		From:      req.AccountNo,
		To:        req.FromNo,
		Type:      "sticker",
		StickerID: message,
	}
}

func PrepareMessageButton(req entity.MessageRequest, message, header, footer string, buttons []string) entity.AdaButtonRequest {
	return entity.AdaButtonRequest{
		Platform:   req.Platform,
		From:       req.AccountNo,
		To:         req.FromNo,
		Text:       message,
		HeaderType: "text",
		Header:     header,
		Footer:     footer,
		Buttons:    buttons,
	}
}

func IsIgnore(str string) bool {
	listIgnore := []string{"ok", "okay", "kk", "oke", "okey", "hmm", "hm", "hmmmm", "wkwk", "wkwkwk", "thanks", "thank you", "terima kasih"}
	for _, val := range listIgnore {
		if strings.EqualFold(val, str) || strings.Contains(val, str) {
			return true
		}
	}
	return false
}

func IsReset(str string) bool {
	listResetMsg := []string{"unified", "hi", "hello", "halo", "helo", "hai", "hey", "Hi Unified", "Hello unified", "Hey"}
	for _, val := range listResetMsg {
		if strings.EqualFold(val, str) {
			return true
		}
	}
	return false
}

func ProcessDate(dateStr string) time.Time {
	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		fmt.Println(err)
		return time.Time{}
	}

	now := time.Now()
	if date.Before(now) {
		date = date.AddDate(1, 0, 0)
	}

	return date
}

func ToString(data interface{}) string {
	return fmt.Sprintf("%v", data)
}

func ToInt64(data interface{}) int64 {
	switch value := data.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case string:
		res, _ := strconv.ParseInt(value, 10, 64)
		return res
	default:
		return 0
	}
}

func ToInt(data interface{}) (int, error) {
	int64Value := ToInt64(data)
	return int(int64Value), nil
}

func Ellipsis(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	if maxLen < 3 {
		maxLen = 3
	}

	return string(runes[0:maxLen-3]) + "..."
}

func ImageToBase64(image []byte) (string, error) {
	var base64Encoding string

	mimeType := http.DetectContentType(image)

	// Prepend the appropriate URI scheme header depending on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	case "image/gif":
		base64Encoding += "data:image/gif;base64,"
	case "image/bmp":
		base64Encoding += "data:image/bmp;base64,"
	case "image/webp":
		base64Encoding += "data:image/webp;base64,"
	case "image/tiff":
		base64Encoding += "data:image/tiff;base64,"
	case "image/svg+xml":
		base64Encoding += "data:image/svg+xml;base64,"
	default:
		return "", fmt.Errorf("invalid image type: %s", mimeType)
	}

	// Append the base64 encoded output
	base64Encoding += base64.StdEncoding.EncodeToString(image)
	return base64Encoding, nil
}

func GenerateOTP() int64 {
	rand.Seed(time.Now().UnixNano())

	// Generate a random 4-digit number
	number := rand.Intn(9000) + 1000
	return int64(number)
}

func FormatUnixTime(unix int64, format string) string {
	t := time.Unix(unix, 0)
	return t.Format(format)
}
