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

type logCount struct {
	reqId int64
	mutex sync.Mutex
}

func (l *logCount) Increment() int64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tmp := l.reqId
	l.reqId++
	return tmp
}

// LogRequestMiddleware wraps an http.Handler to log the API endpoint details
func LogRequestMiddleware(next http.Handler) http.Handler {
	var (
		logCount        = logCount{0, sync.Mutex{}}
		logInOutHeaders = true
	)

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
	hiddenFieldsRegex := regexp.MustCompile(`"(` + strings.Join(hiddenFields, "|") + `)":\s*"(?:[^"\\]|\\.)*"`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := logCount.Increment()

		// read the response into the buffer
		rawBody := new(bytes.Buffer)
		_, err := rawBody.ReadFrom(r.Body)
		if err != nil {
			logger.Error(err)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(rawBody.Bytes()))

		// Check if body is valid JSON (without indenting)
		var jsonCheck any
		err = json.Unmarshal(rawBody.Bytes(), &jsonCheck)
		isJson := err == nil

		identifier := "anon"
		// if sess := session.GetRequestSession(r); sess != nil {
		// 	uid, ok := sess.UID()
		// 	if ok {
		// 		identifier = strconv.FormatInt(uid, 10)
		// 	}

		// 	email, ok := sess.Get("email")
		// 	if ok {
		// 		if emailStr, ok := email.(string); ok {
		// 			identifier = identifier + ":" + emailStr
		// 		}
		// 	}

		// 	isSecure, ok := sess.Get("is_secure")
		// 	if ok {
		// 		if isSecureBool, ok := isSecure.(bool); ok && !isSecureBool {
		// 			identifier = "*" + identifier
		// 		}
		// 	}
		// }

		inMsg := fmt.Sprintf("IN [%v] <%v> %v %v", reqID, identifier, r.Method, r.URL.Path)
		logger.Infof("%s", inMsg)
		if r.Method == "GET" {
			logger.Infof("?%v", r.URL.RawQuery)
		}

		if isJson {
			// Truncate certain fields using regex replacement directly on the raw bytes
			truncatedBodyBytes := hiddenFieldsRegex.ReplaceAll(rawBody.Bytes(), []byte(`"$1": <...>`)) // Use $1 to preserve the matched key name
			logger.Infof(": %s", truncatedBodyBytes)
		} else {
			// unmarshal the buffer into your map[string]any
			mapBody := new(map[string]any)
			json.Unmarshal(rawBody.Bytes(), mapBody)

			if len(*mapBody) > 0 {
				logger.Infof(": %v", *mapBody)
			}
		}

		if logInOutHeaders {
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				logger.Infof("IN HEADER> ", "Authorization", authHeader)
			}
		}

		// Log metadata
		if wrapped, err := requtil.Wrap(r); err == nil {
			if metadata, err := wrapped.Metadata(); err == nil {
				logger.Infof("__metadata: %v\n", metadata)
			}
		}

		// Create a wrapper to capture the response body
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			body:           bytes.NewBuffer(nil),
		}

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// If the response body is not JSON, do not print response body
		var outMsg string
		if strings.Contains(wrapper.Header().Get("Content-Type"), "application/json") {
			outMsg = fmt.Sprintf("OUT [%v] <%v> %v %v: %v", reqID, identifier, r.Method, r.URL.Path, wrapper.body.String())
		} else {
			outMsg = fmt.Sprintf("OUT [%v] <%v> %v %v: <>\n", reqID, identifier, r.Method, r.URL.Path)
		}
		logger.Infof("%s", outMsg)

		if logInOutHeaders {
			// log.Println("OUT STATUS> ", wrapper.statusCode)
			if h := wrapper.Header().Get("X-Auth-Token"); h != "" {
				logger.Infof("OUT HEADER> ", "X-Auth-Token", h)
			} else {
				logger.Infof("OUT HEADER> ", "X-Auth-Token", "<empty>")
			}
		}

		// NOTE: You can't write an invalid status code or it will cause a runtime error,
		// and 0 is the most likely invalid status code.
		if wrapper.statusCode != 0 {
			w.WriteHeader(wrapper.statusCode)
		}
		w.Write(wrapper.body.Bytes())
	})
}
