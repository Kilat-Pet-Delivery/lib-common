package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity provides the foundation for all domain entities.
type BaseEntity struct {
	ID        uuid.UUID `json:"id"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBaseEntity creates a new BaseEntity with a generated UUID.
func NewBaseEntity() BaseEntity {
	now := time.Now().UTC()
	return BaseEntity{
		ID:        uuid.New(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IncrementVersion bumps the version for optimistic locking.
func (e *BaseEntity) IncrementVersion() {
	e.Version++
	e.UpdatedAt = time.Now().UTC()
}

// SetVersion sets the version explicitly (used when loading from DB).
func (e *BaseEntity) SetVersion(v int64) {
	e.Version = v
}

// AuditableEntity extends BaseEntity with audit tracking fields.
type AuditableEntity struct {
	BaseEntity
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewAuditableEntity creates a new AuditableEntity.
func NewAuditableEntity(createdBy *uuid.UUID) AuditableEntity {
	return AuditableEntity{
		BaseEntity: NewBaseEntity(),
		CreatedBy:  createdBy,
	}
}

// SoftDelete marks the entity as deleted.
func (e *AuditableEntity) SoftDelete(deletedBy *uuid.UUID) {
	now := time.Now().UTC()
	e.DeletedAt = &now
	e.UpdatedBy = deletedBy
	e.UpdatedAt = now
}

// IsDeleted returns true if the entity has been soft-deleted.
func (e *AuditableEntity) IsDeleted() bool {
	return e.DeletedAt != nil
}

// DomainEvent represents something that happened in the domain.
type DomainEvent struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	AggregateID uuid.UUID             `json:"aggregate_id"`
	Version     int64                  `json:"version"`
	Payload     map[string]interface{} `json:"payload"`
	OccurredAt  time.Time              `json:"occurred_at"`
}

// NewDomainEvent creates a new domain event.
func NewDomainEvent(eventType string, aggregateID uuid.UUID, version int64, payload map[string]interface{}) DomainEvent {
	return DomainEvent{
		ID:          uuid.New(),
		Type:        eventType,
		AggregateID: aggregateID,
		Version:     version,
		Payload:     payload,
		OccurredAt:  time.Now().UTC(),
	}
}

// AggregateRoot is the base for all aggregate roots with domain event support.
type AggregateRoot struct {
	AuditableEntity
	domainEvents []DomainEvent
}

// NewAggregateRoot creates a new AggregateRoot.
func NewAggregateRoot(createdBy *uuid.UUID) AggregateRoot {
	return AggregateRoot{
		AuditableEntity: NewAuditableEntity(createdBy),
		domainEvents:    make([]DomainEvent, 0),
	}
}

// AddDomainEvent appends a domain event to the aggregate.
func (ar *AggregateRoot) AddDomainEvent(event DomainEvent) {
	ar.domainEvents = append(ar.domainEvents, event)
}

// GetDomainEvents returns all pending domain events.
func (ar *AggregateRoot) GetDomainEvents() []DomainEvent {
	return ar.domainEvents
}

// ClearDomainEvents removes all pending domain events (called after publishing).
func (ar *AggregateRoot) ClearDomainEvents() {
	ar.domainEvents = make([]DomainEvent, 0)
}

// HasDomainEvents returns true if there are pending domain events.
func (ar *AggregateRoot) HasDomainEvents() bool {
	return len(ar.domainEvents) > 0
}
