package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func DummyServer(w http.ResponseWriter, r *http.Request) {}

func EchoServer(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fmt.Println(query)
}

func TimeoutServer(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1101 * time.Millisecond)
}

// код писать тут
func TestServerErrors(t *testing.T) {

	dummySrv := httptest.NewServer(http.HandlerFunc(DummyServer))
	timeoutSrv := httptest.NewServer(http.HandlerFunc(TimeoutServer))

	testErrorsTable := []struct {
		req           SearchRequest
		server        *httptest.Server
		expectedError error
		desc          string
	}{
		{SearchRequest{-10, 1, "male", "gender", OrderByAsc}, dummySrv, fmt.Errorf("limit must be > 0"), "check if Limit < 0"},
		{SearchRequest{10, -11, "male", "gender", OrderByAsc}, dummySrv, fmt.Errorf("offset must be > 0"), "check if offset < 0"},
	}

	for _, test := range testErrorsTable {

		testServer := test.server

		testClient := SearchClient{
			URL:         testServer.URL,
			AccessToken: "ZHOPA",
		}

		_, err := testClient.FindUsers(test.req)

		if err == nil || err.Error() != test.expectedError.Error() {
			t.Errorf("Test: %s - unexpected err, Got %v, Want %v", test.desc, err, test.expectedError)
		}

	}

	t.Run("test timeout error", func(t *testing.T) {

		testClient := SearchClient{
			URL:         timeoutSrv.URL,
			AccessToken: "ZHOPA",
		}

		timeoutRequest := SearchRequest{10, 1, "male", "gender", OrderByAsc}
		values := url.Values{}

		values.Add("limit", strconv.Itoa(timeoutRequest.Limit+1))
		values.Add("offset", strconv.Itoa(timeoutRequest.Offset))
		values.Add("query", timeoutRequest.Query)
		values.Add("order_field", timeoutRequest.OrderField)
		values.Add("order_by", strconv.Itoa(timeoutRequest.OrderBy))

		expectedErr := fmt.Errorf("timeout for %s", values.Encode())

		_, err := testClient.FindUsers(timeoutRequest)

		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("Test: %s - unexpected err, Got %v, Want %v", t.Name(), err, expectedErr)
		}

	})

}
