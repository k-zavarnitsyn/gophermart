package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k-zavarnitsyn/gophermart/internal"
	"github.com/k-zavarnitsyn/gophermart/internal/api"
	"github.com/k-zavarnitsyn/gophermart/internal/app"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	"github.com/k-zavarnitsyn/gophermart/tests/testutils"
	"github.com/stretchr/testify/suite"
)

type header struct {
	header string
	val    string
}

type TestSuite struct {
	suite.Suite

	cfg    *config.Config
	cnt    *container.Container
	api    internal.API
	router *app.Router
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	s.cfg = testutils.GetConfig("../../" + config.DefaultDir)
	s.cnt = container.New(s.cfg)
	s.api = api.New(
		s.cfg,
		s.cnt.Auth(),
		s.cnt.Gophermart(),
		s.cnt.Pinger(),
	)
	s.router = app.NewRouter(s.cnt)
	s.router.InitRoutes(s.api, false)

	s.Require().NoError(s.cnt.Resetter().Reset())
}

func (s *TestSuite) TestApi() {
	empty := ""
	testCases := []struct {
		name         string
		method       string
		path         string
		headers      []header
		body         string
		expectedCode int
		expectedBody string
	}{
		{name: "Ping", method: http.MethodGet, path: "/ping", expectedCode: http.StatusOK, expectedBody: empty},
		{
			name:         "Register",
			method:       http.MethodPost,
			path:         "/api/user/register",
			body:         `{"login":"test","password":"test"}`,
			expectedCode: http.StatusOK,
			expectedBody: empty,
		},
		{
			name:         "Register with no login",
			method:       http.MethodPost,
			path:         "/api/user/register",
			body:         `{"password":"test"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: empty,
		},
		{
			name:         "Login",
			method:       http.MethodPost,
			path:         "/api/user/login",
			body:         `{"login":"test","password":"test"}`,
			expectedCode: http.StatusOK,
			expectedBody: empty,
		},
		// TODO
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			r := httptest.NewRequest(tc.method, tc.path, http.NoBody)
			for _, header := range tc.headers {
				r.Header.Add(header.header, header.val)
			}
			if tc.body != "" {
				r.Body = io.NopCloser(strings.NewReader(tc.body))
			}
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, r)

			s.Assert().Equal(tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				s.Assert().Equal(tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}
