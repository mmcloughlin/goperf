package dashboard

import "github.com/mmcloughlin/cb/pkg/fs"

//go:generate go run make_palette.go -output static/css/palette.css -shades 2
//go:generate go run make_assets.go -pkg dashboard -var assets -output zassets.go templates static

// Embedded asset filesystems.
var (
	AssetFileSystem    = fs.NewMemWithFiles(assets)
	TemplateFileSystem = fs.NewSub(AssetFileSystem, "templates")
	StaticFileSystem   = fs.NewSub(AssetFileSystem, "static")
)
