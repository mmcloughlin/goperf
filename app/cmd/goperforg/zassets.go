// Code generated by make_assets.go. DO NOT EDIT.

package main

var Assets = map[string][]byte{
	"templates/layout/headfoot.gohtml": []byte("{{ define \"header\" }}\n<!DOCTYPE html>\n<html>\n  <head>\n    <link href=\"https://fonts.googleapis.com/css?family=Work+Sans:600|Roboto:400,700|Source+Code+Pro\" rel=\"stylesheet\">\n    <link href=\"/static/css/style.css\" rel=\"stylesheet\">\n\n    <title>GoPerf</title>\n  </head>\n  <body>\n    {{ end }}\n    {{ define \"footer\" }}\n  </body>\n</html>\n{{ end }}\n"),
	"templates/view/bench.gohtml":      []byte("{{ define \"bench\" }}\n{{ template \"header\" }}\n\n<h1>{{ .Benchmark.Name }}</h1>\n<table>\n  <tr>\n    <th>Value</th>\n    <th>Commit</th>\n    <th>Commit Time</th>\n    <th>File</th>\n  </tr>\n  {{ range .Results }}\n  <tr>\n    <td>{{ .Value }}</td>\n    <td>{{ .Commit.SHA }}</td>\n    <td>{{ .Commit.CommitTime }}</td>\n    <td>\n      <a href=\"/file/{{ .File.UUID }}#L{{ .Line }}\">{{ .File.UUID }}</a>\n    </td>\n  </tr>\n  {{ end }}\n</table>\n\n{{ template \"footer\" }}\n{{ end }}\n"),
	"templates/view/file.gohtml":       []byte("{{ define \"file\" }}\n{{ template \"header\" }}\n\n    <h1>{{ .File.Name }}</h1>\n    <pre>\n      {{ range .Lines -}}\n      <span id=\"L{{ .Num }}\">{{ .Num }}</span> {{ .Contents }}\n      {{ end }}\n    </pre>\n\n{{ template \"footer\" }}\n{{ end }}\n"),
	"templates/view/mod.gohtml":        []byte("{{ define \"mod\" }}\n{{ template \"header\" }}\n\n    <h1>{{ .Module.Path }}</h1>\n    <p>Version: {{ .Module.Version }}</p>\n    <ul>\n      {{ range .Packages }}\n      <li>\n        <a href=\"/pkg/{{ .UUID }}\"\n          ><small>{{ .Module.Path }}/</small>{{ .RelativePath }}</a\n        >\n      </li>\n      {{ end }}\n    </ul>\n\n{{ template \"footer\" }}\n{{ end }}\n"),
	"templates/view/mods.gohtml":       []byte("{{ define \"mods\" }}\n{{ template \"header\" }}\n\n    <ul>\n      {{ range .Modules }}\n      <li>\n        <a href=\"/mod/{{ .UUID }}\">{{ .Path }}</a> <small>{{ .Version }}</small>\n      </li>\n      {{ end }}\n    </ul>\n\n{{ template \"footer\" }}\n{{ end }}\n"),
	"templates/view/pkg.gohtml":        []byte("{{ define \"pkg\" }}\n{{ template \"header\" }}\n\n    <h1>{{ .Package.ImportPath }}</h1>\n    <ul>\n      {{ range .Benchmarks }}\n      <li><a href=\"/bench/{{ .UUID }}\">{{ .Name }}</a></li>\n      {{ end }}\n    </ul>\n\n{{ template \"footer\" }}\n{{ end }}\n"),
	"static/css/style.css":             []byte("/*\n\nStyle derived from https://pkg.go.dev/static/css/stylesheet.css and\nhttps://blog.golang.org/go-brand.\n\nOriginal work is Copyright 2019 The Go Authors and BSD-3 licensed\n(https://golang.org/LICENSE).\n\n*/\n\n:root {\n  --gray-1: #202224;\n  --gray-2: #3e4042;\n  --gray-3: #555759;\n  --gray-4: #6e7072;\n  --gray-5: #848688;\n  --gray-6: #aaacae;\n  --gray-7: #c6c8ca;\n  --gray-8: #dcdee0;\n  --gray-9: #f0f1f2;\n  --gray-10: #fafafa;\n\n  --turq-light: #5dc9e2;\n  --turq-med: #00add8;\n  --turq-text: #007d9c;\n\n  --blue: #92e1f3;\n  --green: #00a29c;\n  --pink: #ce3262;\n  --purple: #542c7d;\n  --slate: #253443;\n  --white: #fff;\n  --yellow: #fddd00;\n}\n\nhtml {\n  height: 100%;\n}\n\nbody {\n  color: var(--gray-1);\n  font-family: Roboto, Arial, sans-serif;\n  margin: 0;\n}\n\na,\na:link,\na:visited {\n  color: var(--turq-text);\n  text-decoration: none;\n}\n\na:hover {\n  text-decoration: underline;\n}\n\nh1,\nh2,\nh3,\nh4,\nh5,\nh6 {\n  font-family: \"Work Sans\", Arial, sans-serif;\n}\n\nh1,\nh2,\nh3 {\n  font-weight: bold;\n}\n\nh1 {\n  font-size: 1.5rem;\n}\n\nh2 {\n  font-size: 1.125rem;\n}\n\nh3 {\n  font-size: 1rem;\n}\n\np {\n  font-size: 1rem;\n  line-height: 1.5rem;\n}\n\ncode {\n  font-size: 1rem;\n  font-family: \"Go Mono\", \"Source Code Pro\", monospace;\n}\n"),
	"static/img/go-logo-white.svg":     []byte("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<!-- Copyright 2018 The Go Authors. All rights reserved. -->\n<!-- Generator: Adobe Illustrator 22.1.0, SVG Export Plug-In . SVG Version: 6.00 Build 0)  -->\n<svg version=\"1.1\" id=\"Layer_1\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" x=\"0px\" y=\"0px\"\n\t viewBox=\"0 0 254.5 225\" style=\"enable-background:new 0 0 254.5 225;\" xml:space=\"preserve\">\n<style type=\"text/css\">\n\t.st0{fill:#2DBCAF;}\n\t.st1{fill:#5DC9E1;}\n\t.st2{fill:#FDDD00;}\n\t.st3{fill:#CE3262;}\n\t.st4{fill:#00ACD7;}\n\t.st5{fill:#FFFFFF;}\n</style>\n<g>\n\t<g>\n\t\t<g>\n\t\t\t<g>\n\t\t\t\t<path class=\"st5\" d=\"M40.2,101.1c-0.4,0-0.5-0.2-0.3-0.5l2.1-2.7c0.2-0.3,0.7-0.5,1.1-0.5l35.7,0c0.4,0,0.5,0.3,0.3,0.6\n\t\t\t\t\tl-1.7,2.6c-0.2,0.3-0.7,0.6-1,0.6L40.2,101.1z\"/>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n\t<g>\n\t\t<g>\n\t\t\t<g>\n\t\t\t\t<path class=\"st5\" d=\"M25.1,110.3c-0.4,0-0.5-0.2-0.3-0.5l2.1-2.7c0.2-0.3,0.7-0.5,1.1-0.5l45.6,0c0.4,0,0.6,0.3,0.5,0.6\n\t\t\t\t\tl-0.8,2.4c-0.1,0.4-0.5,0.6-0.9,0.6L25.1,110.3z\"/>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n\t<g>\n\t\t<g>\n\t\t\t<g>\n\t\t\t\t<path class=\"st5\" d=\"M49.3,119.5c-0.4,0-0.5-0.3-0.3-0.6l1.4-2.5c0.2-0.3,0.6-0.6,1-0.6l20,0c0.4,0,0.6,0.3,0.6,0.7l-0.2,2.4\n\t\t\t\t\tc0,0.4-0.4,0.7-0.7,0.7L49.3,119.5z\"/>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n\t<g>\n\t\t<g id=\"CXHf1q_2_\">\n\t\t\t<g>\n\t\t\t\t<g>\n\t\t\t\t\t<path class=\"st5\" d=\"M153.1,99.3c-6.3,1.6-10.6,2.8-16.8,4.4c-1.5,0.4-1.6,0.5-2.9-1c-1.5-1.7-2.6-2.8-4.7-3.8\n\t\t\t\t\t\tc-6.3-3.1-12.4-2.2-18.1,1.5c-6.8,4.4-10.3,10.9-10.2,19c0.1,8,5.6,14.6,13.5,15.7c6.8,0.9,12.5-1.5,17-6.6\n\t\t\t\t\t\tc0.9-1.1,1.7-2.3,2.7-3.7c-3.6,0-8.1,0-19.3,0c-2.1,0-2.6-1.3-1.9-3c1.3-3.1,3.7-8.3,5.1-10.9c0.3-0.6,1-1.6,2.5-1.6\n\t\t\t\t\t\tc5.1,0,23.9,0,36.4,0c-0.2,2.7-0.2,5.4-0.6,8.1c-1.1,7.2-3.8,13.8-8.2,19.6c-7.2,9.5-16.6,15.4-28.5,17\n\t\t\t\t\t\tc-9.8,1.3-18.9-0.6-26.9-6.6c-7.4-5.6-11.6-13-12.7-22.2c-1.3-10.9,1.9-20.7,8.5-29.3c7.1-9.3,16.5-15.2,28-17.3\n\t\t\t\t\t\tc9.4-1.7,18.4-0.6,26.5,4.9c5.3,3.5,9.1,8.3,11.6,14.1C154.7,98.5,154.3,99,153.1,99.3z\"/>\n\t\t\t\t</g>\n\t\t\t\t<g>\n\t\t\t\t\t<path class=\"st5\" d=\"M186.2,154.6c-9.1-0.2-17.4-2.8-24.4-8.8c-5.9-5.1-9.6-11.6-10.8-19.3c-1.8-11.3,1.3-21.3,8.1-30.2\n\t\t\t\t\t\tc7.3-9.6,16.1-14.6,28-16.7c10.2-1.8,19.8-0.8,28.5,5.1c7.9,5.4,12.8,12.7,14.1,22.3c1.7,13.5-2.2,24.5-11.5,33.9\n\t\t\t\t\t\tc-6.6,6.7-14.7,10.9-24,12.8C191.5,154.2,188.8,154.3,186.2,154.6z M210,114.2c-0.1-1.3-0.1-2.3-0.3-3.3\n\t\t\t\t\t\tc-1.8-9.9-10.9-15.5-20.4-13.3c-9.3,2.1-15.3,8-17.5,17.4c-1.8,7.8,2,15.7,9.2,18.9c5.5,2.4,11,2.1,16.3-0.6\n\t\t\t\t\t\tC205.2,129.2,209.5,122.8,210,114.2z\"/>\n\t\t\t\t</g>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n</g>\n</svg>\n"),
}
