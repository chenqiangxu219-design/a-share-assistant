package chat

import (
	"sync"
	"time"
)

// Session holds conversation history for a single session
type Session struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
	Created  time.Time `json:"created"`
}

// ConversationStore manages conversation history across sessions
type ConversationStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	maxMsgs  int // max messages per session
}

// NewConversationStore creates a new conversation store
func NewConversationStore() *ConversationStore {
	return &ConversationStore{
		sessions: make(map[string]*Session),
		maxMsgs:  20, // keep last 20 messages per session
	}
}

// AddMessage adds a message to a session
func (s *ConversationStore) AddMessage(sessionID string, msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[sessionID]; ok {
		session.Messages = append(session.Messages, msg)
		// Trim to max messages
		if len(session.Messages) > s.maxMsgs {
			session.Messages = session.Messages[len(session.Messages)-s.maxMsgs:]
		}
	} else {
		s.sessions[sessionID] = &Session{
			ID:       sessionID,
			Messages: []Message{msg},
			Created:  time.Now(),
		}
	}
}

// GetHistory returns recent messages for a session
func (s *ConversationStore) GetHistory(sessionID string, max int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok || len(session.Messages) == 0 {
		return nil
	}

	count := max
	if count > len(session.Messages) {
		count = len(session.Messages)
	}

	// Return the last `max` messages
	start := len(session.Messages) - count
	result := make([]Message, count)
	copy(result, session.Messages[start:])
	return result
}

// DeleteSession removes a session
func (s *ConversationStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// Cleanup removes sessions older than maxAge
func (s *ConversationStore) Cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, session := range s.sessions {
		if session.Created.Before(cutoff) {
			delete(s.sessions, id)
		}
	}
}
