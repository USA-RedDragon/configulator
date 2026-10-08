//go:build goexperiment.jsonv2

package configulator

import (
	jsonv2 "encoding/json/v2"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// StrictJSON is the default decoder. It rejects unknown members.
func StrictJSON(b []byte, v any) error {
	return jsonv2.Unmarshal(b, v, jsonv2.RejectUnknownMembers(true))
}

func defaultDecoders() Decoders { return Decoders{".json": StrictJSON} }

// decoderFor picks the decoder for path by its lowercased extension.
func decoderFor(d Decoders, path string) (Unmarshal, error) {
	if d == nil {
		d = defaultDecoders()
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return nil, fmt.Errorf("config file %s has no extension; cannot pick a decoder", path)
	}
	u, ok := d[ext]
	if !ok {
		return nil, fmt.Errorf("no decoder configured for %q (file %s)", ext, path)
	}
	return u, nil
}

// resolveFile picks and reads the config file: the config flag if it was
// passed, then opts.Explicit, then opts.Search. An empty file returns nil
// data with res.Path set, which is different from not found.
func resolveFile(opts *FileOptions, readFile func(string) ([]byte, error), flagPath string, flagChanged bool) ([]byte, FileResult, error) {
	res := FileResult{}
	required := ""
	switch {
	case flagChanged:
		required = flagPath
		res.Origin = PathConfigFlag
	case opts.Explicit != "":
		required = opts.Explicit
		res.Origin = PathExplicit
	}

	if required != "" {
		res.Searched = []string{required}
		data, err := readFile(required)
		if err != nil {
			return nil, res, &MissingFileError{Path: required, Searched: opts.Search, Err: err}
		}
		res.Path = required
		if len(data) == 0 {
			return nil, res, nil
		}
		return data, res, nil
	}

	res.Origin = PathSearch
	for _, p := range opts.Search {
		res.Searched = append(res.Searched, p)
		data, err := readFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, res, &SearchPathError{Path: p, Err: err}
		}
		res.Path = p
		if len(data) == 0 {
			return nil, res, nil
		}
		return data, res, nil
	}

	res.Origin = PathNone
	if opts.RequireFound {
		return nil, res, &NoFileFoundError{Searched: res.Searched}
	}
	return nil, res, nil
}

// defaultReadFile returns an error for directories that isn't a not-exist
// error, so a directory in the search paths is reported instead of skipped.
func defaultReadFile(path string) ([]byte, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if st.IsDir() {
		return nil, &fs.PathError{Op: "read", Path: path, Err: fmt.Errorf("is a directory")}
	}
	return os.ReadFile(path)
}
