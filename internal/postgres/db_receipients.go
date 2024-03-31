package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	es "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var _ es.RecipientService = (*RecipientModel)(nil)

type RecipientModel struct {
	db  *bun.DB
	log *logger.Logger
}

func (m *RecipientModel) Create(ctx context.Context, recipient *es.Recipient) error {
	recipient.CreatedAt = time.Now()
	_, err := m.db.NewInsert().
		Model(recipient).
		Ignore(). // on conflict do nothing
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to insert recipient", err)
		return err
	}

	return nil
}

func (m *RecipientModel) GetByID(ctx context.Context, id uuid.UUID) (*es.Recipient, error) {
	recipient := new(es.Recipient)
	err := m.db.NewSelect().Model(recipient).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, es.ErrNotfoundf("no recipient found")
		}
		m.log.Errorc("failed to get recipient by id", err, logger.UUID("recipient_id", id))
		return nil, err
	}

	return recipient, nil
}

func (m *RecipientModel) GetByEmail(ctx context.Context, email string) (*es.Recipient, error) {
	recipient := new(es.Recipient)
	err := m.db.NewSelect().Model(recipient).
		Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, es.ErrNotfoundf("no recipient found")
		}
		m.log.Errorc("failed to get recipient by email", err, logger.Str("email", email))
		return nil, es.ErrInternalf("failed to get recipient by email").WithError(err)
	}

	return recipient, nil
}

func (m *RecipientModel) Update(ctx context.Context, rcpt *es.Recipient) error {
	rcpt.UpdatedAt = time.Now()
	lv := rcpt.LookupVersion
	rcpt.LookupVersion++

	_, err := m.db.NewUpdate().Model(rcpt).
		Where("id = ? AND lookup_version = ?", rcpt.ID, lv).
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to update recipient", err, logger.UUID("recipient_id", rcpt.ID))
		return es.ErrInternalf("failed to update recipient").WithError(err)
	}

	return nil
}

func (m *RecipientModel) DeleteByID(ctx context.Context, id uuid.UUID) (*es.Recipient, error) {
	recipient := new(es.Recipient)
	err := m.db.NewDelete().Model(recipient).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return recipient, nil
}

func (m *RecipientModel) List(ctx context.Context, rcptFilter *filter.RecipientFilter) ([]*es.Recipient, error) {
	recipients := make([]*es.Recipient, 0)
	query := m.db.NewSelect().Model(&recipients)

	if rcptFilter.Name != nil {
		query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(*rcptFilter.Name)+"%")
	}

	if rcptFilter.HostID != nil {
		query.Where("host_id = ?", *rcptFilter.HostID)
	}

	if rcptFilter.Active != nil {
		query.Where("activated = ?", *rcptFilter.Active)
	}

	if rcptFilter.Email != nil {
		query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(*rcptFilter.Email)+"%")
	}

	count, err := query.
		Limit(rcptFilter.Limit()).
		Offset(rcptFilter.Offset()).
		Order(rcptFilter.Order()).
		ScanAndCount(ctx)

	if err != nil {
		return nil, err
	}

	rcptFilter.Pagination = filter.ComputePagination(count, rcptFilter.Page, rcptFilter.PageSize)
	return recipients, nil
}
