package env

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"
	"go.katupy.io/klib/mocks/mock_klib"
	"go.katupy.io/klib/mucache"
)

func Test_defaultCommandLoader_Load(t *testing.T) {
	testCases := []*struct {
		name                 string
		cmd                  *Command
		commandLoader        *defaultCommandLoader
		mockCommandMethodsOn [][]any
		err                  *klib.Error
	}{
		{
			name: "skip-platform",
			cmd: &Command{
				Add:      "yes",
				Platform: "not-foo",
			},
			commandLoader: &defaultCommandLoader{
				platform: "foo",
			},
		},
		{
			name: "method-add",
			cmd: &Command{
				Add: "yes",
			},
			commandLoader: &defaultCommandLoader{},
			mockCommandMethodsOn: [][]any{
				{"Add", &Command{Add: "yes"}, nil},
			},
		},
		{
			name: "method-set",
			cmd: &Command{
				Set: "yes",
			},
			commandLoader: &defaultCommandLoader{},
			mockCommandMethodsOn: [][]any{
				{"Set", &Command{Set: "yes"}, nil},
			},
		},
		{
			name: "method-del",
			cmd: &Command{
				Del: "yes",
			},
			commandLoader: &defaultCommandLoader{},
			mockCommandMethodsOn: [][]any{
				{"Del", &Command{Del: "yes"}, nil},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			mockCommandMethods := NewMockCommandMethods(st)

			for j := range tc.mockCommandMethodsOn {
				on := tc.mockCommandMethodsOn[j]
				mockCommandMethods.On(on[0].(string), on[1]).Return(on[2])
			}

			tc.commandLoader.commandMethods = mockCommandMethods

			err := tc.commandLoader.Load(tc.cmd)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}
		})
	}
}

func Test_defaultCommandMethods_Add(t *testing.T) {
	cache := mucache.New[string, *environVar]()

	testCases := []*struct {
		name                  string
		cmd                   *Command
		compareKey            string
		commandMethods        *defaultCommandMethods
		wantEnv               map[string]*environVar
		mockTemplateHandlerOn []any
		mockPathLoaderOn      []any
		mockPathHandlerOn     [][]any
		err                   *klib.Error
	}{
		{
			name:           "missing-value",
			cmd:            &Command{},
			commandMethods: &defaultCommandMethods{},
			err: &klib.Error{
				ID:     "ce66adb8-2e56-40b7-8268-27a8955296b5",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
			},
		},
		{
			name: "append-and-undelete",
			cmd: &Command{
				Add:    "foo",
				Value:  "=",
				Append: true,
			},
			compareKey: "foo",
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{
						"foo": {
							currentValue: "?",
							delete:       true,
						},
					},
				},
			},
			wantEnv: map[string]*environVar{
				"foo": cache.SetGet(
					"append-and-undelete",
					&environVar{
						currentValue: "?",
						delete:       false,
					},
				),
			},
			mockTemplateHandlerOn: []any{"Handle", "=", "OK", nil},
			mockPathLoaderOn:      []any{"Load", cache.Get("append-and-undelete"), nil},
			mockPathHandlerOn: [][]any{
				{"Add", cache.Get("append-and-undelete"), "OK", -1, nil},
			},
		},
		{
			name: "prepend",
			cmd: &Command{
				Add:   "bar",
				Value: "@",
			},
			compareKey: "bar",
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{},
				},
			},
			wantEnv: map[string]*environVar{
				"bar": cache.SetGet(
					"prepend",
					&environVar{
						key:     "bar",
						created: true,
					},
				),
			},
			mockTemplateHandlerOn: []any{"Handle", "@", "Done", nil},
			mockPathLoaderOn:      []any{"Load", cache.Get("prepend"), nil},
			mockPathHandlerOn: [][]any{
				{"Add", cache.Get("prepend"), "Done", 0, nil},
			},
		},
		{
			name: "prepend-case-insensitive",
			cmd: &Command{
				Add:   "bar",
				Value: "@",
			},
			compareKey: "BAR",
			commandMethods: &defaultCommandMethods{
				container: &container{
					caseInsensitiveEnvironment: true,
					env:                        map[string]*environVar{},
				},
			},
			wantEnv: map[string]*environVar{
				"BAR": cache.SetGet(
					"prepend-case-insensitive",
					&environVar{
						key:     "bar",
						created: true,
					},
				),
			},
			mockTemplateHandlerOn: []any{"Handle", "@", "Done", nil},
			mockPathLoaderOn:      []any{"Load", cache.Get("prepend-case-insensitive"), nil},
			mockPathHandlerOn: [][]any{
				{"Add", cache.Get("prepend-case-insensitive"), "Done", 0, nil},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			mockTemplateHandler := mock_klib.NewMockStringHandler(st)
			mockPathLoader := NewMockPathLoader(st)
			mockPathHandler := NewMockPathHandler(st)

			if len(tc.mockTemplateHandlerOn) > 0 {
				on := tc.mockTemplateHandlerOn
				mockTemplateHandler.On(on[0].(string), on[1]).Return(on[2], on[3])
			}

			if len(tc.mockPathLoaderOn) > 0 {
				on := tc.mockPathLoaderOn
				mockPathLoader.On(on[0].(string), on[1]).Return(on[2])
			}

			for j := range tc.mockPathHandlerOn {
				on := tc.mockPathHandlerOn[j]
				mockPathHandler.On(on[0].(string), on[1], on[2], on[3]).Return(on[4])
			}

			tc.commandMethods.templateHandler = mockTemplateHandler
			tc.commandMethods.pathLoader = mockPathLoader
			tc.commandMethods.pathHandler = mockPathHandler

			err := tc.commandMethods.Add(tc.cmd)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			haveEnv := tc.commandMethods.container.env
			wantEnv := tc.wantEnv

			if assert.Equal(st, len(wantEnv), len(haveEnv), "Env length mismatch") {
				for k := range wantEnv {
					have := haveEnv[k]
					want := wantEnv[k]

					assert.Equal(st, want, have, "Env[%q] mismatch", k)
				}
			}
		})
	}
}

func Test_defaultCommandMethods_Set(t *testing.T) {
	testCases := []*struct {
		name                  string
		cmd                   *Command
		commandMethods        *defaultCommandMethods
		mockTemplateHandlerOn []any
		err                   *klib.Error
		wantEnv               map[string]*environVar
	}{
		{
			name: "create",
			cmd: &Command{
				Set:   "foo",
				Value: "bar",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "bar", "barOK", nil},
			wantEnv: map[string]*environVar{
				"foo": {
					key:          "foo",
					currentValue: "barOK",
					created:      true,
				},
			},
		},
		{
			name: "overwrite",
			cmd: &Command{
				Set:   "foo",
				Value: "bar",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{
						"foo": {
							currentValue: "-",
						},
					},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "bar", "barOK", nil},
			wantEnv: map[string]*environVar{
				"foo": {
					currentValue: "barOK",
				},
			},
		},
		{
			name: "overwrite-case-insensitive-and-undelete",
			cmd: &Command{
				Set:   "foo",
				Value: "bar",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					caseInsensitiveEnvironment: true,
					env: map[string]*environVar{
						"FOO": {
							currentValue: "-",
							delete:       true,
						},
					},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "bar", "barOK", nil},
			wantEnv: map[string]*environVar{
				"FOO": {
					currentValue: "barOK",
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			mockTemplateHandler := mock_klib.NewMockStringHandler(st)

			if len(tc.mockTemplateHandlerOn) > 0 {
				on := tc.mockTemplateHandlerOn
				mockTemplateHandler.On(on[0].(string), on[1]).Return(on[2], on[3])
			}

			tc.commandMethods.templateHandler = mockTemplateHandler

			err := tc.commandMethods.Set(tc.cmd)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			haveEnv := tc.commandMethods.container.env
			wantEnv := tc.wantEnv

			if assert.Equal(st, len(wantEnv), len(haveEnv), "Env length mismatch") {
				for k := range wantEnv {
					have := haveEnv[k]
					want := wantEnv[k]

					assert.Equal(st, want, have, "Env[%q] mismatch", k)
				}
			}
		})
	}
}

func Test_defaultCommandMethods_Del(t *testing.T) {
	keyA := strings.Join([]string{"1,2,3"}, string(os.PathListSeparator))
	keyB := strings.Join([]string{"40,50,60"}, string(os.PathListSeparator))
	keyC := strings.Join([]string{"700,800,900"}, string(os.PathListSeparator))

	testCases := []*struct {
		name           string
		cmd            *Command
		commandMethods *defaultCommandMethods
		err            *klib.Error
		wantEnv        map[string]*environVar
	}{
		{
			name: "non-existing-key",
			cmd: &Command{
				Del:   "foo",
				Value: "bar",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{
						"a": {
							originalValue:    "keyA",
							currentValue:     keyA,
							pathList:         true,
							pathListElements: []string{"1", "2", "3"},
							pathListElementExists: map[string]bool{
								"1": true,
								"2": true,
								"3": true,
							},
						},
						"b": {
							originalValue:    "keyB",
							currentValue:     keyB,
							pathList:         true,
							pathListElements: []string{"40", "50", "60"},
							pathListElementExists: map[string]bool{
								"40": true,
								"50": true,
								"60": true,
							},
						},
						"c": {
							originalValue:    "keyC",
							currentValue:     keyC,
							pathList:         true,
							pathListElements: []string{"700", "800", "900"},
							pathListElementExists: map[string]bool{
								"700": true,
								"800": true,
								"900": true,
							},
						},
						"d": {
							originalValue: "keyD",
							currentValue:  "4",
						},
					},
				},
			},
			wantEnv: map[string]*environVar{
				"a": {
					originalValue:    "keyA",
					currentValue:     keyA,
					pathList:         true,
					pathListElements: []string{"1", "2", "3"},
					pathListElementExists: map[string]bool{
						"1": true,
						"2": true,
						"3": true,
					},
				},
				"b": {
					originalValue:    "keyB",
					currentValue:     keyB,
					pathList:         true,
					pathListElements: []string{"40", "50", "60"},
					pathListElementExists: map[string]bool{
						"40": true,
						"50": true,
						"60": true,
					},
				},
				"c": {
					originalValue:    "keyC",
					currentValue:     keyC,
					pathList:         true,
					pathListElements: []string{"700", "800", "900"},
					pathListElementExists: map[string]bool{
						"700": true,
						"800": true,
						"900": true,
					},
				},
				"d": {
					originalValue: "keyD",
					currentValue:  "4",
				},
			},
		},
		{
			name: "existing-key",
			cmd: &Command{
				Del: "c",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{
						"a": {
							originalValue:    "keyA",
							currentValue:     keyA,
							pathList:         true,
							pathListElements: []string{"1", "2", "3"},
							pathListElementExists: map[string]bool{
								"1": true,
								"2": true,
								"3": true,
							},
						},
						"b": {
							originalValue:    "keyB",
							currentValue:     keyB,
							pathList:         true,
							pathListElements: []string{"40", "50", "60"},
							pathListElementExists: map[string]bool{
								"40": true,
								"50": true,
								"60": true,
							},
						},
						"c": {
							originalValue:    "keyC",
							currentValue:     keyC,
							pathList:         true,
							pathListElements: []string{"700", "800", "900"},
							pathListElementExists: map[string]bool{
								"700": true,
								"800": true,
								"900": true,
							},
						},
						"d": {
							originalValue: "keyD",
							currentValue:  "4",
						},
					},
				},
			},
			wantEnv: map[string]*environVar{
				"a": {
					originalValue:    "keyA",
					currentValue:     keyA,
					pathList:         true,
					pathListElements: []string{"1", "2", "3"},
					pathListElementExists: map[string]bool{
						"1": true,
						"2": true,
						"3": true,
					},
				},
				"b": {
					originalValue:    "keyB",
					currentValue:     keyB,
					pathList:         true,
					pathListElements: []string{"40", "50", "60"},
					pathListElementExists: map[string]bool{
						"40": true,
						"50": true,
						"60": true,
					},
				},
				"c": {
					originalValue: "keyC",
					currentValue:  "",
					delete:        true,
				},
				"d": {
					originalValue: "keyD",
					currentValue:  "4",
				},
			},
		},
		{
			name: "asterisk",
			cmd: &Command{
				Del: "*",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					env: map[string]*environVar{
						"a": {
							originalValue:    "keyA",
							currentValue:     keyA,
							pathList:         true,
							pathListElements: []string{"1", "2", "3"},
							pathListElementExists: map[string]bool{
								"1": true,
								"2": true,
								"3": true,
							},
						},
						"b": {
							originalValue:    "keyB",
							currentValue:     keyB,
							pathList:         true,
							pathListElements: []string{"40", "50", "60"},
							pathListElementExists: map[string]bool{
								"40": true,
								"50": true,
								"60": true,
							},
							reversal: true,
						},
						"c": {
							originalValue:    "keyC",
							currentValue:     keyC,
							pathList:         true,
							pathListElements: []string{"700", "800", "900"},
							pathListElementExists: map[string]bool{
								"700": true,
								"800": true,
								"900": true,
							},
						},
						"d": {
							originalValue: "keyD",
							currentValue:  "4",
						},
					},
				},
			},
			wantEnv: map[string]*environVar{
				"a": {
					originalValue: "keyA",
					currentValue:  "",
					delete:        true,
				},
				"b": {
					originalValue: "keyB",
					currentValue:  "",
					delete:        true,
					reversal:      true,
				},
				"c": {
					originalValue: "keyC",
					currentValue:  "",
					delete:        true,
				},
				"d": {
					originalValue: "keyD",
					currentValue:  "",
					delete:        true,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			err := tc.commandMethods.Del(tc.cmd)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			haveEnv := tc.commandMethods.container.env
			wantEnv := tc.wantEnv

			if assert.Equal(st, len(wantEnv), len(haveEnv), "Env length mismatch") {
				for key := range wantEnv {
					have := haveEnv[key]
					want := wantEnv[key]

					assert.Equal(st, want, have, "Env[%q] value mismatch", key)
				}
			}
		})
	}
}
