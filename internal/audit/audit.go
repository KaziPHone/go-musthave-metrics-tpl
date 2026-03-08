package audit

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

// AuditEvent — структура события аудита
type AuditEvent struct {
	Ts        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Observer — интерфейс получателя событий аудита
type Observer interface {
	OnAuditEvent(event AuditEvent)
}

// Subject — управление подписчиками
type Subject struct {
	observers []Observer
	mu        sync.Mutex
}

func (s *Subject) Attach(obs Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, obs)
}

func (s *Subject) Notify(event AuditEvent) {
	s.mu.Lock()
	observers := make([]Observer, len(s.observers))
	copy(observers, s.observers)
	s.mu.Unlock()

	for _, o := range observers {
		o.OnAuditEvent(event)
	}
}

// FileObserver — пишет события в файл
type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{filePath: filePath}
}

func (f *FileObserver) OnAuditEvent(event AuditEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, _ := json.Marshal(event)
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open audit file: %v", err)
		return
	}
	defer file.Close()

	_, _ = file.WriteString(string(data) + "\n")
}

// HTTPObserver — отправляет события по HTTP
type HTTPObserver struct {
	url string
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{url: url}
}

func (h *HTTPObserver) OnAuditEvent(event AuditEvent) {
	data, _ := json.Marshal(event)
	resp, err := http.Post(h.url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to send audit event to %s: %v", h.url, err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Audit server responded with status: %d", resp.StatusCode)
	}
}
