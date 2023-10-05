package common

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
	"github.com/temukan-co/monolith/core/entity"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func FormatIDR(amount *int64) string {
	if amount == nil {
		return ""
	} else if *amount == 0 {
		return "Gratis"
	}

	// Convert the amount to a string
	amountStr := strconv.FormatInt(*amount, 10)

	// Determine the length of the amount string
	length := len(amountStr)

	// Create a formatted string with IDR format
	formatted := "IDR "
	for i := 0; i < length; i++ {
		formatted += string(amountStr[i])
		if (length-i-1)%3 == 0 && i != length-1 {
			formatted += "."
		}
	}

	return formatted
}

func DoCommonHeader(c *gin.Context) entity.MandatoryRequest {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	name, _ := c.Get("name")
	version := c.GetHeader("version")

	return entity.MandatoryRequest{
		UserID:  ToInt64(userID),
		Email:   ToString(email),
		Name:    ToString(name),
		Version: ToInt64(version),
	}
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
