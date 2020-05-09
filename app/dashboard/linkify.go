package dashboard

import (
	"fmt"
	"html"
	"html/template"
	"net/url"
	"strings"

	"mvdan.cc/xurls/v2"
)

// linkify converts URLs from whitelisted hosts into HTML links.
func linkify(s string) template.HTML {
	output := ""
	i := 0
	matches := urlregexp.FindAllStringIndex(s, -1)
	for _, idxs := range matches {
		start, end := idxs[0], idxs[1]
		output += html.EscapeString(s[i:start])
		output += linkreplace(s[start:end])
		i = end
	}
	output += html.EscapeString(s[i:])
	return template.HTML(output)
}

// Regular expression for extracting URLs.
var urlregexp = xurls.Strict()

func linkreplace(match string) string {
	linkhtml := html.EscapeString(match)
	u, err := url.Parse(match)
	if err != nil {
		return linkhtml
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return linkhtml
	}
	if !linkifyhost(u.Host) {
		return linkhtml
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, u, linkhtml)
}

// URLs from linkifyhosts and their subdomains will be turned into links.
var linkifyhosts = []string{
	"golang.org",
	"googlesource.com",
	"google.com",
	"github.com",
}

func linkifyhost(host string) bool {
	for _, accept := range linkifyhosts {
		if host == accept {
			return true
		}
		if strings.HasSuffix(host, "."+accept) {
			return true
		}
	}
	return false
}
