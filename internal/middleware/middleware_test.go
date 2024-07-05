package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	"github.com/k-zavarnitsyn/gophermart/internal/middleware"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/tests/testutils"
	"github.com/stretchr/testify/suite"
)

type header struct {
	header string
	val    string
}

type TestSuite struct {
	suite.Suite

	cfg *config.Config
	cnt *container.Container
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	s.cfg = testutils.GetConfig("../../" + config.DefaultDir)
	s.cnt = container.New(s.cfg)
}

func (s *TestSuite) EchoResponse(w http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(w, r.Body)
	if err != nil {
		utils.SendInternalError(w, err, "unable to copy request body to response")
		return
	}
}

func (s *TestSuite) TestGzip() {
	router := chi.NewRouter()
	router.Use(middleware.WithGzipRequest, middleware.WithGzipResponse)
	router.Post("/", s.EchoResponse)

	rawData := []byte("Testttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt")

	compressedData, err := utils.Compress(rawData)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		method       string
		headers      []header
		body         []byte
		expectedCode int
		expectedBody []byte
	}{
		{
			name:         "Get gzipped response",
			method:       http.MethodPost,
			headers:      []header{{header: utils.AcceptEncoding, val: "gzip"}},
			body:         rawData,
			expectedCode: http.StatusOK,
			expectedBody: compressedData.Bytes(),
		},
		{
			name:         "Decompress gzipped request",
			method:       http.MethodPost,
			headers:      []header{{header: utils.ContentEncoding, val: "gzip"}},
			body:         compressedData.Bytes(),
			expectedCode: http.StatusOK,
			expectedBody: rawData,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			r := httptest.NewRequest(tc.method, "/", http.NoBody)
			for _, header := range tc.headers {
				r.Header.Add(header.header, header.val)
			}
			if tc.body != nil {
				r.Body = io.NopCloser(bytes.NewReader(tc.body))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			s.Assert().Equal(tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != nil {
				s.Assert().Equal(tc.expectedBody, w.Body.Bytes(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}
