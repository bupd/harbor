// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package csrf

import (
	"net/http"
	"slices"
	"strings"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
)

// protect judges each request on the origin that browser actually used, so it
// needs no configuration and is safe for concurrent use.
var protect = http.NewCrossOriginProtection()

// safeMethods are the methods RFC 7231 defines as safe, which carry no state
// change and so need no cross-origin check.
var safeMethods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}

// check applies the cross-origin protection, then closes the one gap it leaves
// open by design.
//
// CrossOriginProtection admits a request carrying neither Sec-Fetch-Site nor
// Origin, reading it as non-browser traffic. That is the right default for a
// library, but not here: csrfSkipper has already excused the registry, API and
// service routes that non-browser clients use, so an unsafe request arriving
// without either header is one we cannot vouch for. Refusing it matches what the
// previous token scheme did, which rejected any write that failed to present a
// token.
func check(req *http.Request) error {
	// Without Sec-Fetch-Site, Origin must equal Host port included, so a proxy
	// in front of core has to forward the Host the client addressed.
	if err := protect.Check(req); err != nil {
		return err
	}

	if slices.Contains(safeMethods, req.Method) {
		return nil
	}

	if req.Header.Get("Sec-Fetch-Site") == "" && req.Header.Get("Origin") == "" {
		return errors.New("request carries neither Sec-Fetch-Site nor Origin, cannot confirm its origin")
	}

	return nil
}

// Middleware rejects cross-origin requests that carry a session. It reads the
// Fetch metadata headers browsers have sent since 2023, falling back to
// comparing Origin against the Host the client addressed.
//
// Browsers send Sec-Fetch-Site only to secure origins, so the fallback is the
// main path for plain HTTP from a non-localhost host, besides older browsers.
// It compares host and port but not scheme, and it does not consult
// X-Forwarded-Host. A reverse proxy in front of core must therefore forward the
// Host header unchanged, port included: proxy_set_header Host $http_host in
// nginx, ProxyPreserveHost On in Apache. Clients without a session, such as
// docker or robot accounts on basic auth, are excused by csrfSkipper and are
// not affected.
//
// Deliberately no configured endpoint takes part in the decision. Harbor is
// commonly reachable on several ingresses, over different schemes, behind
// proxies that rewrite Host; a single configured origin describes none of that,
// and every attempt to reconstruct "the" expected origin gets one of those
// deployments wrong.
func Middleware() func(handler http.Handler) http.Handler {
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		if err := check(req); err != nil {
			log.Debugf("Rejected cross-origin request for %s: %v", lib.TrimLineBreaks(req.URL.Path), err)
			lib_http.SendError(rw, errors.New(err).WithCode(errors.ForbiddenCode))
			return
		}

		next.ServeHTTP(rw, req)
	}, csrfSkipper)
}

// csrfSkipper makes sure only some of the uris accessed by non-UI client can skip the csrf check
func csrfSkipper(req *http.Request) bool {
	path := req.URL.Path
	if (strings.HasPrefix(path, "/v2/") ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/service/")) && !lib.GetCarrySession(req.Context()) {
		return true
	}
	return false
}
