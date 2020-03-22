package main

import "github.com/mmcloughlin/cb/pkg/fs"

//go:generate go run make_assets.go -pkg main -var Assets -output zassets.go templates/*

// AssetFileSystem returns a read-only filesystem.
func AssetFileSystem() fs.Readable {
	return fs.NewSub(fs.NewMemWithFiles(Assets), "templates")
}
