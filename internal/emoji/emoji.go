// emoji code copied from https://github.com/0xadada/random-emoji
package emoji

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var (
	Emoticons   = newEmojiCat(0x1f600, 0x1f64f)
	Food        = newEmojiCat(0x1f32d, 0x1f37f)
	Animals     = newEmojiCat(0x1f400, 0x1f4d3)
	Expressions = newEmojiCat(0x1f910, 0x1f92f)
)

type Emojier interface {
	Random() string
}

type emojiCategory struct {
	begin int
	end   int
}

func newEmojiCat(begin, end int) Emojier {
	if begin > end {
		panic("begin is greater than end")
	}
	return emojiCategory{begin: begin, end: end}
}

func (cat emojiCategory) Random() string {
	rnd := rand.Intn(cat.end-cat.begin) + cat.begin

	return string(rune(rnd))
}

func Random(categories ...Emojier) string {
	if len(categories) == 0 {
		categories = []Emojier{Emoticons, Food, Animals, Expressions}
	}

	cat := categories[rand.Intn(len(categories))]

	return cat.Random()
}
