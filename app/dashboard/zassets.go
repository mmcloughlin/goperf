// Code generated by make_assets.go. DO NOT EDIT.

package dashboard

var assets = map[string][]byte{
	"templates/bench.gohtml":             []byte("{{ define \"head\" }}\n<script type=\"text/javascript\" src=\"https://www.gstatic.com/charts/loader.js\"></script>\n<script type=\"text/javascript\">\n  google.charts.load('current', {'packages':['corechart']});\n  google.charts.setOnLoadCallback(drawChart);\n\n  points = [\n    {{ range .Points -}}\n    { resultUUID: {{ .ResultUUID | js }}, commitSHA: {{ .CommitSHA | js }} },\n    {{ end }}\n  ]\n\n  function drawChart () {\n    var data = google.visualization.arrayToDataTable([\n      ['commit_index', 'value', 'filtered'],\n      {{ range $idx, $point := .Points -}}\n      [{{ $idx }}, {{ $point.Value }}, {{ index $.Filtered $idx }}],\n      {{ end }}\n    ]);\n\n    var options = {\n      chartArea: {\n        width: '85%',\n        height: '95%'\n      },\n      hAxis: { textPosition: 'none' },\n      axisTitlesPosition: 'none',\n      legend: { position: 'none' },\n      series: [\n        { color: {{ color \"gopher-blue\" | js }}, dataOpacity: 0.5, pointSize: 8 },\n        { color: {{ color \"fuchsia\" | js }}, lineWidth: 3, pointSize: 0, enableInteractivity: false },\n      ],\n      tooltip: { trigger: 'selection' }\n    };\n\n    var chart = new google.visualization.ScatterChart(document.getElementById('chart'));\n\n    chart.setAction({\n      id: 'result',\n      text: 'View Result',\n      action: function() {\n        selection = chart.getSelection();\n        idx = selection[0].row;\n        window.location.href = '/result/' + points[idx].resultUUID;\n      }\n    });\n\n    chart.setAction({\n      id: 'commit',\n      text: 'View Commit',\n      action: function() {\n        selection = chart.getSelection();\n        idx = selection[0].row;\n        window.location.href = '/commit/' + points[idx].commitSHA;\n      }\n    });\n\n    chart.draw(data, options);\n  }\n</script>\n{{ end }}\n\n{{ define \"content\" }}\n<h1>{{ .Benchmark.FullName }}{{ template \"sep\" }}{{ .Benchmark.Unit }}</h1>\n\n<div class=\"meta\">\n  <dl>\n    <dt>Package</dt> <dd>{{ template \"pkg\" .Benchmark.Package }}</dd>\n    <dt>Version</dt> <dd>{{ .Benchmark.Package.Module.Version }}</dd>\n  </dl>\n</div>\n\n<div id=\"chart\"></div>\n\n{{ end }}\n"),
	"templates/commit.gohtml":            []byte("{{ define \"content\" }}\n{{ with .Commit }}\n<h1>commit {{ .SHA }}</h1>\n<pre>\n{{ .Message }}\n</pre>\n{{ end }}\n{{ end }}\n"),
	"templates/file.gohtml":              []byte("{{ define \"content\" }}\n<h1>file {{ .File.UUID }}</h1>\n<pre>\n  {{ range .Lines -}}\n  <span class=\"ln\" id=\"L{{ .Num }}\">{{ .Num }}</span>\n  {{- if .Highlight -}}\n  <span class=\"hl\">{{ .Contents }}</span>\n  {{- else -}}\n  {{ .Contents }}\n  {{- end }}\n  {{ end }}\n</pre>\n{{ end }}\n"),
	"templates/layout/components.gohtml": []byte("{{ define \"mod\" }}<a href=\"/mod/{{ .UUID }}\">{{ .Path }}</a>{{ end }}\n{{ define \"pkg\" }}<a href=\"/pkg/{{ .UUID }}\">{{ .ImportPath }}</a>{{ end }}\n{{ define \"bench\" }}<a href=\"/bench/{{ .UUID }}\">{{ .FullName }}</a>{{ end }}\n{{ define \"commit\" }}<a href=\"/commit/{{ .SHA }}\"><code>{{ slice .SHA 0 10 }}</code></a>{{ end }}\n{{ define \"file\" }}<a href=\"/file/{{ .UUID }}\"><code>{{ template \"uuidshort\" .UUID }}</code></a>{{ end }}\n{{ define \"loc\" }}<a href=\"/file/{{ .File.UUID }}?hl={{ .Line }}#L{{ .Line }}\"><code>{{ template \"uuidshort\" .File.UUID }}#{{ .Line }}</code></a>{{ end }}\n\n{{ define \"uuidshort\" }}{{ slice .String 0 8 }}{{ end }}\n{{ define \"sep\" }} <span class=\"sep\">/</span> {{ end }}\n\n{{ define \"properties\" }}\n<table class=\"properties\">\n{{ range $key, $value := . }}\n    <tr>\n        <td class=\"key code\">{{ $key }}</td>\n        <td class=\"value\">{{ $value }}</td>\n    </tr>\n{{ end }}\n</table>\n{{ end }}\n"),
	"templates/layout/main.gohtml":       []byte("{{ define \"main\" }}\n<!DOCTYPE html>\n<html>\n  <head>\n    <link href=\"https://fonts.googleapis.com/css?family=Work+Sans:600|Roboto:400,700|Source+Code+Pro\" rel=\"stylesheet\" />\n    <link href=\"/static/css/style.css\" rel=\"stylesheet\" />\n    <link rel=\"icon\" href=\"/static/img/favicon.ico\" type=\"image/x-icon\" />\n    {{ block \"head\" . }}{{ end }}\n    <title>{{ block \"title\" . }}GoPerf{{ end }}</title>\n  </head>\n  <body>\n    <header>\n      <nav>\n        <a href=\"/\"><img class=\"logo\" src=\"/static/img/go-logo-white.svg\" alt=\"Go\" /></a>\n      </nav>\n    </header>\n    <main>\n    {{ block \"content\" . }}{{ end }}\n    </main>\n  </body>\n</html>\n{{ end }}\n"),
	"templates/mod.gohtml":               []byte("{{ define \"content\" }}\n<h1>module {{ .Module.Path }}</h1>\n<div class=\"meta\">\n  <dl>\n    <dt>Version</dt> <dd>{{ .Module.Version }}</dd>\n  </dl>\n</div>\n<ul>\n  {{ range .Packages }}\n  <li>{{ template \"pkg\" . }}</li>\n  {{ end }}\n</ul>\n{{ end }}\n"),
	"templates/mods.gohtml":              []byte("{{ define \"content\" }}\n<h1>Modules</h1>\n<table>\n  <tr>\n    <th>Module</th>\n    <th>Version</th>\n  </tr>\n  {{ range .Modules }}\n  <tr>\n    <td>{{ template \"mod\" . }}</td>\n    <td>{{ .Version }}</td>\n  </tr>\n  {{ end }}\n</table>\n{{ end }}\n"),
	"templates/pkg.gohtml":               []byte("{{ define \"content\" }}\n<h1>package {{ .Package.ImportPath }}</h1>\n<div class=\"meta\">\n  <dl>\n    <dt>Module</dt> <dd>{{ template \"mod\" .Package.Module }}</dd>\n  </dl>\n</div>\n<ul>\n  {{ range .Benchmarks }}\n  <li><a href=\"/bench/{{ .UUID }}\">{{ .Name }}</a></li>\n  {{ end }}\n</ul>\n{{ end }}\n"),
	"templates/result.gohtml":            []byte("{{ define \"content\" }}\n\n<h1>{{ .Result.Benchmark.FullName }} {{ template \"sep\" }} Result</h1>\n\n<div class=\"meta\">\n  <dl>\n    <dt>Benchmark</dt> <dd>{{ template \"bench\" .Result.Benchmark }}</dd>\n    <dt>Package</dt> <dd>{{ template \"pkg\" .Result.Benchmark.Package }}</dd>\n    <dt>Commit</dt> <dd>{{ template \"commit\" .Result.Commit }}</dd>\n    <dt>Source</dt> <dd>{{ template \"loc\" .Result }}</dd>\n  </dl>\n</div>\n\n{{ with .Quantity }}\n<div class=\"bignumber\">\n    <p class=\"number\">{{ .Value }}</p>\n    <p class=\"unit\">{{ .Unit }}</p>\n</div>\n{{ end }}\n\n{{ with .Result }}\n{{ if .Environment }}\n<h2>Environment</h2>\n{{ template \"properties\" .Environment }}\n{{ end }}\n\n{{ if .Metadata }}\n<h2>Metadata</h2>\n{{ template \"properties\" .Metadata }}\n{{ end }}\n{{ end }}\n\n{{ end }}\n"),
	"static/css/palette.css":             []byte("/* Generated by make_palette.go. DO NOT EDIT. */\n\n:root {\n\t--aqua: #00A29C;\n\t--aqua-1: #17fff6;\n\t--aqua-2: #8bfffb;\n\t--black: #000000;\n\t--black-1: #555555;\n\t--black-2: #aaaaaa;\n\t--cool-gray: #DBD9D6;\n\t--cool-gray-1: #e7e6e4;\n\t--cool-gray-2: #f3f2f1;\n\t--fuchsia: #CE3262;\n\t--fuchsia-1: #de7696;\n\t--fuchsia-2: #efbbcb;\n\t--gopher-blue: #00ADD8;\n\t--gopher-blue-1: #3bd8ff;\n\t--gopher-blue-2: #9debff;\n\t--light-blue: #5DC9E2;\n\t--light-blue-1: #93dbec;\n\t--light-blue-2: #c9edf5;\n\t--purple: #402B56;\n\t--purple-1: #7f56aa;\n\t--purple-2: #bfaad5;\n\t--slate: #555759;\n\t--slate-1: #8c8f92;\n\t--slate-2: #c6c7c8;\n\t--turquoise: #00758D;\n\t--turquoise-1: #09d5ff;\n\t--turquoise-2: #84eaff;\n\t--yellow: #FDDD00;\n\t--yellow-1: #ffe954;\n\t--yellow-2: #fff4a9;\n}\n"),
	"static/css/style.css":               []byte("/*\n\nStyle derived from https://pkg.go.dev/static/css/stylesheet.css and\nhttps://blog.golang.org/go-brand.\n\nOriginal work is Copyright 2019 The Go Authors and BSD-3 licensed\n(https://golang.org/LICENSE).\n\n*/\n\n@import url(\"/static/css/palette.css\");\n\nhtml {\n  height: 100%;\n}\n\nbody {\n  font-family: Roboto, Arial, sans-serif;\n  margin: 0;\n}\n\na,\na:link,\na:visited {\n  color: var(--turquoise);\n  text-decoration: none;\n}\n\na:hover {\n  text-decoration: underline;\n}\n\nh1,\nh2,\nh3,\nh4,\nh5,\nh6 {\n  font-family: \"Work Sans\", Arial, sans-serif;\n}\n\nh1,\nh2,\nh3 {\n  font-weight: bold;\n}\n\nh1 {\n  font-size: 1.5rem;\n}\n\nh2 {\n  font-size: 1.125rem;\n}\n\nh3 {\n  font-size: 1rem;\n}\n\np {\n  font-size: 1rem;\n  line-height: 1.3em;\n}\n\ncode,\npre,\n.code {\n  font-family: \"Go Mono\", \"Source Code Pro\", monospace;\n  font-size: 0.875rem;\n}\n\npre {\n  background-color: var(--cool-gray-2);\n  overflow-x: auto;\n  padding: 0.625rem;\n  border-radius: 0.3em;\n  border: 0.0625rem solid var(--cool-gray);\n}\n\n.ln {\n  color: var(--slate);\n  display: inline-block;\n  text-align: right;\n  width: 6ch;\n  padding-right: 2ch;\n  user-select: none;\n}\n\n.hl {\n  background-color: var(--yellow);\n}\n\ntable {\n  border-collapse: collapse;\n  width: 100%;\n}\n\ntd,\nth {\n  border-bottom: 1px solid var(--cool-gray);\n  padding: 0.75rem 0;\n  padding-right: 1rem;\n}\n\nth {\n  text-align: left;\n}\n\n.sep {\n  color: var(--slate);\n  font-weight: bold;\n}\n\nheader {\n  background: var(--turquoise);\n  border: none;\n  margin: 0;\n  padding: 0;\n}\n\nheader nav {\n  margin: 0 auto;\n  max-width: 70em;\n}\n\nheader .logo {\n  display: block;\n  width: 5rem;\n  margin: 0;\n  padding: 0;\n}\n\nmain {\n  margin: 0 auto;\n  max-width: 60em;\n}\n\n#chart {\n  width: 100%;\n  height: 500px;\n}\n\n.meta {\n  overflow-x: hidden;\n  font-size: 0.875rem;\n}\n\n.meta dl {\n  position: relative;\n  left: -0.75rem;\n}\n\n.meta dt,\n.meta dd {\n  display: inline-block;\n  padding: 0;\n  margin: 0;\n}\n\n.meta dt {\n  border-left: 1px solid var(--slate);\n  padding-left: 0.75rem;\n}\n\n.meta dt::after {\n  content: \":\";\n}\n\n.meta dd {\n  padding-right: 0.75rem;\n}\n\n.bignumber {\n  text-align: center;\n}\n\n.bignumber .number {\n  font-size: 10rem;\n  margin: 0.2em 0;\n}\n\n.bignumber .unit {\n  font-size: 5rem;\n  color: var(--slate);\n  margin-top: 0;\n}\n\ntable.properties td.key {\n  color: var(--slate);\n  white-space: nowrap;\n  font-weight: bold;\n  text-transform: lowercase;\n}\n\ntable.properties td.value {\n  overflow-wrap: anywhere;\n}\n"),
	"static/img/go-logo-white.svg":       []byte("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<!-- Copyright 2018 The Go Authors. All rights reserved. -->\n<!-- Generator: Adobe Illustrator 22.1.0, SVG Export Plug-In . SVG Version: 6.00 Build 0)  -->\n<svg version=\"1.1\" id=\"Layer_1\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" x=\"0px\" y=\"0px\"\n\t viewBox=\"0 0 254.5 225\" style=\"enable-background:new 0 0 254.5 225;\" xml:space=\"preserve\">\n<style type=\"text/css\">\n\t.st0{fill:#2DBCAF;}\n\t.st1{fill:#5DC9E1;}\n\t.st2{fill:#FDDD00;}\n\t.st3{fill:#CE3262;}\n\t.st4{fill:#00ACD7;}\n\t.st5{fill:#FFFFFF;}\n</style>\n<g>\n\t<g>\n\t\t<g>\n\t\t\t<g>\n\t\t\t\t<path class=\"st5\" d=\"M40.2,101.1c-0.4,0-0.5-0.2-0.3-0.5l2.1-2.7c0.2-0.3,0.7-0.5,1.1-0.5l35.7,0c0.4,0,0.5,0.3,0.3,0.6\n\t\t\t\t\tl-1.7,2.6c-0.2,0.3-0.7,0.6-1,0.6L40.2,101.1z\"/>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n\t<g>\n\t\t<g>\n\t\t\t<g>\n\t\t\t\t<path class=\"st5\" d=\"M25.1,110.3c-0.4,0-0.5-0.2-0.3-0.5l2.1-2.7c0.2-0.3,0.7-0.5,1.1-0.5l45.6,0c0.4,0,0.6,0.3,0.5,0.6\n\t\t\t\t\tl-0.8,2.4c-0.1,0.4-0.5,0.6-0.9,0.6L25.1,110.3z\"/>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n\t<g>\n\t\t<g>\n\t\t\t<g>\n\t\t\t\t<path class=\"st5\" d=\"M49.3,119.5c-0.4,0-0.5-0.3-0.3-0.6l1.4-2.5c0.2-0.3,0.6-0.6,1-0.6l20,0c0.4,0,0.6,0.3,0.6,0.7l-0.2,2.4\n\t\t\t\t\tc0,0.4-0.4,0.7-0.7,0.7L49.3,119.5z\"/>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n\t<g>\n\t\t<g id=\"CXHf1q_2_\">\n\t\t\t<g>\n\t\t\t\t<g>\n\t\t\t\t\t<path class=\"st5\" d=\"M153.1,99.3c-6.3,1.6-10.6,2.8-16.8,4.4c-1.5,0.4-1.6,0.5-2.9-1c-1.5-1.7-2.6-2.8-4.7-3.8\n\t\t\t\t\t\tc-6.3-3.1-12.4-2.2-18.1,1.5c-6.8,4.4-10.3,10.9-10.2,19c0.1,8,5.6,14.6,13.5,15.7c6.8,0.9,12.5-1.5,17-6.6\n\t\t\t\t\t\tc0.9-1.1,1.7-2.3,2.7-3.7c-3.6,0-8.1,0-19.3,0c-2.1,0-2.6-1.3-1.9-3c1.3-3.1,3.7-8.3,5.1-10.9c0.3-0.6,1-1.6,2.5-1.6\n\t\t\t\t\t\tc5.1,0,23.9,0,36.4,0c-0.2,2.7-0.2,5.4-0.6,8.1c-1.1,7.2-3.8,13.8-8.2,19.6c-7.2,9.5-16.6,15.4-28.5,17\n\t\t\t\t\t\tc-9.8,1.3-18.9-0.6-26.9-6.6c-7.4-5.6-11.6-13-12.7-22.2c-1.3-10.9,1.9-20.7,8.5-29.3c7.1-9.3,16.5-15.2,28-17.3\n\t\t\t\t\t\tc9.4-1.7,18.4-0.6,26.5,4.9c5.3,3.5,9.1,8.3,11.6,14.1C154.7,98.5,154.3,99,153.1,99.3z\"/>\n\t\t\t\t</g>\n\t\t\t\t<g>\n\t\t\t\t\t<path class=\"st5\" d=\"M186.2,154.6c-9.1-0.2-17.4-2.8-24.4-8.8c-5.9-5.1-9.6-11.6-10.8-19.3c-1.8-11.3,1.3-21.3,8.1-30.2\n\t\t\t\t\t\tc7.3-9.6,16.1-14.6,28-16.7c10.2-1.8,19.8-0.8,28.5,5.1c7.9,5.4,12.8,12.7,14.1,22.3c1.7,13.5-2.2,24.5-11.5,33.9\n\t\t\t\t\t\tc-6.6,6.7-14.7,10.9-24,12.8C191.5,154.2,188.8,154.3,186.2,154.6z M210,114.2c-0.1-1.3-0.1-2.3-0.3-3.3\n\t\t\t\t\t\tc-1.8-9.9-10.9-15.5-20.4-13.3c-9.3,2.1-15.3,8-17.5,17.4c-1.8,7.8,2,15.7,9.2,18.9c5.5,2.4,11,2.1,16.3-0.6\n\t\t\t\t\t\tC205.2,129.2,209.5,122.8,210,114.2z\"/>\n\t\t\t\t</g>\n\t\t\t</g>\n\t\t</g>\n\t</g>\n</g>\n</svg>\n"),
}
