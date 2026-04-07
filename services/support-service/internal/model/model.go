package model

import (
	"time"
)

type TicketStatus string

const (
	StatusOpen       TicketStatus = "OPEN"
	StatusInProgress TicketStatus = "IN_PROGRESS"
	StatusResolved   TicketStatus = "RESOLVED"
	StatusClosed     TicketStatus = "CLOSED"
	StatusReopened   TicketStatus = "REOPENED"
)

type TicketPriority string

const (
	PriorityLow    TicketPriority = "LOW"
	PriorityMedium TicketPriority = "MEDIUM"
	PriorityHigh   TicketPriority = "HIGH"
	PriorityUrgent TicketPriority = "URGENT"
)

type AuditAction string

const (
	ActionTicketCreated AuditAction = "TICKET_CREATED"
	ActionStatusUpdated AuditAction = "STATUS_UPDATED"
	ActionAssigned      AuditAction = "ASSIGNED"
	ActionMessageAdded  AuditAction = "MESSAGE_ADDED"
)

type Ticket struct {
	ID              string         `json:"id"`
	CustomerID      string         `json:"customer_id"`
	AssignedAgentID *string        `json:"assigned_agent_id"`
	Status          TicketStatus   `json:"status"`
	Priority        TicketPriority `json:"priority"`
	Subject         string         `json:"subject"`
	Version         int            `json:"version"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Message struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticket_id"`
	SenderID   string    `json:"sender_id"`
	SenderRole string    `json:"sender_role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type IdempotencyKey struct {
	Key        string    `json:"key"`
	CustomerID string    `json:"customer_id"`
	TicketID   *string   `json:"ticket_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLog struct {
	ID          string      `json:"id"`
	TicketID    string      `json:"ticket_id"`
	Action      AuditAction `json:"action"`
	PerformedBy string      `json:"performed_by"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
