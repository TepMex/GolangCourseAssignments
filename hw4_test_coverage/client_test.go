package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
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

type SpyRequestServer struct {
	request url.Values
	handler http.HandlerFunc
}

func (srv *SpyRequestServer) SpyRequest(w http.ResponseWriter, r *http.Request) {
	srv.request = r.URL.Query()
}

func responseServerFabric(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func jsonErrorServerFabric(errorText string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonWithError, _ := json.Marshal(SearchErrorResponse{errorText})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonWithError)
	}))
}

func successfulServerFabric(response []User) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonWithoutError, _ := json.Marshal(response)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonWithoutError)
	}))
}

// код писать тут
func TestServerErrors(t *testing.T) {

	ordinalReq := SearchRequest{10, 11, "male", "gender", OrderByAsc}

	dummySrv := httptest.NewServer(http.HandlerFunc(DummyServer))
	timeoutSrv := httptest.NewServer(http.HandlerFunc(TimeoutServer))

	testErrorsTable := []struct {
		req           SearchRequest
		server        *httptest.Server
		expectedError error
		desc          string
	}{
		{SearchRequest{-10, 1, "male", "gender", OrderByAsc}, dummySrv, fmt.Errorf("limit must be > 0"), "check error if Limit < 0"},
		{SearchRequest{10, -11, "male", "gender", OrderByAsc}, dummySrv, fmt.Errorf("offset must be > 0"), "check error if offset < 0"},
		{ordinalReq, responseServerFabric(http.StatusUnauthorized), fmt.Errorf("Bad AccessToken"), "check error if unauthorized"},
		{ordinalReq, responseServerFabric(http.StatusInternalServerError), fmt.Errorf("SearchServer fatal error"), "check error if server fatal errors with 5xx"},
		{ordinalReq, jsonErrorServerFabric("TOKEN :)"), fmt.Errorf("unknown bad request error: TOKEN :)"), "check unknown bad request error"},
		{ordinalReq, jsonErrorServerFabric("ErrorBadOrderField"), fmt.Errorf("OrderFeld %s invalid", ordinalReq.OrderField), "check unknown bad request error"},
	}

	for _, test := range testErrorsTable {

		testServer := test.server

		testClient := SearchClient{
			URL:         testServer.URL,
			AccessToken: "TOKEN :)",
		}

		_, err := testClient.FindUsers(test.req)

		if err == nil || err.Error() != test.expectedError.Error() {
			t.Errorf("Test: %s - unexpected err, Got %v, Want %v", test.desc, err, test.expectedError)
		}

	}

	t.Run("test unknown error handling", func(t *testing.T) {
		testingRegex := regexp.MustCompile(`^unknown error .*`)
		req := ordinalReq
		unknownErrSrv := httptest.NewServer(http.HandlerFunc(TimeoutServer))

		var errGot error

		testClient := SearchClient{
			URL:         unknownErrSrv.URL,
			AccessToken: "TOKEN :)",
		}

		errCh := make(chan error)
		go func(e chan error) {
			_, err := testClient.FindUsers(req)
			e <- err
		}(errCh)

		time.Sleep(100 * time.Millisecond)
		unknownErrSrv.CloseClientConnections()

		errGot = <-errCh

		if !testingRegex.MatchString(errGot.Error()) {
			t.Errorf("Test: %s - unexpected err, Got %v, Want %v", t.Name(), errGot, testingRegex.String())
		}

	})

	t.Run("test timeout error", func(t *testing.T) {

		testClient := SearchClient{
			URL:         timeoutSrv.URL,
			AccessToken: "TOKEN :)",
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

	t.Run("test json unpacking error", func(t *testing.T) {
		req := ordinalReq
		server := responseServerFabric(http.StatusBadRequest)

		regexExpected := regexp.MustCompile("^cant unpack error json: .*")

		testClient := SearchClient{
			URL:         server.URL,
			AccessToken: "TOKEN :)",
		}

		_, err := testClient.FindUsers(req)

		if !regexExpected.MatchString(err.Error()) {
			t.Errorf("Test: %s - unexpected err, Got %v, Want %v", t.Name(), err, regexExpected.String())
		}

	})

}

func TestSpyRequest(t *testing.T) {

	spy := SpyRequestServer{}
	spyReqSrv := httptest.NewServer(http.HandlerFunc(spy.SpyRequest))

	t.Run("check if limit > 25 then request limit = 26", func(t *testing.T) {
		testClient := SearchClient{
			URL:         spyReqSrv.URL,
			AccessToken: "TOKEN :)",
		}

		limitMoreThan25Req := SearchRequest{255, 1, "male", "gender", OrderByAsc}

		testClient.FindUsers(limitMoreThan25Req)

		got := spy.request.Get("limit")
		expected := "26"
		if got != expected {
			t.Errorf("Got limit=%s, want limit=%s", got, expected)
		}
	})
}

func TestSuccessfulPath(t *testing.T) {

	stubData := []User{
		{22, "Foo", 33, "I am Foo", "male"},
		{33, "Bar", 33, "I am Bar", "female"},
		{22, "Foo", 33, "I am Foo", "male"},
		{33, "Bar", 33, "I am Bar", "female"},
		{22, "Foo", 33, "I am Foo", "male"},
		{33, "Bar", 33, "I am Bar", "female"},
		{22, "Foo", 33, "I am Foo", "male"},
		{33, "Bar", 33, "I am Bar", "female"},
		{22, "Foo", 33, "I am Foo", "male"},
		{33, "Bar", 33, "I am Bar", "female"},
	}

	reqWithoutNextPage := SearchRequest{10, 1, "male", "gender", OrderByAsc}
	reqWithNextPage := SearchRequest{5, 1, "male", "gender", OrderByAsc}

	server1 := successfulServerFabric(stubData)
	server2 := successfulServerFabric(stubData[0:6])

	client1 := SearchClient{
		URL:         server1.URL,
		AccessToken: "TOKEN :)",
	}

	client2 := SearchClient{
		URL:         server2.URL,
		AccessToken: "TOKEN :)",
	}

	t.Run("check success when limit >= data available", func(t *testing.T) {
		resp, err := client1.FindUsers(reqWithoutNextPage)

		if err != nil {
			t.Errorf("Error expected to be nil, got %s", err)
		}

		expectedUsers := 10
		expectedNextPage := false

		if len(resp.Users) != expectedUsers {
			t.Errorf("Unexpected amount of users, want %d, got %d", expectedUsers, len(resp.Users))
		}

		if resp.NextPage != expectedNextPage {
			t.Errorf("Expected resp.NextPage=%v, got %v", expectedNextPage, resp.NextPage)
		}
	})

	t.Run("check success when limit < data available (next page required)", func(t *testing.T) {
		resp, err := client2.FindUsers(reqWithNextPage)

		if err != nil {
			t.Errorf("Error expected to be nil, got %s", err)
		}

		expectedUsers := 5
		expectedNextPage := true

		if len(resp.Users) != expectedUsers {
			t.Errorf("Unexpected amount of users, want %d, got %d", expectedUsers, len(resp.Users))
		}

		if resp.NextPage != expectedNextPage {
			t.Errorf("Expected resp.NextPage=%v, got %v", expectedNextPage, resp.NextPage)
		}
	})

}
