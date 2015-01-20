package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace"
)

// ValidateHandler determines whether or not an API key is valid for a specific account.
func ValidateHandler(c *Context, w http.ResponseWriter, r *http.Request) {
	if !MethodOk(w, r, "GET") {
		return
	}

	if err := r.ParseForm(); err != nil {
		APIError{
			Message: fmt.Sprintf("Unable to parse URL parameters: %v", err),
		}.Log("").Report(w, http.StatusBadRequest)
		return
	}

	accountName, apiKey := r.FormValue("accountName"), r.FormValue("apiKey")
	if accountName == "" || apiKey == "" {
		APIError{
			UserMessage: `Missing required query parameters "accountName" and "apiKey".`,
			LogMessage:  "Key validation request missing required query parameters.",
		}.Log("").Report(w, http.StatusBadRequest)
		return
	}

	var message string

	// Is the key cached?
	if c.KeyStore.IsIn(accountName, apiKey) {
		w.WriteHeader(http.StatusNoContent)
		message = "Cached API key successfully validated."

		log.WithFields(log.Fields{
			"account": accountName,
		}).Info(message)
		return

	}

	// Key is not cached, do the full call to Rackspace Identity
	ao := gophercloud.AuthOptions{
		Username: accountName,
		APIKey:   apiKey,
	}

	// We only care to know if there was an error or not
	_, err := rackspace.AuthenticatedClient(ao)

	if err == nil {
		w.WriteHeader(http.StatusNoContent)

		// API key was valid, so we can cache it
		c.KeyStore.Add(accountName, apiKey)

		message = "API key successfully validated and cached."

		log.WithFields(log.Fields{
			"account": accountName,
		}).Info(message)
		return

	}

	w.WriteHeader(http.StatusNotFound)
	message = "Rackspace Identity rejected"

	log.WithFields(log.Fields{
		"account": accountName,
		"err":     err,
	}).Info("Rackspace Identity Error.")

}
