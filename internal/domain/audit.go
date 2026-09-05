package domain

import "github.com/google/uuid"

const (
	AuditEntityTransaction = "transaction"

	AuditActionCreated = "CREATED"
)

type Actor struct {
	UserID    uuid.UUID
	Email     string
	IPAddress string
	UserAgent string
}

type AuditEntry struct {
	EntityType string
	EntityID   uuid.UUID
	Action     string
	OldValue   any
	NewValue   any
	Actor      Actor
}

func NewAuditEntry(actor *Actor, entityType, action string, entityID uuid.UUID, oldValue, newValue any) *AuditEntry {
	if actor == nil {
		return nil
	}

	return &AuditEntry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		OldValue:   oldValue,
		NewValue:   newValue,
		Actor:      *actor,
	}
}
