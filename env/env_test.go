package env

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"
	"go.katupy.io/klib/mucache"
	"go.katupy.io/xpdt/conf"
)

func Test_container_loadEnviron(t *testing.T) {
	testCases := []*struct {
		name      string
		container *container
		environ   []string
		err       *klib.Error
		wantEnv   map[string]*environVar
	}{
		{
			name:      "empty-environ",
			container: &container{},
			wantEnv:   map[string]*environVar{},
		},
		{
			name:      "case-sensitive",
			container: &container{},
			environ: []string{
				"foo=bar1",
				"Foo=Bar2",
				"HAVE==StartingEqual",
			},
			wantEnv: map[string]*environVar{
				"foo": {
					key:           "foo",
					originalValue: "bar1",
					currentValue:  "bar1",
				},
				"Foo": {
					key:           "Foo",
					originalValue: "Bar2",
					currentValue:  "Bar2",
				},
				"HAVE": {
					key:           "HAVE",
					originalValue: "=StartingEqual",
					currentValue:  "=StartingEqual",
				},
			},
		},
		{
			name: "case-insensitive",
			container: &container{
				caseInsensitiveEnvironment: true,
			},
			environ: []string{
				"foo=bar1",
				"Foo=Bar2",
				"have=two=intermediary=equals",
			},
			wantEnv: map[string]*environVar{
				"FOO": {
					key:           "Foo",
					originalValue: "Bar2",
					currentValue:  "Bar2",
				},
				"HAVE": {
					key:           "have",
					originalValue: "two=intermediary=equals",
					currentValue:  "two=intermediary=equals",
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			err := tc.container.loadEnviron(tc.environ)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			haveEnv := tc.container.env
			wantEnv := tc.wantEnv

			if assert.Equal(st, len(wantEnv), len(haveEnv), "Env length mismatch") {
				for k := range wantEnv {
					have := haveEnv[k]
					want := wantEnv[k]

					assert.Equal(st, want, have, "Env[%q] value mismatch", k)
				}
			}
		})
	}
}

func Test_container_applyReverse(t *testing.T) {
	testCases := []*struct {
		name      string
		container *container
		err       *klib.Error
		wantEnv   map[string]*environVar
	}{
		{
			name: "serialization-error",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: "invalid-json",
					},
				},
			},
			err: &klib.Error{
				ID:     "57221f4e-bf90-4998-81b9-2544447fcf33",
				Status: http.StatusBadRequest,
				Code:   klib.CodeSerializationError,
			},
		},
		{
			name: "missing-key",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `["-"]`,
					},
				},
			},
			err: &klib.Error{
				ID:     "d8fd6be3-9555-464d-b6c7-2189545ba9b0",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
			},
		},
		{
			name: "unsupported-command",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `["-",""]`,
					},
				},
			},
			err: &klib.Error{
				ID:     "52b39360-02a8-4646-9b1e-2cfc75f7b025",
				Status: http.StatusBadRequest,
				Code:   klib.CodeInvalidValue,
			},
		},
		{
			name: "empty",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `[]`,
					},
				},
			},
			wantEnv: map[string]*environVar{
				conf.EnvReverseVar: {
					originalValue: `[]`,
					delete:        true,
				},
			},
		},
		{
			name: "set-new",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `["SET","foo","bar"]`,
					},
				},
			},
			wantEnv: map[string]*environVar{
				conf.EnvReverseVar: {
					originalValue: `["SET","foo","bar"]`,
					delete:        true,
				},
				"foo": {
					key:           "foo",
					originalValue: "bar",
					currentValue:  "bar",
					reversal:      true,
				},
			},
		},
		{
			name: "set-overwrite",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `["SET","foo","bar"]`,
					},
					"foo": {
						key:           "foo",
						originalValue: "bar-old",
						currentValue:  "bar-old",
					},
				},
			},
			wantEnv: map[string]*environVar{
				conf.EnvReverseVar: {
					originalValue: `["SET","foo","bar"]`,
					delete:        true,
				},
				"foo": {
					key:           "foo",
					originalValue: "bar",
					currentValue:  "bar",
					reversal:      true,
				},
			},
		},
		{
			name: "del",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `["DEL","foo"]`,
					},
					"foo": {
						key:           "foo",
						originalValue: "bar-old",
						currentValue:  "bar-old",
					},
				},
			},
			wantEnv: map[string]*environVar{
				conf.EnvReverseVar: {
					originalValue: `["DEL","foo"]`,
					delete:        true,
				},
				"foo": {
					key:            "foo",
					delete:         true,
					reversal:       true,
					reversalDelete: true,
				},
			},
		},
		{
			name: "multiple-ops",
			container: &container{
				caseInsensitiveEnvironment: true,
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						originalValue: `["DEL","FOO","SET","bar","foo"]`,
					},
					"FOO": {
						key:           "FOO",
						originalValue: "bar-old",
						currentValue:  "bar-old",
					},
				},
			},
			wantEnv: map[string]*environVar{
				conf.EnvReverseVar: {
					originalValue: `["DEL","FOO","SET","bar","foo"]`,
					delete:        true,
				},
				"BAR": {
					key:           "bar",
					originalValue: "foo",
					currentValue:  "foo",
					reversal:      true,
				},
				"FOO": {
					key:            "FOO",
					delete:         true,
					reversal:       true,
					reversalDelete: true,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			err := tc.container.applyReverse()
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			haveEnv := tc.container.env
			wantEnv := tc.wantEnv

			if assert.Equal(st, len(wantEnv), len(haveEnv), "Env length mismatch") {
				for k := range wantEnv {
					have := haveEnv[k]
					want := wantEnv[k]

					assert.Equal(st, want, have, "Env[%q] value mismatch", k)
				}
			}
		})
	}
}

func Test_container_makeDiff(t *testing.T) {
	cache := mucache.New[string, string]()

	testCases := []*struct {
		name        string
		container   *container
		wantDiff    []string
		wantReverse []string
	}{
		{
			name:        "empty",
			container:   &container{},
			wantDiff:    []string{},
			wantReverse: []string{},
		},
		{
			name: "add-entry",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "",
						currentValue:  "bar",
					},
				},
			},
			wantDiff: []string{
				"SET", "foo", "bar",
			},
			wantReverse: []string{
				"DEL", "foo",
			},
		},
		{
			name: "add-empty-entry",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "",
						currentValue:  "",
						created:       true,
					},
				},
			},
			wantDiff: []string{
				"SET", "foo", "",
			},
			wantReverse: []string{
				"DEL", "foo",
			},
		},
		{
			name: "set-entry-path-list",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "bar1",
						currentValue: cache.SetGet(
							"set-entry-path-list",
							strings.Join([]string{"bar1", "bar2", "bar3"}, string(os.PathListSeparator)),
						),
					},
				},
			},
			wantDiff: []string{
				"SET", "foo", cache.Get("set-entry-path-list"),
			},
			wantReverse: []string{
				"SET", "foo", "bar1",
			},
		},
		{
			name: "set-entry-has-reversal-delete",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:            "foo",
						originalValue:  "bar-old",
						currentValue:   "bar",
						reversal:       true,
						reversalDelete: true,
					},
				},
			},
			wantDiff: []string{
				"SET", "foo", "bar",
			},
			wantReverse: []string{
				"DEL", "foo",
			},
		},
		{
			name: "set-entry-has-reversal",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "bar-old",
						currentValue:  "bar",
						reversal:      true,
					},
				},
			},
			wantDiff: []string{
				"SET", "foo", "bar",
			},
			wantReverse: []string{
				"SET", "foo", "bar-old",
			},
		},
		{
			name: "del-entry",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "bar",
						currentValue:  "",
						delete:        true,
					},
				},
			},
			wantDiff: []string{
				"DEL", "foo",
			},
			wantReverse: []string{
				"SET", "foo", "bar",
			},
		},
		{
			name: "del-entry-with-reversal-delete",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:            "foo",
						originalValue:  "",
						currentValue:   "",
						delete:         true,
						reversal:       true,
						reversalDelete: true,
					},
				},
			},
			wantDiff: []string{
				"DEL", "foo",
			},
			wantReverse: []string{},
		},
		{
			name: "del-reverse-env",
			container: &container{
				env: map[string]*environVar{
					conf.EnvReverseVar: {
						key:           conf.EnvReverseVar,
						originalValue: "bar",
						currentValue:  "",
						delete:        true,
					},
				},
			},
			wantDiff: []string{
				"DEL", conf.EnvReverseVar,
			},
			wantReverse: []string{},
		},
		{
			name: "skip-unchanged-entry",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "bar",
						currentValue:  "bar",
					},
				},
			},
			wantDiff:    []string{},
			wantReverse: []string{},
		},
		{
			name: "apply-reversal-on-unchanged-reversal-entry",
			container: &container{
				env: map[string]*environVar{
					"foo": {
						key:           "foo",
						originalValue: "bar",
						currentValue:  "bar",
						reversal:      true,
					},
				},
			},
			wantDiff: []string{
				"SET", "foo", "bar",
			},
			wantReverse: []string{},
		},
		{
			name: "multiple-entries",
			container: &container{
				env: map[string]*environVar{
					"a": {
						key:           "a",
						originalValue: "1",
						currentValue:  "3",
					},
					"b": {
						key:           "b",
						originalValue: "2",
						currentValue:  "2",
					},
					"c": {
						key:           "c",
						originalValue: "3",
						currentValue:  "1",
					},
					"d": {
						key:           "d",
						originalValue: "4",
						currentValue:  "",
						delete:        true,
					},
					"e": {
						key:           "e",
						originalValue: "",
						currentValue:  "5",
					},
					"f": {
						key:           "f",
						originalValue: "",
						currentValue:  "",
						created:       true,
					},
					"g": {
						key:           "g",
						originalValue: "6",
						currentValue:  "",
						delete:        true,
					},
				},
			},
			wantDiff: []string{
				"SET", "a", "3",
				"SET", "c", "1",
				"DEL", "d",
				"SET", "e", "5",
				"SET", "f", "",
				"DEL", "g",
			},
			wantReverse: []string{
				"SET", "a", "1",
				"SET", "c", "3",
				"SET", "d", "4",
				"DEL", "e",
				"DEL", "f",
				"SET", "g", "6",
			},
		},
	}

	sortCmdsByKeys := func(slice []string) []string {
		keys := []string{}
		cmds := make(map[string][]string)

		for i := 0; i < len(slice); i++ {
			cmd := slice[i]
			key := slice[i+1]
			keys = append(keys, key)

			switch cmd {
			case "SET":
				cmds[key] = []string{"SET", key, slice[i+2]}
				i += 2
			case "DEL":
				cmds[key] = []string{"DEL", key}
				i += 1
			}
		}

		sort.Strings(keys)
		out := []string{}

		for i := range keys {
			out = append(out, cmds[keys[i]]...)
		}

		return out
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			tc.container.makeDiff()

			haveDiff := sortCmdsByKeys(tc.container.diff)
			wantDiff := tc.wantDiff

			if assert.Equal(st, len(wantDiff), len(haveDiff), "Diff length mismatch") {
				for j := range wantDiff {
					have := haveDiff[j]
					want := wantDiff[j]

					assert.Equal(st, want, have, "Diff[%d] value mismatch", j)
				}
			}

			haveReverse := sortCmdsByKeys(tc.container.reverse)
			wantReverse := tc.wantReverse

			if assert.Equal(st, len(wantReverse), len(haveReverse), "Reverse length mismatch") {
				for j := range wantReverse {
					have := haveReverse[j]
					want := wantReverse[j]

					assert.Equal(st, want, have, "Reverse[%d] value mismatch", j)
				}
			}
		})
	}
}
