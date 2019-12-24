package geoip

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~diamondburned/geotz/geotz"
	"git.sr.ht/~diamondburned/geotz/records"
	"gitlab.com/shihoya-inc/errchi"
)

var Records geotz.Records

func init() {
	r, err := records.DecodeBytes(records.Raw)
	if err == nil {
		Records = r
	}
}

func Middleware(next errchi.Handler) errchi.Handler {
	return errchi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
		if addr := r.RemoteAddr; addr != "" {
			ip := net.ParseIP(strings.Split(addr, ":")[0])

			if my := Records.GetRecord(ip); my != nil {
				return next.ServeHTTP(w, r.WithContext(
					context.WithValue(r.Context(), "tz", my.Location())))
			}
		}

		var fwd = r.Header.Get("X-Forwarded-For")

		for _, addr := range strings.Split(fwd, ", ") {
			if my := Records.GetRecord(net.ParseIP(addr)); my != nil {
				return next.ServeHTTP(w, r.WithContext(
					context.WithValue(r.Context(), "tz", my.Location())))
			}
		}

		return next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), "tz", time.Local)))
	})
}
