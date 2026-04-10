package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/support-service/internal/model"
)

// Transactional Ticket Creation with Idempotency
func (r *Repository) CreateTicketWithIdempotency(ctx context.Context, key string, ticket *model.Ticket) (*model.Ticket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Attempt to insert idempotency key
	var existingTicketID sql.NullString
	err = tx.QueryRowContext(ctx, 
		"INSERT INTO idempotency_keys (key, customer_id) VALUES ($1, $2) ON CONFLICT (key, customer_id) DO UPDATE SET key = EXCLUDED.key RETURNING ticket_id", 
		key, ticket.CustomerID).Scan(&existingTicketID)
	
	if err == nil && existingTicketID.Valid {
		// Ticket already exists for this key
		parsedID, _ := uuid.Parse(existingTicketID.String)
		return r.getTicketByID(ctx, tx, parsedID)
	}

	// 2. Create the ticket
	ticket.ID = util.GenerateUUID()
	err = tx.QueryRowContext(ctx,
		"INSERT INTO tickets (id, customer_id, order_id, subject, status, priority, version) VALUES ($1, $2, $3, $4, $5, $6, 1) RETURNING created_at, updated_at",
		ticket.ID, ticket.CustomerID, ticket.OrderID, ticket.Subject, ticket.Status, ticket.Priority).Scan(&ticket.CreatedAt, &ticket.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 3. Update idempotency key with ticket ID
	_, err = tx.ExecContext(ctx, "UPDATE idempotency_keys SET ticket_id = $1 WHERE key = $2 AND customer_id = $3", ticket.ID, key, ticket.CustomerID)
	if err != nil {
		return nil, err
	}

	// 4. Audit Log
	auditID := util.GenerateUUID()
	_, err = tx.ExecContext(ctx, "INSERT INTO audit_logs (id, ticket_id, action, performed_by) VALUES ($1, $2, $3, $4)", auditID, ticket.ID, model.ActionTicketCreated, ticket.CustomerID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (r *Repository) getTicketByID(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*model.Ticket, error) {
	t := &model.Ticket{}
	query := "SELECT id, customer_id, order_id, assigned_agent_id, status, priority, subject, version, created_at, updated_at FROM tickets WHERE id = $1"
	
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.CustomerID, &t.OrderID, &t.AssignedAgentID, &t.Status, &t.Priority, &t.Subject, &t.Version, &t.CreatedAt, &t.UpdatedAt)
	} else {
		err = r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.CustomerID, &t.OrderID, &t.AssignedAgentID, &t.Status, &t.Priority, &t.Subject, &t.Version, &t.CreatedAt, &t.UpdatedAt)
	}
	
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) GetTicket(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	return r.getTicketByID(ctx, nil, id)
}

func (r *Repository) ListTickets(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Ticket, error) {
	query := "SELECT id, customer_id, order_id, assigned_agent_id, status, priority, subject, version, created_at, updated_at FROM tickets WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	for k, v := range filters {
		query += fmt.Sprintf(" AND %s = $%d", k, argCount)
		args = append(args, v)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []*model.Ticket{}
	for rows.Next() {
		t := &model.Ticket{}
		if err := rows.Scan(&t.ID, &t.CustomerID, &t.OrderID, &t.AssignedAgentID, &t.Status, &t.Priority, &t.Subject, &t.Version, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

func (r *Repository) UpdateTicket(ctx context.Context, ticket *model.Ticket, action model.AuditAction, performedBy uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Optimistic locking: WHERE version = $version
	res, err := tx.ExecContext(ctx,
		"UPDATE tickets SET status = $1, priority = $2, assigned_agent_id = $3, version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $4 AND version = $5",
		ticket.Status, ticket.Priority, ticket.AssignedAgentID, ticket.ID, ticket.Version)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("concurrency conflict or ticket not found")
	}

	// Audit Log
	auditID := util.GenerateUUID()
	_, err = tx.ExecContext(ctx, "INSERT INTO audit_logs (id, ticket_id, action, performed_by) VALUES ($1, $2, $3, $4)", auditID, ticket.ID, action, performedBy)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) AddMessage(ctx context.Context, msg *model.Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	msg.ID = util.GenerateUUID()
	err = tx.QueryRowContext(ctx,
		"INSERT INTO messages (id, ticket_id, sender_id, sender_role, content) VALUES ($1, $2, $3, $4, $5) RETURNING created_at",
		msg.ID, msg.TicketID, msg.SenderID, msg.SenderRole, msg.Content).Scan(&msg.CreatedAt)
	if err != nil {
		return err
	}

	// Audit Log
	auditID := util.GenerateUUID()
	_, err = tx.ExecContext(ctx, "INSERT INTO audit_logs (id, ticket_id, action, performed_by) VALUES ($1, $2, $3, $4)", auditID, msg.TicketID, model.ActionMessageAdded, msg.SenderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) ListMessages(ctx context.Context, ticketID uuid.UUID, limit, offset int) ([]*model.Message, error) {
	query := "SELECT id, ticket_id, sender_id, sender_role, content, created_at FROM messages WHERE ticket_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3"
	rows, err := r.db.QueryContext(ctx, query, ticketID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*model.Message{}
	for rows.Next() {
		m := &model.Message{}
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.SenderRole, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
