package env

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"go.katupy.io/klib"
	"go.katupy.io/klib/mucache"
)

func Test_defaultFileLoader_Load(t *testing.T) {
	cache := mucache.New[string, *File]()

	testCases := []*struct {
		name                string
		file                *File
		fileLoader          *defaultFileLoader
		mockCommandLoaderOn [][]any
		err                 *klib.Error
	}{
		{
			name: "failed-to-chdir",
			file: &File{
				dir: string(os.PathListSeparator),
			},
			fileLoader: &defaultFileLoader{},
			err: &klib.Error{
				ID:     "869494ce-40c2-43f5-a5ef-a5a5d3b6201f",
				Status: http.StatusInternalServerError,
				Code:   klib.CodeFilesystemError,
			},
		},
		{
			name: "ok",
			file: cache.SetGet(
				"ok",
				&File{
					dir: ".",
					Commands: []*Command{
						{},
						{},
						{},
					},
				},
			),
			fileLoader: &defaultFileLoader{},
			mockCommandLoaderOn: [][]any{
				{"Load", &Command{index: 0, file: cache.Get("ok")}, nil},
				{"Load", &Command{index: 1, file: cache.Get("ok")}, nil},
				{"Load", &Command{index: 2, file: cache.Get("ok")}, nil},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			mockCommandLoader := NewMockCommandLoader(st)

			for j := range tc.mockCommandLoaderOn {
				on := tc.mockCommandLoaderOn[j]
				mockCommandLoader.On(on[0].(string), on[1]).Return(on[2])
			}

			tc.fileLoader.commandLoader = mockCommandLoader

			err := tc.fileLoader.Load(tc.file)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}
		})
	}
}
