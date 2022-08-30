package cryptoproxy

import (
	"net/http"
	"net/url"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func BuildChain(f http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) == 0 {
		return f
	}

	return m[0](BuildChain(f, m[1:cap(m)]...))
}

func NewFilterMiddleware(allowedPaths []string) Middleware {
	pathIndex := make(map[string]struct{})
	for _, p := range allowedPaths {
		pathIndex[p] = struct{}{}
	}

	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, ok := pathIndex[r.URL.Path]; !ok {
				http.NotFound(w, r)

				return
			}

			f(w, r)
		}
	}
}

func NewRequestEnricherMiddleware(origin *url.URL, token string) Middleware {
	const authHeaderParam = "Authorization"

	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// prepare request for redirection to origin
			r.Host = origin.Host
			r.URL.Host = origin.Host
			r.URL.Scheme = origin.Scheme
			r.RequestURI = ""

			// set auth token before sending to origin
			r.Header[authHeaderParam] = append(r.Header[authHeaderParam], token)

			f(w, r)
		}
	}
}
