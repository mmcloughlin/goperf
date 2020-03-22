package main

import "github.com/mmcloughlin/cb/pkg/fs"

//go:generate go run make_assets.go -pkg main -var Assets -output zassets.go templates

// Embedded asset filesystems.
var (
	AssetFileSystem    = fs.NewMemWithFiles(Assets)
	TemplateFileSystem = fs.NewSub(AssetFileSystem, "templates")
)
