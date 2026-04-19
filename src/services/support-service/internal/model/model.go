package model

import (
	"time"

	"github.com/google/uuid"
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
	ID              uuid.UUID      `json:"id"`
	CustomerID      uuid.UUID      `json:"customer_id"`
	OrderID         *uuid.UUID     `json:"order_id,omitempty"`
	AssignedAgentID *uuid.UUID     `json:"assigned_agent_id,omitempty"`
	Status          TicketStatus   `json:"status"`
	Priority        TicketPriority `json:"priority"`
	Subject         string         `json:"subject"`
	Version         int            `json:"version"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Message struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	SenderID   uuid.UUID `json:"sender_id"`
	SenderRole string    `json:"sender_role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type IdempotencyKey struct {
	Key        string    `json:"key"`
	CustomerID uuid.UUID `json:"customer_id"`
	TicketID   *uuid.UUID `json:"ticket_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLog struct {
	ID          uuid.UUID   `json:"id"`
	TicketID    uuid.UUID   `json:"ticket_id"`
	Action      AuditAction `json:"action"`
	PerformedBy uuid.UUID   `json:"performed_by"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
