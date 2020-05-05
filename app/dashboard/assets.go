package dashboard

import "github.com/mmcloughlin/cb/pkg/fs"

//go:generate go run make_palette.go -style static/css/style.css -shades 5
//go:generate wget --quiet -O static/img/favicon.ico https://github.com/golang/go/raw/b0da26a668fd6d4e351a00ca76695c5a233e84a2/favicon.ico
//go:generate go run make_assets.go -pkg dashboard -var assets -output zassets.go templates static

// Embedded asset filesystems.
var (
	AssetFileSystem    = fs.NewMemWithFiles(assets)
	TemplateFileSystem = fs.NewSub(AssetFileSystem, "templates")
	StaticFileSystem   = fs.NewSub(AssetFileSystem, "static")
)
