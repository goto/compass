package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func AssertEqualProto(t *testing.T, expected, actual proto.Message) {
	t.Helper()

	if diff := cmp.Diff(actual, expected, protocmp.Transform()); diff != "" {
		msg := fmt.Sprintf(
			"Not equal:\n"+
				"expected:\n\t'%s'\n"+
				"actual:\n\t'%s'\n"+
				"diff (-expected +actual):\n%s",
			expected, actual, diff,
		)
		assert.Fail(t, msg)
	}
}

func Marshal(t *testing.T, v interface{}) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}

type ArgMatcher interface{ Matches(interface{}) bool }

func OfTypeContext() ArgMatcher {
	return mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil })
}

func AreSlicesEqualIgnoringOrder[T any](s1, s2 []T, compare func(l, r T) int) bool {
	if len(s1) != len(s2) {
		return false
	}

	s1Duplicate := make([]T, len(s1))
	copy(s1Duplicate, s1)
	s2Duplicate := make([]T, len(s2))
	copy(s2Duplicate, s2)

	sort.Slice(s1Duplicate, func(i, j int) bool {
		return compare(s1Duplicate[i], s1Duplicate[j]) < 0
	})

	sort.Slice(s2Duplicate, func(i, j int) bool {
		return compare(s2Duplicate[i], s2Duplicate[j]) < 0
	})

	for i := 0; i < len(s1Duplicate); i++ {
		if compare(s1Duplicate[i], s2Duplicate[i]) != 0 {
			return false
		}
	}

	return true
}
