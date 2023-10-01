package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
)

// Commonly used strings.
const (
	APIOK         = "ok"
	APIERR        = "error"
	APIForbidden  = "Forbidden"
	APINoEndpoint = "No endpoint found"
)

// Main response structure.
type APIGeneralResp struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// Typical API responses are done with JSON. To make it easier to respond, this function will marshal/send json to a response writer.
func (s *HTTPServer) JSONResponse(w http.ResponseWriter, resp interface{}) {
	// Encode response as json.
	js, err := json.Marshal(resp)
	if err != nil {
		// Error should not happen normally...
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// If no error, we can set content type header and send response.
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	w.Write([]byte{'\n'})
}

// There are quite a few request that send a general response on error. This function is to make it easy to build/send a general response.
func (s *HTTPServer) APISendGeneralResp(w http.ResponseWriter, status, err string) {
	resp := APIGeneralResp{}
	resp.Status = status
	resp.Error = err
	s.JSONResponse(w, resp)
}

// Verifies that the client connectiong is authenticated.
func (s *HTTPServer) APIAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")

		if s.config.APIKey != "" && s.config.APIKey != apiKey {
			s.APISendGeneralResp(w, APIERR, APIForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Setup HTTP router with routes for the API calls.
func (s *HTTPServer) RegisterAPIRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	// Requires authentication.
	api.Use(s.APIAuthenticationMiddleware)

	// Just a test call.
	api.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		s.APISendGeneralResp(w, APIOK, "")
	})

	// Send message to slack channel for the current service.
	// Defaults to admin if no service currently occuring.
	api.HandleFunc("/send_message", func(w http.ResponseWriter, r *http.Request) {
		// Get message, either from URL query or multi part form.
		var message string
		err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
		if err == nil {
			message = r.Form.Get("message")
		}
		if message == "" {
			message = r.URL.Query().Get("message")
		}

		// If no message provided, fail.
		if message == "" {
			log.Println("No message provided")
			s.APISendGeneralResp(w, APIERR, "No message provided")
			return
		}

		// Get current time and default conversation.
		now := time.Now().UTC()
		conversation := app.config.Slack.DefaultConversation

		// Find plan times that are occuring right now.
		var planTime PlanTimes
		app.db.Where("time_type='service' AND starts_at < ? AND ends_at > ?", now, now).First(&planTime)
		if planTime.Plan != 0 {
			// If plan found, check for the slack channel.
			var channel SlackChannels
			app.db.Where("pc_plan = ?", planTime.Plan).First(&channel)
			if channel.ID != "" {
				// If slack channel found, update the conversation to the channel ID.
				conversation = channel.ID
			}
		}

		// If no conversation found, likely will happen if no admin is configured, return error.
		if conversation == "" {
			log.Println("No conversation found")
			s.APISendGeneralResp(w, APIERR, "No conversation found")
			return
		}

		// Send message to Slack.
		_, _, err = app.slack.PostMessage(conversation, slack.MsgOptionText(message, false))
		if err != nil {
			log.Println("Error sending message:", err)
			s.APISendGeneralResp(w, APIERR, "Error sending message")
			return
		}

		// Return a success.
		s.APISendGeneralResp(w, APIOK, "")
	}).Methods(http.MethodPost)

	// If nothing else, we return a not found response.
	api.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.APISendGeneralResp(w, APIERR, APINoEndpoint)
	})
}
