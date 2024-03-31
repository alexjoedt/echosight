package postgres

import (
	"context"
	"time"

	es "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/uptrace/bun"
)

var _ es.PreferenceService = (*PreferenceModel)(nil)

type PreferenceModel struct {
	db  *bun.DB
	log *logger.Logger
}

func (p *PreferenceModel) AllPreferences(ctx context.Context) (*es.Preferences, error) {
	var prefs []*es.Preference
	err := p.db.NewSelect().Model(&prefs).Scan(ctx)
	if err != nil {
		p.log.Errorc("", err)
		return nil, es.ErrNotfoundf("no preferences found").WithError(err)
	}

	var preferences es.Preferences
	for _, pf := range prefs {
		preferences.Set(pf.Name, pf.Value)
	}

	return &preferences, nil
}

func (p *PreferenceModel) GetByName(ctx context.Context, name string) (*es.Preference, error) {
	var pref *es.Preference
	err := p.db.NewSelect().Model(pref).Scan(ctx)
	if err != nil {
		p.log.Errorc("", err, logger.Str("name", name))
		return nil, es.ErrInternalf("no preference found with name: '%s'", name).WithError(err)
	}

	return pref, nil
}

func (p *PreferenceModel) List(ctx context.Context, prefFilter *filter.PreferenceFilter) (*es.Preferences, error) {
	var prefs []*es.Preference
	query := p.db.NewSelect().Model(&prefs)

	if prefFilter.Name != "" {
		query.Where("name LIKE ?", "%"+prefFilter.Name+"%")
	}

	if prefFilter.Value != "" {
		query.Where("value = ?", prefFilter.Value)
	}

	err := query.Scan(ctx)
	if err != nil {
		p.log.Errorc("", err)
		return nil, es.ErrInternalf("failed to list preferences").WithError(err)
	}

	var preferences es.Preferences
	for _, pf := range prefs {
		preferences.Set(pf.Name, pf.Value)
	}

	return &preferences, nil
}

// Set sets the preference by name. If a pref exists with the given name, it will be
// overwritten.
func (p *PreferenceModel) Set(ctx context.Context, pref *es.Preference) error {
	_, err := p.db.NewDelete().Model(pref).Where("name = ?", pref.Name).Exec(ctx)
	if err != nil {
		p.log.Debugw("failed to delete preference", logger.Str("error", err.Error()), logger.Str("name", pref.Name))
	}

	_, err = p.db.NewInsert().Model(pref).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p *PreferenceModel) Update(ctx context.Context, pref *es.Preference) error {
	pref.UpdatedAt = time.Now()
	lv := pref.LookupVersion
	pref.LookupVersion++
	p.db.NewUpdate().Model(pref).
		Where("id = ?", pref.ID).
		Where("lookup_version = ?", lv).Scan(ctx)
	return nil
}

func (p *PreferenceModel) SetAll(ctx context.Context, prefs *es.Preferences) error {
	for name, pref := range prefs.Map() {
		if err := p.Set(ctx, &es.Preference{
			Name:  name,
			Value: pref,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (p *PreferenceModel) DeleteByName(ctx context.Context, name string) error {
	_, err := p.db.NewDelete().Model(&es.Preference{}).Where("name = ?", name).Exec(ctx)
	if err != nil {
		p.log.Errorc("", err, logger.Str("name", name))
		return es.ErrInternalf("failed to delete preference: '%s'", name).WithError(err)
	}

	return nil
}
