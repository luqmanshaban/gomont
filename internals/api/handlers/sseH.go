package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/api/utils"
	"github.com/luqmanshaban/gomont/internals/sse"
	"github.com/luqmanshaban/gomont/internals/store"
)

// EventsHandler streams monitor status changes to the dashboard over a
// single Server-Sent Events connection. The browser's EventSource API
// can't send Authorization headers, so the JWT is passed as a query
// parameter instead (?token=...) and verified with the same
// utils.ValidateToken used by AuthMiddleware for normal requests.
type EventsHandler struct {
	Broker    *sse.Broker
	URLStore  *store.URLStore
	JWTSecret string
}

// Stream handles GET /events?token=<jwt>. It resolves the authenticated
// user's monitor IDs, subscribes to updates for all of them, and writes
// each subsequent event to the response as it arrives — until the client
// disconnects (browser tab closed, navigation away, etc.), detected via
// r.Context().Done().
func (h *EventsHandler) Stream(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := utils.ValidateToken(h.JWTSecret, token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}
	userID := claims.ID

	urls, err := h.URLStore.GetURLsByUserID(userID)
	if err != nil {
		slog.Error("sse: failed to load monitors for subscription", "user_id", userID, "error", err)
		http.Error(w, "failed to load monitors", http.StatusInternalServerError)
		return
	}

	monitorIDs := make([]int, 0, len(urls))
	for _, u := range urls {
		monitorIDs = append(monitorIDs, u.ID)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		// Should not happen with Go's standard net/http server, but SSE is
		// meaningless without the ability to flush partial writes to the
		// client immediately.
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := h.Broker.Subscribe(monitorIDs)
	defer unsubscribe()

	// Send an initial comment immediately. Some browsers/proxies don't
	// consider the connection "open" (and won't fire EventSource's onopen)
	// until the first byte arrives, so this avoids a perceived hang on
	// slow networks.
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected (tab closed, navigated away, lost
			// connection). Nothing further to send; defer above cleans up
			// the subscription.
			return

		case event, open := <-events:
			if !open {
				return
			}
			data, err := event.MarshalData()
			if err != nil {
				slog.Error("sse: failed to marshal event data", "error", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, data)
			flusher.Flush()
		}
	}
}