package gravatar

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func GetUrlOrDefault(email *string) string {
	if email == nil {
		return "https://www.gravatar.com/avatar/?d=wavatar"
	}

	r := md5.Sum([]byte(strings.TrimSpace(strings.ToLower(*email))))
	imageURL := fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=wavatar", r)

	return imageURL
}
