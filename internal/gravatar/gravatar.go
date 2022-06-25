package gravatar

import (
	"crypto/md5"
	"strings"
)

func GetUrlOrDefault(email *string) string {
	if email == nil {
		return "https://www.gravatar.com/avatar/?d=wavatar"
	}

	r := md5.Sum([]byte(strings.TrimSpace(strings.ToLower(*email))))
	imageURL := "https://www.gravatar.com/avatar/" + string(r[:]) + "?d=wavatar"

	return imageURL
}
