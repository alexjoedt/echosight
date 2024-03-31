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

var _ echosight.HostService = (*HostModel)(nil)

type HostModel struct {
	db  *bun.DB
	log *logger.Logger
}

func (m *HostModel) Create(ctx context.Context, host *echosight.Host) error {
	host.CreatedAt = time.Now()
	_, err := m.db.NewInsert().
		Model(host).
		Ignore(). // on conflict do nothing
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to insert host", err)
		return echosight.ErrInternalf("failed to insert host")
	}

	return nil
}

func (m *HostModel) GetByID(ctx context.Context, id uuid.UUID) (*echosight.Host, error) {
	detector := new(echosight.Host)
	err := m.db.NewSelect().Model(detector).Relation("Detectors").
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, echosight.ErrNotfoundf("no host found")
		}
		m.log.Errorc("failed to get host by id", err, logger.UUID("host_id", id))
		return nil, err
	}

	return detector, nil
}

func (m *HostModel) GetByName(ctx context.Context, name string) (*echosight.Host, error) {
	detector := new(echosight.Host)
	err := m.db.NewSelect().Model(detector).
		Where("name = ?", name).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, echosight.ErrNotfoundf("no host found")
		}
		m.log.Errorc("failed to get host by name", err, logger.Str("name", name))
		return nil, echosight.ErrInternalf("failed to get host by name").WithError(err)
	}

	return detector, nil
}

func (m *HostModel) Update(ctx context.Context, host *echosight.Host) error {
	host.UpdatedAt = time.Now()
	lv := host.LookupVersion
	host.LookupVersion++

	_, err := m.db.NewUpdate().Model(host).
		Where("id = ? AND lookup_version = ?", host.ID, lv).
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to update detector", err, logger.UUID("user_id", host.ID))
		return echosight.ErrInternalf("failed to update detector").WithError(err)
	}

	return nil
}

func (m *HostModel) DeleteByID(ctx context.Context, id uuid.UUID) (*echosight.Host, error) {
	host := new(echosight.Host)
	err := m.db.NewDelete().Model(host).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		m.log.Errorc("failed to delete host", err)
		return nil, echosight.ErrInternalf("failed to delete host")
	}
	return host, nil
}

func (m *HostModel) List(ctx context.Context, hostFilter *filter.HostFilter) ([]*echosight.Host, error) {
	users := make([]*echosight.Host, 0)
	query := m.db.NewSelect().Model(&users)

	if hostFilter.Name != "" {
		query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(hostFilter.Name)+"%")
	}

	if hostFilter.DNS != "" {
		query.Where("LOWER(dns) LIKE ?", "%"+strings.ToLower(hostFilter.DNS)+"%")
	}

	if hostFilter.IPv4 != "" {
		query.Where("LOWER(ipv4) LIKE ?", "%"+strings.ToLower(hostFilter.IPv4)+"%")
	}

	if hostFilter.IPv6 != "" {
		query.Where("LOWER(ipv6) LIKE ?", "%"+strings.ToLower(hostFilter.IPv6)+"%")
	}

	if hostFilter.Location != "" {
		query.Where("LOWER(location) LIKE ?", "%"+strings.ToLower(hostFilter.Location)+"%")
	}

	count, err := query.Relation("Detectors").
		Limit(hostFilter.Limit()).
		Offset(hostFilter.Offset()).
		Order(hostFilter.Order()).
		ScanAndCount(ctx)

	if err != nil {
		return nil, err
	}

	hostFilter.Pagination = filter.ComputePagination(count, hostFilter.Page, hostFilter.PageSize)
	return users, nil
}
