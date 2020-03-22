// Code generated by make_assets.go. DO NOT EDIT.

package main

var Assets = map[string][]byte{
	"templates/bench.html": []byte("<!DOCTYPE html>\n<html>\n  <head>\n    <title>{{ .Benchmark.Name }}</title>\n  </head>\n  <body>\n    <h1>{{ .Benchmark.Name }}</h1>\n    <table>\n      <tr>\n        <th>Value</th>\n        <th>Commit</th>\n        <th>Commit Time</th>\n        <th>File</th>\n      </tr>\n      {{ range .Results }}\n      <tr>\n        <td>{{ .Value }}</td>\n        <td>{{ .Commit.SHA }}</td>\n        <td>{{ .Commit.CommitTime }}</td>\n        <td>\n          <a href=\"/file/{{ .File.UUID }}#L{{ .Line }}\">{{ .File.UUID }}</a>\n        </td>\n      </tr>\n      {{ end }}\n    </table>\n  </body>\n</html>\n"),
	"templates/file.html":  []byte("<!DOCTYPE html>\n<html>\n  <head>\n    <title>{{ .File.Name }}</title>\n  </head>\n  <body>\n    <h1>{{ .File.Name }}</h1>\n    <pre>\n      {{ range .Lines -}}\n      <span id=\"L{{ .Num }}\">{{ .Num }}</span> {{ .Contents }}\n      {{ end }}\n    </pre>\n  </body>\n</html>\n"),
	"templates/mod.html":   []byte("<!DOCTYPE html>\n<html>\n  <head>\n    <title>{{ .Module.Path }}</title>\n  </head>\n  <body>\n    <h1>{{ .Module.Path }}</h1>\n    <p>Version: {{ .Module.Version }}</p>\n    <ul>\n      {{ range .Packages }}\n      <li>\n        <a href=\"/pkg/{{ .UUID }}\"\n          ><small>{{ .Module.Path }}/</small>{{ .RelativePath }}</a\n        >\n      </li>\n      {{ end }}\n    </ul>\n  </body>\n</html>\n"),
	"templates/mods.html":  []byte("<!DOCTYPE html>\n<html>\n  <head>\n    <title>Modules</title>\n  </head>\n  <body>\n    <ul>\n      {{ range .Modules }}\n      <li>\n        <a href=\"/mod/{{ .UUID }}\">{{ .Path }}</a> <small>{{ .Version }}</small>\n      </li>\n      {{ end }}\n    </ul>\n  </body>\n</html>\n"),
	"templates/pkg.html":   []byte("<!DOCTYPE html>\n<html>\n  <head>\n    <title>{{ .Package.ImportPath }}</title>\n  </head>\n  <body>\n    <h1>{{ .Package.ImportPath }}</h1>\n    <ul>\n      {{ range .Benchmarks }}\n      <li><a href=\"/bench/{{ .UUID }}\">{{ .Name }}</a></li>\n      {{ end }}\n    </ul>\n  </body>\n</html>\n"),
}
