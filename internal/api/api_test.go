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
		// {name: "Get instead of post", method: http.MethodGet, path: "/update/counter/test/1", expectedCode: http.StatusMethodNotAllowed, expectedBody: empty},
		// {name: "Put instead of post", method: http.MethodPut, path: "/update/counter/test/1", expectedCode: http.StatusMethodNotAllowed, expectedBody: empty},
		// {name: "Delete instead of post", method: http.MethodDelete, path: "/update/counter/test/1", expectedCode: http.StatusMethodNotAllowed, expectedBody: empty},
		// {name: "Post new counter", method: http.MethodPost, path: "/update/counter/test/1", expectedCode: http.StatusOK, expectedBody: empty},
		// {name: "Update counter", method: http.MethodPost, path: "/update/counter/test/10", expectedCode: http.StatusOK, expectedBody: empty},
		// {name: "Post new gauge", method: http.MethodPost, path: "/update/gauge/test/1", expectedCode: http.StatusOK, expectedBody: empty},
		// {name: "Update gauge", method: http.MethodPost, path: "/update/gauge/test/10", expectedCode: http.StatusOK, expectedBody: empty},
		// {name: "Get counter", method: http.MethodGet, path: "/value/counter/test", expectedCode: http.StatusOK, expectedBody: "11"},
		// {name: "Get gauge", method: http.MethodGet, path: "/value/gauge/test", expectedCode: http.StatusOK, expectedBody: "10"},
		// {name: "Get all", method: http.MethodGet, path: "/", expectedCode: http.StatusOK, expectedBody: s.getSelectAllBodyResponse1()},
		// // end of test cases for getSelectAllBodyResponse1
		//
		// {
		// 	name:         "Update gauge JSON",
		// 	method:       http.MethodPost,
		// 	path:         "/update/",
		// 	headers:      []header{{header: api.ContentType, val: api.ContentTypeJSON}},
		// 	body:         entity.Gophermart{ID: "test", MType: entity.TypeGauge, Value: utils.ToPointer(15.0)}.String(),
		// 	expectedCode: http.StatusOK,
		// 	expectedBody: `{"id":"test","type":"gauge","value":15}`,
		// },
		// {
		// 	name:         "Update counter JSON",
		// 	method:       http.MethodPost,
		// 	path:         "/update/",
		// 	headers:      []header{{header: api.ContentType, val: api.ContentTypeJSON}},
		// 	body:         entity.Gophermart{ID: "test", MType: entity.TypeCounter, Delta: utils.ToPointer(int64(10))}.String(),
		// 	expectedCode: http.StatusOK,
		// 	expectedBody: `{"id":"test","type":"counter","delta":21}`,
		// },
		// {
		// 	name:         "Get gauge JSON",
		// 	method:       http.MethodPost,
		// 	path:         "/value/",
		// 	headers:      []header{{header: api.ContentType, val: api.ContentTypeJSON}},
		// 	body:         entity.Gophermart{ID: "test", MType: entity.TypeGauge}.String(),
		// 	expectedCode: http.StatusOK,
		// 	expectedBody: `{"id":"test","type":"gauge","value":15}`,
		// },
		// {
		// 	name:         "Get counter JSON",
		// 	method:       http.MethodPost,
		// 	path:         "/value/",
		// 	headers:      []header{{header: api.ContentType, val: api.ContentTypeJSON}},
		// 	body:         entity.Gophermart{ID: "test", MType: entity.TypeCounter}.String(),
		// 	expectedCode: http.StatusOK,
		// 	expectedBody: `{"id":"test","type":"counter","delta":21}`,
		// },
		// {
		// 	name:    "Update batch JSON",
		// 	method:  http.MethodPost,
		// 	path:    "/updates/",
		// 	headers: []header{{header: api.ContentType, val: api.ContentTypeJSON}},
		// 	body: "[" + strings.Join([]string{
		// 		entity.Gophermart{ID: "test", MType: entity.TypeCounter, Delta: utils.ToPointer(int64(4))}.String(),
		// 		entity.Gophermart{ID: "test", MType: entity.TypeCounter, Delta: utils.ToPointer(int64(5))}.String(),
		// 		entity.Gophermart{ID: "test2", MType: entity.TypeCounter, Delta: utils.ToPointer(int64(1))}.String(),
		// 		entity.Gophermart{ID: "test2", MType: entity.TypeCounter, Delta: utils.ToPointer(int64(1))}.String(),
		// 		entity.Gophermart{ID: "test", MType: entity.TypeGauge, Value: utils.ToPointer(15.0)}.String(),
		// 		entity.Gophermart{ID: "test", MType: entity.TypeGauge, Value: utils.ToPointer(13.0)}.String(),
		// 		entity.Gophermart{ID: "test2", MType: entity.TypeGauge, Value: utils.ToPointer(1.0)}.String(),
		// 		entity.Gophermart{ID: "test2", MType: entity.TypeGauge, Value: utils.ToPointer(13.0)}.String(),
		// 	}, ", ") + "]",
		// 	expectedCode: http.StatusOK,
		// 	expectedBody: empty,
		// },
		// {name: "Get counter", method: http.MethodGet, path: "/value/counter/test", expectedCode: http.StatusOK, expectedBody: "30"},
		// {name: "Get counter", method: http.MethodGet, path: "/value/counter/test2", expectedCode: http.StatusOK, expectedBody: "2"},
		// {name: "Get gauge", method: http.MethodGet, path: "/value/gauge/test", expectedCode: http.StatusOK, expectedBody: "13"},
		// {name: "Get gauge", method: http.MethodGet, path: "/value/gauge/test2", expectedCode: http.StatusOK, expectedBody: "13"},
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
