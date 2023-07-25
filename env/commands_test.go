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
	testCases := []*struct {
		name                  string
		cmd                   *Command
		compareKey            string
		commandMethods        *defaultCommandMethods
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
					delEnv: map[string]bool{
						"foo": true,
					},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "=", "OK", nil},
			mockPathLoaderOn:      []any{"Load", "foo", nil},
			mockPathHandlerOn: [][]any{
				{"Add", "foo", "OK", -1, nil},
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
				container: &container{},
			},
			mockTemplateHandlerOn: []any{"Handle", "@", "Done", nil},
			mockPathLoaderOn:      []any{"Load", "bar", nil},
			mockPathHandlerOn: [][]any{
				{"Add", "bar", "Done", 0, nil},
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
					delEnv: map[string]bool{
						"BAR": true,
					},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "@", "Done", nil},
			mockPathLoaderOn:      []any{"Load", "BAR", nil},
			mockPathHandlerOn: [][]any{
				{"Add", "BAR", "Done", 0, nil},
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

			assert.False(st, tc.commandMethods.container.delEnv[tc.compareKey], "Key should not be deleted")
		})
	}
}

func Test_defaultCommandMethods_Set(t *testing.T) {
	testCases := []*struct {
		name                  string
		cmd                   *Command
		compareKey            string
		commandMethods        *defaultCommandMethods
		mockTemplateHandlerOn []any
		err                   *klib.Error
	}{
		{
			name: "create-and-undelete",
			cmd: &Command{
				Set:   "foo",
				Value: "bar",
			},
			compareKey: "foo",
			commandMethods: &defaultCommandMethods{
				container: &container{
					curEnv: map[string]string{},
					delEnv: map[string]bool{
						"foo": true,
					},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "bar", "barOK", nil},
		},
		{
			name: "overwrite",
			cmd: &Command{
				Set:   "foo",
				Value: "bar",
			},
			compareKey: "foo",
			commandMethods: &defaultCommandMethods{
				container: &container{
					curEnv: map[string]string{
						"foo": "-",
					},
					delEnv: map[string]bool{},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "bar", "barOK", nil},
		},
		{
			name: "overwrite-case-insensitive",
			cmd: &Command{
				Set:   "foo",
				Value: "bar",
			},
			compareKey: "FOO",
			commandMethods: &defaultCommandMethods{
				container: &container{
					caseInsensitiveEnvironment: true,
					curEnv: map[string]string{
						"FOO": "-",
					},
					delEnv: map[string]bool{
						"FOO": true,
					},
				},
			},
			mockTemplateHandlerOn: []any{"Handle", "bar", "barOK", nil},
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

			want := tc.cmd.Value + "OK"
			have := tc.commandMethods.container.curEnv[tc.compareKey]

			assert.Equal(st, want, have, "Value mismatch")
			assert.False(st, tc.commandMethods.container.delEnv[tc.compareKey], "Key should not be deleted")
		})
	}
}

func Test_defaultCommandMethods_Del(t *testing.T) {
	keyA := strings.Join([]string{"1,2,3"}, string(os.PathListSeparator))
	keyB := strings.Join([]string{"40,50,60"}, string(os.PathListSeparator))
	keyC := strings.Join([]string{"700,800,900"}, string(os.PathListSeparator))

	testCases := []*struct {
		name                      string
		cmd                       *Command
		commandMethods            *defaultCommandMethods
		err                       *klib.Error
		wantCurEnv                map[string]string
		wantDelEnv                map[string]bool
		wantPathListExists        map[string]bool
		wantPathListElements      map[string][]string
		wantPathListElementExists map[string]map[string]bool
	}{
		{
			name: "non-existing-key",
			cmd: &Command{
				Del:   "foo",
				Value: "bar",
			},
			commandMethods: &defaultCommandMethods{
				container: &container{
					curEnv: map[string]string{
						"a": keyA,
						"b": keyB,
						"c": keyC,
						"d": "4",
					},
					delEnv: map[string]bool{},
					pathListExists: map[string]bool{
						"a": true,
						"b": true,
						"c": true,
					},
					pathListElements: map[string][]string{
						"a": {"1", "2", "3"},
						"b": {"40", "50", "60"},
						"c": {"700", "800", "900"},
					},
					pathListElementExists: map[string]map[string]bool{
						"a": {
							"1": true,
							"2": true,
							"3": true,
						},
						"b": {
							"40": true,
							"50": true,
							"60": true,
						},
						"c": {
							"700": true,
							"800": true,
							"900": true,
						},
					},
				},
			},
			wantCurEnv: map[string]string{
				"a": keyA,
				"b": keyB,
				"c": keyC,
				"d": "4",
			},
			wantDelEnv: map[string]bool{},
			wantPathListExists: map[string]bool{
				"a": true,
				"b": true,
				"c": true,
			},
			wantPathListElements: map[string][]string{
				"a": {"1", "2", "3"},
				"b": {"40", "50", "60"},
				"c": {"700", "800", "900"},
			},
			wantPathListElementExists: map[string]map[string]bool{
				"a": {
					"1": true,
					"2": true,
					"3": true,
				},
				"b": {
					"40": true,
					"50": true,
					"60": true,
				},
				"c": {
					"700": true,
					"800": true,
					"900": true,
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
					curEnv: map[string]string{
						"a": keyA,
						"b": keyB,
						"c": keyC,
						"d": "4",
					},
					delEnv: map[string]bool{},
					pathListExists: map[string]bool{
						"a": true,
						"b": true,
						"c": true,
					},
					pathListElements: map[string][]string{
						"a": {"1", "2", "3"},
						"b": {"40", "50", "60"},
						"c": {"700", "800", "900"},
					},
					pathListElementExists: map[string]map[string]bool{
						"a": {
							"1": true,
							"2": true,
							"3": true,
						},
						"b": {
							"40": true,
							"50": true,
							"60": true,
						},
						"c": {
							"700": true,
							"800": true,
							"900": true,
						},
					},
				},
			},
			wantCurEnv: map[string]string{
				"a": keyA,
				"b": keyB,
				"d": "4",
			},
			wantDelEnv: map[string]bool{
				"c": true,
			},
			wantPathListExists: map[string]bool{
				"a": true,
				"b": true,
			},
			wantPathListElements: map[string][]string{
				"a": {"1", "2", "3"},
				"b": {"40", "50", "60"},
			},
			wantPathListElementExists: map[string]map[string]bool{
				"a": {
					"1": true,
					"2": true,
					"3": true,
				},
				"b": {
					"40": true,
					"50": true,
					"60": true,
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
					curEnv: map[string]string{
						"a": keyA,
						"b": keyB,
						"c": keyC,
						"d": "4",
					},
					delEnv: map[string]bool{},
					pathListExists: map[string]bool{
						"a": true,
						"b": true,
						"c": true,
					},
					pathListElements: map[string][]string{
						"a": {"1", "2", "3"},
						"b": {"40", "50", "60"},
						"c": {"700", "800", "900"},
					},
					pathListElementExists: map[string]map[string]bool{
						"a": {
							"1": true,
							"2": true,
							"3": true,
						},
						"b": {
							"40": true,
							"50": true,
							"60": true,
						},
						"c": {
							"700": true,
							"800": true,
							"900": true,
						},
					},
				},
			},
			wantCurEnv: map[string]string{},
			wantDelEnv: map[string]bool{
				"a": true,
				"b": true,
				"c": true,
				"d": true,
			},
			wantPathListExists:        map[string]bool{},
			wantPathListElements:      map[string][]string{},
			wantPathListElementExists: map[string]map[string]bool{},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			err := tc.commandMethods.Del(tc.cmd)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			wantCurEnv := tc.wantCurEnv
			haveCurEnv := tc.commandMethods.container.curEnv

			if assert.Equal(st, len(wantCurEnv), len(haveCurEnv), "CurEnv length mismatch") {
				for key := range wantCurEnv {
					want := wantCurEnv[key]
					have := haveCurEnv[key]

					assert.Equal(st, want, have, "CurEnv[%q] value mismatch", key)
				}
			}

			wantDelEnv := tc.wantDelEnv
			haveDelEnv := tc.commandMethods.container.delEnv

			if assert.Equal(st, len(wantDelEnv), len(haveDelEnv), "DelEnv length mismatch") {
				for key := range wantDelEnv {
					want := wantDelEnv[key]
					have := haveDelEnv[key]

					assert.Equal(st, want, have, "DelEnv[%q] value mismatch", key)
				}
			}

			wantPathListExists := tc.wantPathListExists
			havePathListExists := tc.commandMethods.container.pathListExists

			if assert.Equal(st, len(wantPathListExists), len(havePathListExists), "PathListExists length mismatch") {
				for key := range wantPathListExists {
					want := wantPathListExists[key]
					have := havePathListExists[key]

					assert.Equal(st, want, have, "PathListExists[%q] value mismatch", key)
				}
			}

			wantPathListElements := tc.wantPathListElements
			havePathListElements := tc.commandMethods.container.pathListElements

			if assert.Equal(st, len(wantPathListElements), len(havePathListElements), "PathListElements length mismatch") {
				for key := range wantPathListElements {
					want := wantPathListElements[key]
					have := havePathListElements[key]

					assert.Equal(st, want, have, "PathListElements[%q] value mismatch", key)
				}
			}

			wantPathListElementExists := tc.wantPathListElementExists
			havePathListElementExists := tc.commandMethods.container.pathListElementExists

			if assert.Equal(st, len(wantPathListElementExists), len(havePathListElementExists), "PathListElementExists length mismatch") {
				for key := range wantPathListElementExists {
					want := wantPathListElementExists[key]
					have := havePathListElementExists[key]

					assert.Equal(st, want, have, "PathListElementExists[%q] value mismatch", key)
				}
			}
		})
	}
}
