package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"math-ai.com/math-ai/internal/session"
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
		// var (
		// 	reqID int64 = m.nextRequestID()
		// )

		rawBody, isJSON, err := m.readRequestBody(r)
		if err != nil {
			log.Printf("logRequestMiddleware: read request body error: %v", err)
			return
		}

		token, identifier := m.requestIdentifier(r)

		// Extract last 6 characters of token (or entire token if shorter)
		tokenDisplay := token
		if len(token) > 6 {
			tokenDisplay = token[len(token)-6:]
		}

		inMsg := fmt.Sprintf("IN [%v] [%v] %v %v", tokenDisplay, identifier, r.Method, r.URL.Path)
		log.Printf("%s", inMsg)

		if r.Method == http.MethodGet && strings.TrimSpace(r.URL.RawQuery) != "" {
			msg := fmt.Sprintf("IN [%v] [%v] QUERY PARAMS> %v", tokenDisplay, identifier, r.URL.RawQuery)
			log.Println(msg)
		}

		if isJSON {
			truncatedBodyBytes := m.truncateSensitiveFields(rawBody.Bytes())
			msg := fmt.Sprintf("IN [%v] [%v] REQUEST BODY> %s", tokenDisplay, identifier, truncatedBodyBytes)
			log.Println(msg)
		} else {
			mapBody := m.decodeBodyToMap(rawBody.Bytes())
			if len(mapBody) > 0 {
				msg := fmt.Sprintf("IN [%v] [%v] REQUEST BODY> %v", tokenDisplay, identifier, mapBody)
				log.Println(msg)
			}
		}

		if m.logHeaders {
			for name, values := range r.Header {
				for _, value := range values {
					msg := fmt.Sprintf("IN [%v] [%v] HEADER> %v: %v", tokenDisplay, identifier, name, value)
					log.Println(msg)
				}
			}
		}

		if metadata := m.requestMetadata(r); metadata != nil {
			msg := fmt.Sprintf("IN [%v] [%v] __metadata: %v", tokenDisplay, identifier, metadata)
			log.Println(msg)
		}

		wrapper := m.newResponseWrapper(w)
		next.ServeHTTP(wrapper, r)

		outMsg := m.outboundMessage(tokenDisplay, identifier, r, wrapper)
		log.Printf("%s", outMsg)

		if m.logHeaders {
			if h := wrapper.Header().Get("X-Auth-Token"); h != "" {
				msg := fmt.Sprintf("OUT [%v] [%v] HEADER> %v: %v", tokenDisplay, identifier, "X-Auth-Token", h)
				log.Println(msg)
			} else {
				msg := fmt.Sprintf("OUT [%v] [%v] HEADER> %v: %v", tokenDisplay, identifier, "X-Auth-Token", "<empty>")
				log.Println(msg)
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

func (m *requestLoggerMiddleware) requestIdentifier(r *http.Request) (string, string) {
	identifier := "anon"
	token := "anon"

	if sess := session.GetRequestSession(r); sess != nil {
		if uid, ok := sess.UID(); ok {
			identifier = strconv.FormatInt(uid, 10)
		}

		if key, ok := sess.Get("key"); ok {
			if keyStr, ok := key.(string); ok {
				token = keyStr
			}
		}
	}

	return token, identifier
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

func (m *requestLoggerMiddleware) outboundMessage(tokenDisplay string, identifier string, r *http.Request, wrapper *responseWriterWrapper) string {
	if strings.Contains(wrapper.Header().Get("Content-Type"), "application/json") {
		return fmt.Sprintf("OUT [%v] [%v] %v %v: %v", tokenDisplay, identifier, r.Method, r.URL.Path, wrapper.body.String())
	}

	inMsg := fmt.Sprintf("IN [%v] [%v] %v %v", tokenDisplay, identifier, r.Method, r.URL.Path)
	return inMsg
}

func (m *requestLoggerMiddleware) flushResponse(w http.ResponseWriter, wrapper *responseWriterWrapper) {
	if wrapper.statusCode != 0 {
		w.WriteHeader(wrapper.statusCode)
	}
	_, _ = w.Write(wrapper.body.Bytes())
}
