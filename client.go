package cryptoproxy

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type Client struct {
	limiter *Limiter
}

func NewClient(limiter *Limiter) *Client {
	return &Client{limiter}
}

func (c *Client) Call(req *http.Request, timeLimitMs int, limitCallback func()) (*http.Response, error) {
	if timeLimitMs == 0 {
		c.limiter.Wait()
		return c.call(req)
	}

	return c.callWithLimit(req, timeLimitMs, limitCallback)
}

func (c *Client) callWithLimit(req *http.Request, timeLimitMs int, limitCallback func()) (*http.Response, error) {
	type Response struct {
		res *http.Response
		err error
	}

	timer := time.NewTimer(time.Duration(timeLimitMs) * time.Millisecond)
	limitReached := false

	// wait for rate limit first
	select {
	case <-c.limiter.Wait():
	case <-timer.C:
		log.Debug().Str("path", req.URL.Path).Msg("call limit reached due to rate limiter")
		limitCallback()
		limitReached = true
	}

	// If time limit reached, just make a call after waiting for limiter
	if limitReached {
		c.limiter.Wait()
		return c.call(req)
	}

	var result Response
	ch := make(chan Response, 1)

	// launch async request
	go func() {
		res, err := c.call(req)
		ch <- Response{res, err}
	}()

	// wait for whatever finishes first
	select {
	case result = <-ch:
	case <-timer.C:
		log.Debug().Str("path", req.URL.Path).Msg("call limit reached due to slowness of endpoint")

		limitCallback()
		limitReached = true
	}

	// time limit reached, but we would like to wait for response
	if limitReached {
		result = <-ch
	}

	return result.res, result.err
}

func (c *Client) call(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
