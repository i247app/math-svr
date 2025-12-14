package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/requtil"
)

type requestLoggerMiddleware struct {
	hiddenFieldsRegex *regexp.Regexp
	logHeaders        bool
	mutex             sync.Mutex
	reqID             int64
}

func LogRequestMiddleware(next http.Handler) http.Handler {
	middleware := newRequestLoggerMiddleware()
	return middleware.Handle(next)
}

func newRequestLoggerMiddleware() *requestLoggerMiddleware {
	hiddenFields := []string{
		"image_data",
		"image_data_back",
		"img_front_data",
		"img_back_data",
		"img_url_front",
		"img_url_back",
		"doc_url",
		"doc_data",
	}

	return &requestLoggerMiddleware{
		hiddenFieldsRegex: regexp.MustCompile(`"(` + strings.Join(hiddenFields, "|") + `)":\s*"(?:[^"\\]|\\.)*"`),
		logHeaders:        true,
	}
}

func (m *requestLoggerMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			reqID int64 = m.nextRequestID()
		)

		logger := logger.GetLogger(r.Context())

		rawBody, isJSON, err := m.readRequestBody(r)
		if err != nil {
			logger.Errorf("logRequestMiddleware: read request body error: %v", err)
			return
		}

		logger.Infof("IN <%v> %v %v", reqID, r.Method, r.URL.Path)

		if r.Method == http.MethodGet && strings.TrimSpace(r.URL.RawQuery) != "" {
			logger.Infof("IN QUERY PARAMS %v", r.URL.RawQuery)
		}

		if isJSON {
			truncatedBodyBytes := m.truncateSensitiveFields(rawBody.Bytes())
			logger.Infof("IN REQUEST BODY %v", string(truncatedBodyBytes))
		} else {
			mapBody := m.decodeBodyToMap(rawBody.Bytes())
			if len(mapBody) > 0 {
				logger.Infof("IN REQUEST BODY %v", mapBody)
			}
		}

		if m.logHeaders {
			for name, values := range r.Header {
				for _, value := range values {
					logger.Infof("IN HEADER %v: %v", name, value)
				}
			}
		}

		if metadata := m.requestMetadata(r); metadata != nil {
			logger.Infof("IN __metadata %v", metadata)
		}

		wrapper := m.newResponseWrapper(w)
		next.ServeHTTP(wrapper, r)

		outMsg := m.outboundMessage(r, wrapper)
		logger.Infof("%s", outMsg)

		if m.logHeaders {
			if h := wrapper.Header().Get("X-Auth-Token"); h != "" {
				logger.Infof("OUT HEADER %v: %v", "X-Auth-Token", h)
			} else {
				logger.Infof("OUT HEADER %v: %v", "X-Auth-Token", "<empty>")
			}
		}

		m.flushResponse(w, wrapper)
	})
}

func (m *requestLoggerMiddleware) nextRequestID() int64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	next := m.reqID
	m.reqID++
	return next
}

func (m *requestLoggerMiddleware) readRequestBody(r *http.Request) (*bytes.Buffer, bool, error) {
	rawBody := new(bytes.Buffer)
	if _, err := rawBody.ReadFrom(r.Body); err != nil {
		return nil, false, fmt.Errorf("read request body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(rawBody.Bytes()))

	var jsonCheck any
	err := json.Unmarshal(rawBody.Bytes(), &jsonCheck)
	return rawBody, err == nil, nil
}

func (m *requestLoggerMiddleware) truncateSensitiveFields(body []byte) []byte {
	if m.hiddenFieldsRegex == nil {
		return body
	}

	return m.hiddenFieldsRegex.ReplaceAll(body, []byte(`"$1": <...>`))
}

func (m *requestLoggerMiddleware) decodeBodyToMap(body []byte) map[string]any {
	result := map[string]any{}
	_ = json.Unmarshal(body, &result)
	return result
}

func (m *requestLoggerMiddleware) requestMetadata(r *http.Request) *requtil.RequestMetadata {
	wrapped, err := requtil.Wrap(r)
	if err != nil {
		return nil
	}

	metadata, err := wrapped.Metadata()
	if err != nil {
		return nil
	}

	return metadata
}

func (m *requestLoggerMiddleware) newResponseWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		body:           bytes.NewBuffer(nil),
	}
}

func (m *requestLoggerMiddleware) outboundMessage(r *http.Request, wrapper *responseWriterWrapper) string {
	if strings.Contains(wrapper.Header().Get("Content-Type"), "application/json") {
		return fmt.Sprintf("OUT %v %v: %v", r.Method, r.URL.Path, wrapper.body.String())
	}

	inMsg := fmt.Sprintf("IN %v %v", r.Method, r.URL.Path)
	return inMsg
}

func (m *requestLoggerMiddleware) flushResponse(w http.ResponseWriter, wrapper *responseWriterWrapper) {
	if wrapper.statusCode != 0 {
		w.WriteHeader(wrapper.statusCode)
	}
	_, _ = w.Write(wrapper.body.Bytes())
}
