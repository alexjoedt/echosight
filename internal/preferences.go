package echosight

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/uptrace/bun"
)

// Preferences.
// The map is inside a struct because this way we can change the implementation
type Preferences struct {
	index int
	keys  []string
	prefs map[string]string
}

func (p *Preferences) Next() bool {
	if p.keys == nil {
		p.keys = make([]string, 0)
		for k := range p.prefs {
			p.keys = append(p.keys, k)
		}
	}

	return p.index < len(p.prefs)
}

func (p *Preferences) Pref() *Preference {
	if p.Next() {
		key := p.keys[p.index]
		pref := Preference{
			Name:  key,
			Value: p.prefs[key],
		}
		p.index++
		return &pref
	}
	return nil
}

func (p *Preferences) Decode(v any) error {
	return mapstructure.Decode(p.prefs, v)
}

func (p *Preferences) Map() map[string]string {
	return p.prefs
}

func (p *Preferences) Set(key string, value string) {
	if p.prefs == nil {
		p.prefs = make(map[string]string, 0)
	}

	p.prefs[key] = value
}

func (p *Preferences) Get(key string) string {
	if p.prefs == nil {
		return ""
	}

	if v, ok := p.prefs[key]; ok {
		return v
	}
	return ""
}

func (p *Preferences) Has(key string) bool {
	if p.prefs == nil {
		return false
	}

	v, ok := p.prefs[key]
	return ok && v != ""
}

func (p *Preferences) Delete(key string) {
	if p.prefs == nil {
		return
	}
	_, ok := p.prefs[key]
	if ok {
		delete(p.prefs, key)
	}
}

// Crypt crypts the value of the given key and removes the uncrypted entry
func (p *Preferences) Crypt(crypter Crypter, key string) error {
	if !p.Has(key) {
		return fmt.Errorf("key not present")
	}

	if strings.HasSuffix(key, "_crypt") {
		return fmt.Errorf("value already marked as crypted")
	}

	crypted, err := crypter.Encrypt(p.Get(key))
	if err != nil {
		return err
	}

	p.Set(key+"_crypt", crypted)
	p.Delete(key)

	return nil
}

func (p *Preferences) CryptValues(crypter Crypter, suffixe ...string) error {
	for _, suffix := range suffixe {
		for k := range p.prefs {
			if strings.HasSuffix(k, suffix) {
				if err := p.Crypt(crypter, k); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *Preferences) UnmarshalJSON(b []byte) error {
	var v map[string]string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	p.prefs = v
	return nil
}

func (p *Preferences) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.prefs)
}

type Preference struct {
	bun.BaseModel `bun:"table:preferences"`
	ID            uuid.UUID `json:"id" bun:"type:uuid,pk,default:uuid_generate_v4()"`
	LookupVersion int       `json:"lookupVersion" bun:",default:1"`
	Name          string    `json:"name"`
	Value         string    `json:"value"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

var (
	// PrefSafelist list of valid preference names
	PrefSafelist = []string{
		"smtp_host",
		"smtp_port",
		"smtp_user",
		"smtp_sender",
		"smtp_password",
		"smtp_password_crypt",
		"smtp_enabled",
		"telegram_bot_token",
		"telegram_chat_ids", // comma seperated list of chat ids
		"telegram_enabled",
	}
)

func (p *Preferences) Validate(v *validator.Validator) {
	for p.Next() {
		pref := p.Pref()
		v.Check(pref.Name != "", pref.Name, "name must not be empty")
		v.Check(pref.Value != "", pref.Name+".value", "value must not be empty")
		v.Check(validator.PermittedValue(pref.Name, PrefSafelist...), "name", "invalid preference name")
	}
}
