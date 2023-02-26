//go:build test

package matcher

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

// EqualCmp is a more powerful and safer alternative to gomega.Equal for comparing whether two
// values are semantically equal.
func EqualCmp(expected interface{}, options ...cmp.Option) types.GomegaMatcher {
	if _, ok := expected.(proto.Message); ok {
		options = append(options, protocmp.Transform())
	}

	return &equalCmpMatcher{
		expected: expected,
		options:  options,
	}
}

type equalCmpMatcher struct {
	expected interface{}
	options  cmp.Options
}

func (matcher *equalCmpMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && matcher.expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  This is to avoid mistakes where both sides of an assertion are erroneously uninitialized")
	}
	return cmp.Equal(actual, matcher.expected, matcher.options), nil
}

func (matcher *equalCmpMatcher) FailureMessage(actual interface{}) (message string) {
	actualString, actualOK := actual.(string)
	expectedString, expectedOK := matcher.expected.(string)
	if actualOK && expectedOK {
		return format.MessageWithDiff(actualString, "to equal", expectedString)
	}

	diff := cmp.Diff(actual, matcher.expected, matcher.options)

	expected := matcher.expected
	if actualProto, ok := actual.(proto.Message); ok {
		actual = protojson.Format(actualProto)
	}
	if expectedProto, ok := expected.(proto.Message); ok {
		expected = protojson.Format(expectedProto)
	}

	return format.Message(actual, "to equal", expected) +
		"\n\nDiff:\n" + format.IndentString(diff, 1)
}

func (matcher *equalCmpMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	diff := cmp.Diff(actual, matcher.expected, matcher.options)
	return format.Message(actual, "not to equal", matcher.expected) +
		"\n\nDiff:\n" + format.IndentString(diff, 1)
}
