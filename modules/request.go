package modules

import (
	"context"
	"github.com/corpix/uarand"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

func RequestFunc(ip string, url string, timeout int) []string {
	n := time.Now()

	req, err := http.NewRequest("GET", ip, nil)
	if err != nil {
		return []string{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	req.Host = url

	req.Header.Set("User-Agent", uarand.GetRandom())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []string{}
	}

	last, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return []string{}
	}

	return []string{resp.Status, ip, string(last), strconv.FormatInt(time.Since(n).Milliseconds(), 10)}
}
