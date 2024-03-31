package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var _ echosight.DetectorService = (*DetectorModel)(nil)

type DetectorModel struct {
	db  *bun.DB
	log *logger.Logger
}

func (m *DetectorModel) Create(ctx context.Context, detector *echosight.Detector) error {
	detector.CreatedAt = time.Now()

	_, err := m.db.NewInsert().
		Model(detector).
		Ignore(). // on conflict do nothing
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to insert detector", err)
		return echosight.ErrInternalf("failed to insert detector")
	}

	return nil
}

func (m *DetectorModel) GetByID(ctx context.Context, id uuid.UUID) (*echosight.Detector, error) {
	detector := new(echosight.Detector)
	err := m.db.NewSelect().Model(detector).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, echosight.ErrNotfoundf("no detector found")
		}
		m.log.Errorc("failed to get detector by id", err, logger.UUID("host_id", id))
		return nil, err
	}

	return detector, nil
}

func (m *DetectorModel) GetByName(ctx context.Context, name string) (*echosight.Detector, error) {
	detector := new(echosight.Detector)
	err := m.db.NewSelect().Model(detector).
		Where("name = ?", name).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, echosight.ErrNotfoundf("no detector found")
		}
		m.log.Errorc("failed to get detector by name", err, logger.Str("name", name))
		return nil, echosight.ErrInternalf("failed to get detector by name").WithError(err)
	}

	return detector, nil
}

func (m *DetectorModel) Update(ctx context.Context, detector *echosight.Detector) error {
	detector.UpdatedAt = time.Now()
	lv := detector.LookupVersion
	detector.LookupVersion++

	_, err := m.db.NewUpdate().Model(detector).
		Where("id = ? AND lookup_version = ?", detector.ID, lv).
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to update detector", err, logger.UUID("user_id", detector.ID))
		return echosight.ErrInternalf("failed to update detector").WithError(err)
	}

	return nil
}

func (m *DetectorModel) DeleteByID(ctx context.Context, id uuid.UUID) (*echosight.Detector, error) {
	detector := new(echosight.Detector)
	err := m.db.NewDelete().Model(detector).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		m.log.Errorc("failed to delete detector", err)
		return nil, echosight.ErrInternalf("failed to delete detector")
	}
	return detector, nil
}

func (m *DetectorModel) List(ctx context.Context, detectorFilter *filter.DetectorFilter) ([]*echosight.Detector, error) {
	users := make([]*echosight.Detector, 0)
	query := m.db.NewSelect().Model(&users)

	if detectorFilter.HostID != nil {
		query.Where("host_id = ?", *detectorFilter.HostID)
	}

	if detectorFilter.Name != nil {
		query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(*detectorFilter.Name)+"%")
	}

	if detectorFilter.Type != nil {
		query.Where("LOWER(dns) LIKE ?", "%"+strings.ToLower(*detectorFilter.Type)+"%")
	}

	if detectorFilter.Active != nil {
		query.Where("active = ?", *detectorFilter.Active)
	}

	count, err := query.
		Limit(detectorFilter.Limit()).
		Offset(detectorFilter.Offset()).
		Order(detectorFilter.Order()).
		ScanAndCount(ctx)

	if err != nil {
		return nil, err
	}

	detectorFilter.Pagination = filter.ComputePagination(count, detectorFilter.Page, detectorFilter.PageSize)
	return users, nil
}
