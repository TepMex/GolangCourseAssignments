package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%v", r)
}

// код писать тут
func TestServer(t *testing.T) {
	testReq := SearchRequest{
		Limit:      10,
		Offset:     1,
		Query:      "male",
		OrderField: "gender",
		OrderBy:    OrderByAsc,
	}

	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))

	testClient := SearchClient{
		URL:         testServer.URL,
		AccessToken: "ZHOPA",
	}

	testClient.FindUsers(testReq)
}
