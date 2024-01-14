package allsrv

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// ObserveHandler provides observability to an http handler.
func ObserveHandler(name string, met *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &handlerMW{
			name: name,
			next: next,
			met:  met,
		}
	}
}

type handlerMW struct {
	name string
	next http.Handler
	met  *metrics.Metrics
}

func (h *handlerMW) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "http_request_"+h.name)
	defer span.Finish()
	span.LogFields(log.String("url_path", r.URL.Path))

	start := time.Now()
	name := []string{metricsPrefix, h.name, r.URL.Path}

	labels := []metrics.Label{
		{
			Name:  "method",
			Value: r.Method,
		},
		{
			Name:  "url_path",
			Value: r.URL.Path,
		},
	}

	h.met.IncrCounterWithLabels(append(name, "reqs"), 1, labels)

	reqBody := &readRec{ReadCloser: r.Body}
	r.Body = reqBody

	rec := &responseWriterRec{ResponseWriter: w}

	h.next.ServeHTTP(rec, r.WithContext(ctx))

	if rec.code == 0 {
		rec.code = http.StatusOK
	}

	labels = append(labels,
		metrics.Label{
			Name:  "status",
			Value: strconv.Itoa(rec.code),
		},
		metrics.Label{
			Name:  "request_body_size",
			Value: strconv.Itoa(reqBody.size),
		},
		metrics.Label{
			Name:  "response_body_size",
			Value: strconv.Itoa(rec.size),
		},
	)
	if rec.code > 299 {
		h.met.IncrCounterWithLabels(append(name, "errs"), 1, labels)
	}

	h.met.MeasureSinceWithLabels(append(name, "dur"), start, labels)
}

type readRec struct {
	size int
	io.ReadCloser
}

func (r *readRec) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.size += n
	return n, err
}

type responseWriterRec struct {
	size int
	code int
	http.ResponseWriter
}

func (r *responseWriterRec) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

func (r *responseWriterRec) WriteHeader(statusCode int) {
	r.code = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
