package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AlejandroMBJS/goBastion/internal/app/chat"
	frameworkrouter "github.com/AlejandroMBJS/goBastion/internal/framework/router"
	"github.com/AlejandroMBJS/goBastion/internal/framework/view"

	"github.com/google/uuid"
)

var (
	// Global message broker (in production, use dependency injection)
	messageBroker *chat.Broker
)

// InitChatBroker initializes the global message broker
func InitChatBroker(ctx context.Context) {
	messageBroker = chat.NewBroker(ctx)
	messageBroker.Start()
}

// RegisterChatRoutes registers all chat-related routes
// Demonstrates: SSE, HTMX integration, concurrent message handling
func RegisterChatRoutes(r *frameworkrouter.Router, views *view.Engine) {
	// Chat room page (renders the UI)
	r.GET("/chat", handleChatRoom(views))
	r.GET("/chat/{room}", handleChatRoom(views))

	// Server-Sent Events endpoint for real-time messages
	r.GET("/chat/{room}/stream", handleChatStream)

	// HTMX endpoint to send messages
	r.POST("/chat/{room}/send", handleSendMessage(views))

	// HTMX endpoint to get message history
	r.GET("/chat/{room}/history", handleChatHistory(views))

	// API endpoint for stats (demonstrates concurrent operations)
	r.GET("/api/chat/stats", handleChatStats)
}

// handleChatRoom renders the chat room interface
func handleChatRoom(views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		roomID := params["room"]
		if roomID == "" {
			roomID = "general"
		}

		// Get current user from context (set by JWT middleware)
		username := "Anonymous"
		if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
			if name, ok := claims["name"].(string); ok {
				username = name
			}
		}

		data := map[string]interface{}{
			"Title":    "Chat Room - " + roomID,
			"RoomID":   roomID,
			"Username": username,
			"UserID":   uuid.New().String()[:8],
		}

		views.Render(w, "chat/room", data)
	}
}

// handleChatStream implements Server-Sent Events for real-time messages
// Demonstrates: Long-lived connections, goroutines, channels, context cancellation
func handleChatStream(w http.ResponseWriter, r *http.Request, params map[string]string) {
	roomID := params["room"]
	if roomID == "" {
		roomID = "general"
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create a client
	client := &chat.Client{
		ID:       uuid.New().String(),
		Messages: make(chan chat.Message, 10),
		RoomID:   roomID,
	}

	// Register client with broker
	messageBroker.Register(client)
	defer messageBroker.Unregister(client)

	// Create a context that cancels when client disconnects
	ctx := r.Context()
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"message\":\"Connected to room %s\"}\n\n", roomID)
	flusher.Flush()

	// Event loop: send messages as they arrive
	for {
		select {
		case msg, ok := <-client.Messages:
			if !ok {
				// Channel closed, client disconnected
				return
			}

			// Convert message to JSON
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			// Send SSE message
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case <-ctx.Done():
			// Client disconnected
			return

		case <-time.After(30 * time.Second):
			// Send keepalive ping
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// handleSendMessage handles sending a new message (HTMX endpoint)
// Returns HTML fragment that HTMX will insert into the DOM
func handleSendMessage(views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		roomID := params["room"]
		if roomID == "" {
			roomID = "general"
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		content := r.FormValue("message")
		username := r.FormValue("username")

		if content == "" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create message
		msg := chat.Message{
			ID:        uuid.New().String(),
			User:      username,
			Content:   content,
			Timestamp: time.Now(),
			RoomID:    roomID,
		}

		// Broadcast message (non-blocking, handled by broker goroutine)
		messageBroker.Broadcast(msg)

		// Return empty response (HTMX will clear the form)
		w.WriteHeader(http.StatusOK)
	}
}

// handleChatHistory returns message history as HTML fragments (for HTMX)
func handleChatHistory(views *view.Engine) frameworkrouter.Handler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		roomID := params["room"]
		if roomID == "" {
			roomID = "general"
		}

		// Get message history
		messages := messageBroker.GetHistory(roomID, 50)

		data := map[string]interface{}{
			"Messages": messages,
		}

		// Render partial template
		views.Render(w, "chat/messages", data)
	}
}

// handleChatStats returns real-time statistics
// Demonstrates: Concurrent operations, JSON API
func handleChatStats(w http.ResponseWriter, r *http.Request, params map[string]string) {
	stats := map[string]interface{}{
		"active_clients": messageBroker.ActiveClients(),
		"active_rooms":   messageBroker.ActiveRooms(),
		"timestamp":      time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Worker Pool Example: Background job processing
// Demonstrates: Worker pool pattern, fan-out/fan-in, WaitGroups

type Job struct {
	ID   int
	Data string
}

type Result struct {
	Job    Job
	Output string
	Error  error
}

// ProcessJobsWithWorkerPool demonstrates the worker pool pattern
// This would be called from a background service or cron job
func ProcessJobsWithWorkerPool(jobs []Job, numWorkers int) []Result {
	// Create channels
	jobChan := make(chan Job, len(jobs))
	resultChan := make(chan Result, len(jobs))

	// Start workers
	ctx := context.Background()
	for i := 0; i < numWorkers; i++ {
		go worker(ctx, i, jobChan, resultChan)
	}

	// Send jobs to workers
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	// Collect results
	results := make([]Result, 0, len(jobs))
	for i := 0; i < len(jobs); i++ {
		result := <-resultChan
		results = append(results, result)
	}

	return results
}

// worker processes jobs from the job channel
func worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			// Simulate processing
			time.Sleep(100 * time.Millisecond)

			result := Result{
				Job:    job,
				Output: fmt.Sprintf("Worker %d processed job %d: %s", id, job.ID, job.Data),
				Error:  nil,
			}

			results <- result
		}
	}
}

// Rate Limiter Example using Token Bucket
// Demonstrates: Concurrency patterns for rate limiting

type RateLimiter struct {
	tokens    chan struct{}
	maxTokens int
	refillInterval time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewRateLimiter(maxTokens int, refillInterval time.Duration) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		tokens:         make(chan struct{}, maxTokens),
		maxTokens:      maxTokens,
		refillInterval: refillInterval,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Fill initial tokens
	for i := 0; i < maxTokens; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refill goroutine
	go rl.refill()

	return rl
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.refillInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case rl.tokens <- struct{}{}:
				// Token added
			default:
				// Bucket full
			}
		case <-rl.ctx.Done():
			return
		}
	}
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *RateLimiter) Stop() {
	rl.cancel()
}
