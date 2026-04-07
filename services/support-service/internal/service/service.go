package service

import (
	"context"
	"fmt"

	"github.com/vishalss1/CartGO/services/support-service/internal/model"
	"github.com/vishalss1/CartGO/services/support-service/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTicket(ctx context.Context, key string, customerID string, subject string) (*model.Ticket, error) {
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}

	ticket := &model.Ticket{
		CustomerID: customerID,
		Subject:    subject,
		Status:     model.StatusOpen,
		Priority:   model.PriorityMedium,
	}

	return s.repo.CreateTicketWithIdempotency(ctx, key, ticket)
}

func (s *Service) ListTickets(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Ticket, error) {
	return s.repo.ListTickets(ctx, filters, limit, offset)
}

func (s *Service) GetTicket(ctx context.Context, id string) (*model.Ticket, error) {
	return s.repo.GetTicket(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, ticketID string, newStatus model.TicketStatus, performedBy string, role string) error {
	if role != "SUPPORT_AGENT" && role != "ADMIN" {
		return fmt.Errorf("unauthorized: only agents or admins can update status")
	}

	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	// Strict State Machine Transitions
	valid := false
	switch ticket.Status {
	case model.StatusOpen:
		if newStatus == model.StatusInProgress {
			valid = true
		}
	case model.StatusInProgress:
		if newStatus == model.StatusResolved {
			valid = true
		}
	case model.StatusResolved:
		if newStatus == model.StatusClosed || newStatus == model.StatusReopened {
			valid = true
		}
	case model.StatusReopened:
		if newStatus == model.StatusInProgress {
			valid = true
		}
	}

	if !valid {
		return fmt.Errorf("invalid state transition from %s to %s", ticket.Status, newStatus)
	}

	ticket.Status = newStatus
	return s.repo.UpdateTicket(ctx, ticket, model.ActionStatusUpdated, performedBy)
}

func (s *Service) AssignTicket(ctx context.Context, ticketID string, agentID string, performedBy string, role string) error {
	if role != "SUPPORT_AGENT" && role != "ADMIN" {
		return fmt.Errorf("unauthorized: only agents or admins can assign tickets")
	}

	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == model.StatusClosed {
		return fmt.Errorf("cannot assign a closed ticket")
	}

	ticket.AssignedAgentID = &agentID
	return s.repo.UpdateTicket(ctx, ticket, model.ActionAssigned, performedBy)
}

func (s *Service) AddMessage(ctx context.Context, ticketID string, senderID string, role string, content string) error {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.Status == model.StatusClosed {
		return fmt.Errorf("cannot add messages to a closed ticket")
	}

	// RBAC: Only owner, assigned agent, or admin
	authorized := false
	if senderID == ticket.CustomerID {
		authorized = true
	} else if ticket.AssignedAgentID != nil && senderID == *ticket.AssignedAgentID {
		authorized = true
	} else if role == "ADMIN" {
		authorized = true
	}

	if !authorized {
		return fmt.Errorf("unauthorized to message this ticket")
	}

	msg := &model.Message{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderRole: role,
		Content:    content,
	}

	return s.repo.AddMessage(ctx, msg)
}

func (s *Service) ListMessages(ctx context.Context, ticketID string, customerID string, role string, limit, offset int) ([]*model.Message, error) {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Access control
	if role != "ADMIN" && role != "SUPPORT_AGENT" && ticket.CustomerID != customerID {
		return nil, fmt.Errorf("unauthorized access to ticket messages")
	}

	return s.repo.ListMessages(ctx, ticketID, limit, offset)
}
