package env

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"
	"go.katupy.io/klib/must"
)

func Test_defaultPathHandler_Add(t *testing.T) {
	testCases := []*struct {
		name         string
		pathHandler  *defaultPathHandler
		key          string
		value        string
		position     int
		err          *klib.Error
		wantPathList []string
	}{
		{
			name:  "non-abs-path-list-value",
			key:   "foo",
			value: "bar",
			pathHandler: &defaultPathHandler{
				container: &container{
					pathListElementExists: map[string]map[string]bool{},
				},
			},
			err: &klib.Error{
				ID:     "0d3fb866-c0be-420d-89e7-11b0f05ff132",
				Status: http.StatusBadRequest,
				Code:   klib.CodeInvalidValue,
			},
		},
		{
			name:  "new-path-list",
			key:   "foo",
			value: must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{
				container: &container{
					pathListElements:      map[string][]string{},
					pathListElementExists: map[string]map[string]bool{},
				},
			},
			wantPathList: []string{
				must.FilepathAbs("bar"),
			},
		},
		{
			name:  "prepend",
			key:   "foo",
			value: must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{
				container: &container{
					pathListElements: map[string][]string{
						"foo": {
							must.FilepathAbs("a"),
							must.FilepathAbs("b"),
						},
					},
					pathListElementExists: map[string]map[string]bool{},
				},
			},
			wantPathList: []string{
				must.FilepathAbs("bar"),
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
			},
		},
		{
			name:     "append",
			key:      "foo",
			value:    must.FilepathAbs("bar"),
			position: -1,
			pathHandler: &defaultPathHandler{
				container: &container{
					pathListElements: map[string][]string{
						"foo": {
							must.FilepathAbs("a"),
							must.FilepathAbs("b"),
						},
					},
					pathListElementExists: map[string]map[string]bool{},
				},
			},
			wantPathList: []string{
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
		{
			name:  "ignore-existing-path-case-sensitive",
			key:   "foo",
			value: must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{
				caseSensitiveFilesystem: true,
				container: &container{
					pathListElements: map[string][]string{
						"foo": {
							must.FilepathAbs("a"),
							must.FilepathAbs("b"),
							must.FilepathAbs("bar"),
						},
					},
					pathListElementExists: map[string]map[string]bool{
						"foo": {
							must.FilepathAbs("a"):   true,
							must.FilepathAbs("b"):   true,
							must.FilepathAbs("bar"): true,
						},
					},
				},
			},
			wantPathList: []string{
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
		{
			name:  "ignore-existing-path-case-insensitive",
			key:   "foo",
			value: must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{
				container: &container{
					pathListElements: map[string][]string{
						"foo": {
							must.FilepathAbs("a"),
							must.FilepathAbs("b"),
							must.FilepathAbs("bar"),
						},
					},
					pathListElementExists: map[string]map[string]bool{
						"foo": {
							strings.ToUpper(must.FilepathAbs("a")):   true,
							strings.ToUpper(must.FilepathAbs("b")):   true,
							strings.ToUpper(must.FilepathAbs("bar")): true,
						},
					},
				},
			},
			wantPathList: []string{
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
		{
			name:  "duplicate-similar-path-case-sensitive",
			key:   "foo",
			value: strings.ToUpper(must.FilepathAbs("bar")),
			pathHandler: &defaultPathHandler{
				caseSensitiveFilesystem: true,
				container: &container{
					pathListElements: map[string][]string{
						"foo": {
							must.FilepathAbs("a"),
							must.FilepathAbs("b"),
							must.FilepathAbs("bar"),
						},
					},
					pathListElementExists: map[string]map[string]bool{
						"foo": {
							must.FilepathAbs("a"):   true,
							must.FilepathAbs("b"):   true,
							must.FilepathAbs("bar"): true,
						},
					},
				},
			},
			wantPathList: []string{
				strings.ToUpper(must.FilepathAbs("bar")),
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			err := tc.pathHandler.Add(tc.key, tc.value, tc.position)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			wantPathList := tc.wantPathList
			havePathList := tc.pathHandler.container.pathListElements[tc.key]

			if assert.Equal(st, len(wantPathList), len(havePathList), "Path list length mismatch") {
				for i := range wantPathList {
					want := wantPathList[i]
					have := havePathList[i]

					assert.Equal(st, want, have, "PathList[%d] value mismatch", i)
				}
			}
		})
	}
}

func Test_defaultPathLoader_Load(t *testing.T) {
	pathA := must.FilepathAbs("a")
	pathB := must.FilepathAbs("b")
	pathC := must.FilepathAbs("c")

	testCases := []*struct {
		name              string
		key               string
		pathLoader        *defaultPathLoader
		mockPathHandlerOn [][]any
		err               *klib.Error
	}{
		{
			name: "path-list-exists",
			key:  "foo",
			pathLoader: &defaultPathLoader{
				container: &container{
					pathListExists: map[string]bool{
						"foo": true,
					},
				},
			},
		},
		{
			name: "empty-value",
			key:  "foo",
			pathLoader: &defaultPathLoader{
				container: &container{
					pathListExists: map[string]bool{},
				},
			},
		},
		{
			name: "create-path-list",
			key:  "foo",
			pathLoader: &defaultPathLoader{
				container: &container{
					curEnv: map[string]string{
						"foo": strings.Join(
							[]string{
								pathA,
								pathB,
								pathC,
							},
							string(os.PathListSeparator),
						),
					},
					pathListExists: map[string]bool{},
				},
			},
			mockPathHandlerOn: [][]any{
				{"Add", "foo", pathA, -1, nil},
				{"Add", "foo", pathB, -1, nil},
				{"Add", "foo", pathC, -1, nil},
			},
		},
		{
			name: "ignore-empty-path-element",
			key:  "foo",
			pathLoader: &defaultPathLoader{
				container: &container{
					curEnv: map[string]string{
						"foo": strings.Join(
							[]string{
								pathA,
								pathC,
							},
							string(os.PathListSeparator),
						),
					},
					pathListExists: map[string]bool{},
				},
			},
			mockPathHandlerOn: [][]any{
				{"Add", "foo", pathA, -1, nil},
				{"Add", "foo", pathC, -1, nil},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			if len(tc.mockPathHandlerOn) > 0 {
				mockPathHandler := NewMockPathHandler(st)

				for j := range tc.mockPathHandlerOn {
					on := tc.mockPathHandlerOn[j]
					mockPathHandler.On(on[0].(string), on[1], on[2], on[3]).Return(on[4])
				}

				tc.pathLoader.pathHandler = mockPathHandler
			}

			err := tc.pathLoader.Load(tc.key)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}
		})
	}
}
