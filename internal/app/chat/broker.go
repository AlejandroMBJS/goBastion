package chat

import (
	"context"
	"sync"
	"time"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	RoomID    string    `json:"room_id"`
}

// Client represents a connected chat client
type Client struct {
	ID       string
	Messages chan Message
	RoomID   string
}

// Broker manages real-time message broadcasting using channels and goroutines
// This demonstrates Go's concurrency patterns: channels, goroutines, and mutexes
type Broker struct {
	// Clients holds all connected clients
	clients map[string]*Client
	// Mutex for thread-safe client map operations
	mu sync.RWMutex

	// Channels for concurrent operations
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Message history per room (limited to last 100 messages)
	history map[string][]Message
	histMu  sync.RWMutex
}

// NewBroker creates a new message broker with initialized channels
func NewBroker(ctx context.Context) *Broker {
	ctx, cancel := context.WithCancel(ctx)

	return &Broker{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 100), // Buffered channel for performance
		ctx:        ctx,
		cancel:     cancel,
		history:    make(map[string][]Message),
	}
}

// Start runs the broker's main event loop
// This is a long-running goroutine that handles all concurrent operations
func (b *Broker) Start() {
	go func() {
		for {
			select {
			case client := <-b.register:
				b.registerClient(client)

			case client := <-b.unregister:
				b.unregisterClient(client)

			case message := <-b.broadcast:
				b.broadcastMessage(message)

			case <-b.ctx.Done():
				b.shutdown()
				return
			}
		}
	}()
}

// registerClient adds a new client to the broker
func (b *Broker) registerClient(client *Client) {
	b.mu.Lock()
	b.clients[client.ID] = client
	b.mu.Unlock()

	// Send message history to new client
	b.histMu.RLock()
	history := b.history[client.RoomID]
	b.histMu.RUnlock()

	// Send last 20 messages to new client
	start := len(history) - 20
	if start < 0 {
		start = 0
	}

	for _, msg := range history[start:] {
		select {
		case client.Messages <- msg:
		case <-time.After(100 * time.Millisecond):
			// Client not responding, skip
		}
	}
}

// unregisterClient removes a client and closes its channel
func (b *Broker) unregisterClient(client *Client) {
	b.mu.Lock()
	if _, ok := b.clients[client.ID]; ok {
		delete(b.clients, client.ID)
		close(client.Messages)
	}
	b.mu.Unlock()
}

// broadcastMessage sends a message to all clients in the same room
func (b *Broker) broadcastMessage(message Message) {
	// Save to history
	b.histMu.Lock()
	if b.history[message.RoomID] == nil {
		b.history[message.RoomID] = make([]Message, 0, 100)
	}
	b.history[message.RoomID] = append(b.history[message.RoomID], message)

	// Keep only last 100 messages
	if len(b.history[message.RoomID]) > 100 {
		b.history[message.RoomID] = b.history[message.RoomID][1:]
	}
	b.histMu.Unlock()

	// Broadcast to all clients in the room
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, client := range b.clients {
		if client.RoomID == message.RoomID {
			select {
			case client.Messages <- message:
				// Message sent successfully
			case <-time.After(100 * time.Millisecond):
				// Client not responding, skip
			}
		}
	}
}

// Register adds a client to the broker
func (b *Broker) Register(client *Client) {
	b.register <- client
}

// Unregister removes a client from the broker
func (b *Broker) Unregister(client *Client) {
	b.unregister <- client
}

// Broadcast sends a message to all clients
func (b *Broker) Broadcast(message Message) {
	b.broadcast <- message
}

// GetHistory returns message history for a room
func (b *Broker) GetHistory(roomID string, limit int) []Message {
	b.histMu.RLock()
	defer b.histMu.RUnlock()

	history := b.history[roomID]
	if history == nil {
		return []Message{}
	}

	start := len(history) - limit
	if start < 0 {
		start = 0
	}

	return history[start:]
}

// shutdown gracefully shuts down the broker
func (b *Broker) shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, client := range b.clients {
		close(client.Messages)
	}
	b.clients = make(map[string]*Client)
}

// Stop gracefully stops the broker
func (b *Broker) Stop() {
	b.cancel()
}

// ActiveClients returns the number of active clients
func (b *Broker) ActiveClients() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ActiveRooms returns the number of active rooms
func (b *Broker) ActiveRooms() int {
	b.histMu.RLock()
	defer b.histMu.RUnlock()
	return len(b.history)
}
