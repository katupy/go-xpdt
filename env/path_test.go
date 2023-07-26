package env

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"
	"go.katupy.io/klib/mucache"
	"go.katupy.io/klib/must"
)

func Test_defaultPathHandler_Add(t *testing.T) {
	testCases := []*struct {
		name         string
		pathHandler  *defaultPathHandler
		envVar       *environVar
		value        string
		position     int
		err          *klib.Error
		wantPathList []string
	}{
		{
			name: "new-path-list",
			envVar: &environVar{
				key: "foo",
			},
			value:       must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{},
			wantPathList: []string{
				must.FilepathAbs("bar"),
			},
		},
		{
			name: "prepend",
			envVar: &environVar{
				key: "foo",
				pathListElements: []string{
					must.FilepathAbs("a"),
					must.FilepathAbs("b"),
				},
			},
			value:       must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{},
			wantPathList: []string{
				must.FilepathAbs("bar"),
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
			},
		},
		{
			name: "append",
			envVar: &environVar{
				key: "foo",
				pathListElements: []string{
					must.FilepathAbs("a"),
					must.FilepathAbs("b"),
				},
			},
			value:       must.FilepathAbs("bar"),
			position:    -1,
			pathHandler: &defaultPathHandler{},
			wantPathList: []string{
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
		{
			name: "ignore-existing-path-case-sensitive",
			envVar: &environVar{
				key: "foo",
				pathListElements: []string{
					must.FilepathAbs("a"),
					must.FilepathAbs("b"),
					must.FilepathAbs("bar"),
				},
				pathListElementExists: map[string]bool{
					must.FilepathAbs("a"):   true,
					must.FilepathAbs("b"):   true,
					must.FilepathAbs("bar"): true,
				},
			},
			value: must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{
				caseSensitiveFilesystem: true,
			},
			wantPathList: []string{
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
		{
			name: "ignore-existing-path-case-insensitive",
			envVar: &environVar{
				key: "foo",
				pathListElements: []string{
					must.FilepathAbs("a"),
					must.FilepathAbs("b"),
					must.FilepathAbs("bar"),
				},
				pathListElementExists: map[string]bool{
					strings.ToUpper(must.FilepathAbs("a")):   true,
					strings.ToUpper(must.FilepathAbs("b")):   true,
					strings.ToUpper(must.FilepathAbs("bar")): true,
				},
			},
			value:       must.FilepathAbs("bar"),
			pathHandler: &defaultPathHandler{},
			wantPathList: []string{
				must.FilepathAbs("a"),
				must.FilepathAbs("b"),
				must.FilepathAbs("bar"),
			},
		},
		{
			name: "duplicate-similar-path-case-sensitive",
			envVar: &environVar{
				key: "foo",
				pathListElements: []string{
					must.FilepathAbs("a"),
					must.FilepathAbs("b"),
					must.FilepathAbs("bar"),
				},
				pathListElementExists: map[string]bool{
					must.FilepathAbs("a"):   true,
					must.FilepathAbs("b"):   true,
					must.FilepathAbs("bar"): true,
				},
			},
			value: strings.ToUpper(must.FilepathAbs("bar")),
			pathHandler: &defaultPathHandler{
				caseSensitiveFilesystem: true,
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
			err := tc.pathHandler.Add(tc.envVar, tc.value, tc.position)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			havePathList := tc.envVar.pathListElements
			wantPathList := tc.wantPathList

			if assert.Equal(st, len(wantPathList), len(havePathList), "Path list length mismatch") {
				for i := range wantPathList {
					have := havePathList[i]
					want := wantPathList[i]

					assert.Equal(st, want, have, "PathList[%d] value mismatch", i)
				}
			}
		})
	}
}

func Test_defaultPathLoader_Load(t *testing.T) {
	cache := mucache.New[string, *environVar]()
	pathA := must.FilepathAbs("a")
	pathB := must.FilepathAbs("b")
	pathC := must.FilepathAbs("c")

	testCases := []*struct {
		name              string
		envVar            *environVar
		pathLoader        *defaultPathLoader
		mockPathHandlerOn [][]any
		err               *klib.Error
	}{
		{
			name: "path-list-exists",
			envVar: &environVar{
				pathList: true,
			},
			pathLoader: &defaultPathLoader{},
		},
		{
			name:       "empty-value",
			envVar:     &environVar{},
			pathLoader: &defaultPathLoader{},
		},
		{
			name: "create-path-list",
			envVar: cache.SetGet(
				"create-path-list",
				&environVar{
					currentValue: strings.Join(
						[]string{
							pathA,
							pathB,
							pathC,
						},
						string(os.PathListSeparator),
					),
				},
			),
			pathLoader: &defaultPathLoader{},
			mockPathHandlerOn: [][]any{
				{"Add", cache.Get("create-path-list"), pathA, -1, nil},
				{"Add", cache.Get("create-path-list"), pathB, -1, nil},
				{"Add", cache.Get("create-path-list"), pathC, -1, nil},
			},
		},
		{
			name: "ignore-empty-path-element",
			envVar: cache.SetGet(
				"ignore-empty-path-element",
				&environVar{
					currentValue: strings.Join(
						[]string{
							pathA,
							"",
							pathC,
						},
						string(os.PathListSeparator),
					),
				},
			),
			pathLoader: &defaultPathLoader{},
			mockPathHandlerOn: [][]any{
				{"Add", cache.Get("ignore-empty-path-element"), pathA, -1, nil},
				{"Add", cache.Get("ignore-empty-path-element"), pathC, -1, nil},
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

			err := tc.pathLoader.Load(tc.envVar)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}
		})
	}
}
