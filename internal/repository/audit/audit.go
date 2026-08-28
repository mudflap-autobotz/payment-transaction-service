package audit

import (
	"context"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuditDB struct {
	bun.BaseModel `bun:"table:audit_logs,alias:al"`

	ID              uuid.UUID  `bun:"id,pk,nullzero,default:uuidv7()"`
	EntityType      string     `bun:"entity_type"`
	EntityID        *uuid.UUID `bun:"entity_id"`
	Action          string     `bun:"action"`
	OldValue        any        `bun:"old_value,type:jsonb,nullzero"`
	NewValue        any        `bun:"new_value,type:jsonb,nullzero"`
	ChangedByUserID *uuid.UUID `bun:"changed_by_user_id"`
	ChangedByEmail  string     `bun:"changed_by_email,nullzero"`
	IPAddress       string     `bun:"ip_address,type:inet,nullzero"`
	UserAgent       string     `bun:"user_agent,nullzero"`
	CreatedAt       time.Time  `bun:"created_at,nullzero,default:now()"`
}

func Insert(ctx context.Context, db bun.IDB, entry *domain.AuditEntry) error {
	if entry == nil {
		return nil
	}

	auditDB := &AuditDB{
		EntityType:     entry.EntityType,
		Action:         entry.Action,
		OldValue:       entry.OldValue,
		NewValue:       entry.NewValue,
		ChangedByEmail: entry.Actor.Email,
		IPAddress:      entry.Actor.IPAddress,
		UserAgent:      entry.Actor.UserAgent,
	}

	if entry.EntityID != uuid.Nil {
		entityID := entry.EntityID
		auditDB.EntityID = &entityID
	}

	if entry.Actor.UserID != uuid.Nil {
		actorID := entry.Actor.UserID
		auditDB.ChangedByUserID = &actorID
	}

	_, err := db.NewInsert().Model(auditDB).Exec(ctx)

	return err
}
