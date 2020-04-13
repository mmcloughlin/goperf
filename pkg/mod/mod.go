package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	pathpkg "path"
	"strings"
	"time"

	"github.com/golang/groupcache/lru"
	"golang.org/x/mod/module"

	"github.com/mmcloughlin/cb/internal/errutil"
)

// Reference: https://github.com/golang/go/blob/b5c66de0892d0e9f3f59126eeebc31070e79143b/src/cmd/go/internal/modfetch/repo.go#L56-L65
//
//	// A Rev describes a single revision in a module repository.
//	type RevInfo struct {
//		Version string    // suggested version string for this revision
//		Time    time.Time // commit time
//
//		// These fields are used for Stat of arbitrary rev,
//		// but they are not recorded when talking about module versions.
//		Name  string `json:"-"` // complete ID in underlying repository
//		Short string `json:"-"` // shortened ID, for use in pseudo-version
//	}
//

// RevInfo describes a single revision to a module repository.
type RevInfo struct {
	Version string
	Time    time.Time
}

type ModuleDatabase interface {
	Stat(ctx context.Context, path, rev string) (*RevInfo, error)
}

type modcache struct {
	cache *lru.Cache
	mod   ModuleDatabase
}

// NewModuleCache provides an in-memory cache in front of a ModuleDatabase.
func NewModuleCache(mod ModuleDatabase, maxentries int) ModuleDatabase {
	return &modcache{
		cache: lru.New(maxentries),
		mod:   mod,
	}
}

func (c *modcache) Stat(ctx context.Context, path, rev string) (*RevInfo, error) {
	type key struct {
		path, rev string
	}
	k := key{path, rev}

	if info, ok := c.cache.Get(k); ok {
		return info.(*RevInfo), nil
	}

	info, err := c.mod.Stat(ctx, path, rev)
	if err != nil {
		return nil, err
	}

	c.cache.Add(k, info)
	return info, nil
}

type modproxy struct {
	c   *http.Client
	url *url.URL
}

// NewModuleProxy builds a module database backed by the module proxy API at the
// supplied base URL, using the given HTTP client for requests.
func NewModuleProxy(c *http.Client, u *url.URL) ModuleDatabase {
	return &modproxy{
		c:   c,
		url: u,
	}
}

// NewOfficialModuleProxy builds a module database backed by the official module
// proxy at https://proxy.golang.org, using the given client for requests.
func NewOfficialModuleProxy(c *http.Client) ModuleDatabase {
	return NewModuleProxy(c, &url.URL{
		Scheme: "https",
		Host:   "proxy.golang.org",
	})
}

// Reference: https://github.com/golang/go/blob/b5c66de0892d0e9f3f59126eeebc31070e79143b/src/cmd/go/internal/modfetch/proxy.go#L354-L374
//
//	func (p *proxyRepo) Stat(rev string) (*RevInfo, error) {
//		encRev, err := module.EscapeVersion(rev)
//		if err != nil {
//			return nil, p.versionError(rev, err)
//		}
//		data, err := p.getBytes("@v/" + encRev + ".info")
//		if err != nil {
//			return nil, p.versionError(rev, err)
//		}
//		info := new(RevInfo)
//		if err := json.Unmarshal(data, info); err != nil {
//			return nil, p.versionError(rev, err)
//		}
//		if info.Version != rev && rev == module.CanonicalVersion(rev) && module.Check(p.path, rev) == nil {
//			// If we request a correct, appropriate version for the module path, the
//			// proxy must return either exactly that version or an error — not some
//			// arbitrary other version.
//			return nil, p.versionError(rev, fmt.Errorf("proxy returned info for version %s instead of requested version", info.Version))
//		}
//		return info, nil
//	}
//

func (p *modproxy) Stat(ctx context.Context, path, rev string) (*RevInfo, error) {
	// Apply escaping rules.
	path, err := module.EscapePath(path)
	if err != nil {
		return nil, err
	}

	v, err := module.EscapeVersion(rev)
	if err != nil {
		return nil, err
	}

	// Issue request.
	endpoint := pathpkg.Join(path, "@v", v+".info")
	b, err := p.get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	// Unmarshal.
	info := new(RevInfo)
	if err := json.Unmarshal(b, info); err != nil {
		return nil, err
	}

	// Validate. If revision was already in canonical form it should have been returned as-is.
	if info.Version != rev && rev == module.CanonicalVersion(rev) && module.Check(path, rev) == nil {
		return nil, fmt.Errorf("proxy returned info for version %q instead of requested version", info.Version)
	}

	return info, nil
}

func (p *modproxy) get(ctx context.Context, path string) (_ []byte, err error) {
	// Build URL.
	target := *p.url
	target.Path = pathpkg.Join(target.Path, path)
	target.RawPath = pathpkg.Join(target.RawPath, pathescape(path))

	// Issue request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := p.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer errutil.CheckClose(&err, res.Body)

	return ioutil.ReadAll(res.Body)
}

func pathescape(s string) string {
	// Reference: https://github.com/golang/go/blob/b5c66de0892d0e9f3f59126eeebc31070e79143b/src/cmd/go/internal/modfetch/proxy.go#L432-L437
	//
	//	// pathEscape escapes s so it can be used in a path.
	//	// That is, it escapes things like ? and # (which really shouldn't appear anyway).
	//	// It does not escape / to %2F: our REST API is designed so that / can be left as is.
	//	func pathEscape(s string) string {
	//		return strings.ReplaceAll(url.PathEscape(s), "%2F", "/")
	//	}
	//
	return strings.ReplaceAll(url.PathEscape(s), "%2F", "/")
}
