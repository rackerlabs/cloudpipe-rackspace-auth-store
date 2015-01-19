package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateHandlerSuccess(t *testing.T) {
	r := HTTPRequest(t, "GET", "https://localhost/v1/validate?accountName=someone&apiKey=ff01ab", "")
	w := httptest.NewRecorder()
	c := &Context{}

	ValidateHandler(c, w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected response code %d, but was %d", http.StatusNoContent, w.Code)
	}
}

func TestValidateHandlerReject(t *testing.T) {

	//Since we always accept right now, this will outright reject
	r := HTTPRequest(t, "GET", "https://localhost/v1/validate?accountName=someone&apiKey=ff01ab", "")
	w := httptest.NewRecorder()
	c := &Context{}

	ValidateHandler(c, w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected response code %d, but was %d", http.StatusNotFound, w.Code)
	}
}
