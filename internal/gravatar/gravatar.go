package gravatar

import (
	"crypto/md5"
	"strings"
)

func GetURL(email string) string {
	r := md5.Sum([]byte(strings.TrimSpace(strings.ToLower(email))))
	return string(r[:])
}
