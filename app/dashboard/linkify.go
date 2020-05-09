package dashboard

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"mvdan.cc/xurls/v2"
)

// linkify converts URLs from whitelisted hosts into HTML links.
func linkify(s string) template.HTML {
	return template.HTML(urlregexp.ReplaceAllStringFunc(s, func(match string) string {
		u, err := url.Parse(match)
		if err != nil {
			return match
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return match
		}
		if !linkifyhost(u.Host) {
			return match
		}
		return fmt.Sprintf(`<a href="%s">%s</a>`, u, u)
	}))
}

// Regular expression for extracting URLs.
var urlregexp = xurls.Strict()

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
