package util

import (
	"strings"
	"testing"
)

func TestExtractDSCacheUtilKeyValues(t *testing.T) {
	t.Run("typical kv lines", func(t *testing.T) {
		// verify "normal" text has extracted user info
		const normativeSample = `
name: ec2-user
password: ********
uid: 501
gid: 20
dir: /Users/ec2-user
shell: /bin/bash
gecos: ec2-user
`
		extracted := extractDSCacheUtilKeyValues([]byte(normativeSample), []string{"uid", "gid", "name"})
		if len(extracted) != 3 {
			t.Errorf("got back extracted keys: %+v", extracted)
		}

		for expectedKey, expectedValue := range map[string]string{
			"name": "ec2-user",
			"uid": "501",
			"gid": "20",
		} {
			v, ok := extracted[expectedKey]
			if !ok {
				t.Errorf("extracted kv should have key %q", expectedKey)
			}
			if !strings.EqualFold(v, expectedValue) {
				t.Errorf("extracted kv should have key %q with value %+v",
					expectedKey, expectedValue)
			}
		}
	})

	t.Run("mixed kv lines", func(t *testing.T) {
		// verify keys with mixed characteristics are
		// extracted with values intact
		const mixedSample = `
# busted line
-ignored-line-
foo: bar baz
qux: test
neato key: and key value
with sep: : foo
x: y

: bad
`
		extracted := extractDSCacheUtilKeyValues([]byte(mixedSample),
			[]string{"qux", "neato key", "with sep"})
		if len(extracted) != 3 {
			t.Errorf("got back extracted keys: %+v", extracted)
		}

		for expectedKey, expectedValue := range map[string]string{
			"qux": "test",
			"neato key": "and key value",
			"with sep": ": foo",
		} {
			v, ok := extracted[expectedKey]
			if !ok {
				t.Errorf("extracted kv should have key %q", expectedKey)
			}
			if !strings.EqualFold(v, expectedValue) {
				t.Errorf("extracted kv should have key %q with value %q",
					expectedKey, expectedValue)
			}

		}
	})

	t.Run("empty", func(t *testing.T) {
		// verify an "empty" output results in empty extraction
		const emptyNonZeroSample = `

`
		extracted := extractDSCacheUtilKeyValues([]byte(emptyNonZeroSample), []string{"qux"})
		if len(extracted) != 0 {
			t.Errorf("got back extracted keys: %+v", extracted)
		}

		const someKey = "qux"		
		v, ok := extracted[someKey]
		if ok {
			t.Errorf("extracted kv should NOT have key %q", someKey)
		}
		if v != "" {
			t.Errorf("extracted kv should NOT have value for key %q", someKey)
		}
	})
}
