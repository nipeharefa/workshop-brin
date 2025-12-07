package models

import (
	"time"

	"github.com/google/uuid"
)

// UserContext represents user information sent to N8N
type UserContext struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Phone  string    `json:"phone"`
	Email  string    `json:"email"`
}

// N8NRequest represents the payload sent to N8N workflow
type N8NRequest struct {
	UserContext *UserContext `json:"user_context"`
	Message     string       `json:"message"`
	MessageID   string       `json:"message_id"`
	Timestamp   time.Time    `json:"timestamp"`
}

// N8NResponse represents the response received from N8N workflow
type N8NResponse struct {
	MessageID string `json:"message_id"`
	Phone     string `json:"phone"`
	Response  string `json:"response"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
	Server    string    `json:"server"`
}

// Signal represents a stock trading signal received from N8N
type Signal struct {
	Ticker           string  `json:"ticker" binding:"required,min=1,max=10"`
	LastDate         string  `json:"last_date" binding:"required"`
	LastClose        int     `json:"last_close"`
	EntryPrice       int     `json:"entry_price" binding:"required"`
	EntryGapPercent  float64 `json:"entry_gap_percent" binding:"required"`
	Stop             float64 `json:"stop" binding:"required,gt=0"`
	Target           float64 `json:"target" binding:"required,gt=0"`
	RiskReward       float64 `json:"risk_reward" binding:"required,gt=0"`
	BacktestWinRate  float64 `json:"backtest_win_rate" binding:"required,gte=0,lte=100"`
	TotalTrades      int     `json:"total_trades" binding:"required,gte=0"`
	ConfluenceScore  float64 `json:"confluence_score" binding:"required,gte=0"`
	ConfluenceHits   string  `json:"confluence_hits"`
	OverallSentiment string  `json:"overall_sentiment"`
	ConfidenceScore  float64 `json:"confidence_score"`
	SentimentScore   float64 `json:"sentiment_score"`
	AnalysisSummary  string  `json:"analysis_summary"`
}

// SignalResponse represents the response when processing a signal
type SignalResponse struct {
	Ticker           string    `json:"ticker"`
	UsersNotified    int       `json:"users_notified"`
	Timestamp        time.Time `json:"timestamp"`
	ProcessingTimeMs int64     `json:"processing_time_ms"`
}

// WorkflowConfig represents global workflow routing configuration
type WorkflowConfig struct {
	ID           int       `json:"id" db:"id"`
	WorkflowType string    `json:"workflow_type" db:"workflow_type"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// FlowiseRequest represents the payload sent to Flowise prediction endpoint
type FlowiseRequest struct {
	Question       string                 `json:"question"`
	OverrideConfig *FlowiseOverrideConfig `json:"overrideConfig,omitempty"`
}

// FlowiseOverrideConfig represents runtime configuration for Flowise
type FlowiseOverrideConfig struct {
	SessionID string                 `json:"sessionId,omitempty"`
	Vars      map[string]interface{} `json:"vars,omitempty"`
}

// FlowiseResponse represents the response received from Flowise prediction
type FlowiseResponse struct {
	Text      string `json:"text"`
	ChatID    string `json:"chatId,omitempty"`
	MessageID string `json:"message_id"`
	Phone     string `json:"phone"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
