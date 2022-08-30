package cryptoproxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type response struct {
	code int
	body io.Reader
}

// TODO: add max number of inflight requests
type InMemoryExecutor struct {
	clients                 map[string]*Client
	ingressRequestTimeoutMs int
}

func NewInMemoryExecutor(clients map[string]*Client, ingressRequestTimeoutMs int) *InMemoryExecutor {
	return &InMemoryExecutor{clients, ingressRequestTimeoutMs}
}

func (s *InMemoryExecutor) Handler(rw http.ResponseWriter, req *http.Request) {
	respCh := make(chan response, 1)
	go s.asyncHandler(respCh, req.Clone(context.Background()))
	resp := <-respCh

	rw.WriteHeader(resp.code)
	if resp.body != nil {
		io.Copy(rw, resp.body)
	}
}

func (s *InMemoryExecutor) asyncHandler(respCh chan response, req *http.Request) {
	client := s.clients[req.URL.Path]
	isStreamClosed := false
	acceptHandler := func() {
		isStreamClosed = true
		respCh <- response{
			code: http.StatusAccepted,
		}
	}

	// Try to make first call directly
	res, err := client.Call(req, s.ingressRequestTimeoutMs, acceptHandler)

	successHandler := func() {
		// request successfully executed, check output stream and return response
		if !isStreamClosed {
			isStreamClosed = true
			respCh <- response{
				code: res.StatusCode,
				body: res.Body,
			}
		} else if req.URL.Query().Has("callback") {
			_, err := http.Post(req.URL.Query().Get("callback"), "text/plain", res.Body)
			if err != nil {
				log.Warn().Err(err).Msg("failed to call callback")
			}
		}
	}

	if isSucces(res, err) {
		successHandler()
		return
	}

	if !isRetriableFailure(res, err) {
		if !isStreamClosed {
			isStreamClosed = true

			status := http.StatusInternalServerError
			var message io.Reader = strings.NewReader("")

			if err == nil { // restore original response
				status = res.StatusCode
				message = res.Body
			}

			respCh <- response{
				code: status,
				body: message,
			}
		}

		log.Warn().
			Err(err).
			Str("method", req.Method).
			Str("uri", req.RequestURI).
			Msg("request failed with non-retriable error")

		return
	}

	if !isStreamClosed {
		acceptHandler()
	}

	log.Debug().Msg("initial request failed, enteting retry loop")
	// TODO: extract magic numbers to config
	if err := retry(100, 30*time.Second, 5*time.Minute, func() error {
		res, err := client.Call(req, 0, nil)

		// for simplicty we will assume that once we entered retry loop, we should just continue retrying
		if isSucces(res, err) {
			successHandler()
			log.Debug().Msg("request succedded after several retries")
			return nil
		} else if !isRetriableFailure(res, err) {
			log.Warn().Msg("request was dropped after retry, due to non-retriable error")
			return nil
		}

		// TODO: return constant error
		return fmt.Errorf("")
	}); err != nil {
		log.Error().Err(err).Msg("final attempt failed")
	}
}

func isSucces(res *http.Response, err error) bool {
	return err == nil && res.StatusCode == http.StatusOK
}

// For simplicty we will retry only 5XX exceptions and 429 (too many requests)
func isRetriableFailure(res *http.Response, err error) bool {
	return err == nil && (res.StatusCode == http.StatusTooManyRequests ||
		res.StatusCode >= http.StatusInternalServerError /* 500 */)
}

func retry(attempts int, sleep time.Duration, maxSleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep *= 2
		}

		if err = f(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
