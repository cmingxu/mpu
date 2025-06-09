package model

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type TplType string

const (
	Sign TplType = "sign" // 星座
)

var TemplateCreationSchema = `
CREATE TABLE IF NOT EXISTS templates(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL, -- 模板名称
	created_at timestamp NOT NULL -- 创建时间
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_name ON templates(name);
`

var TemplateInitializationStat = `
INSERT INTO templates (name, created_at)VALUES('sign', current_timestamp);
`

type Template struct {
	Id        int64     `json:"id" db:"id"`                 // 模板ID
	Name      string    `json:"name" db:"name"`             // 模板名称
	CreatedAt time.Time `json:"created_at" db:"created_at"` // 创建时间
}

func NewTemplate(name string) *Template {
	return &Template{
		Name: name,
	}
}

func (t *Template) Create() error {
	if _, err := db.NamedExec("INSERT INTO templates (name, createdAt) VALUES (:name, current_timestamp())", t); err != nil {
		return errors.Wrap(err, "failed to create template")
	}

	return nil
}

func (t *Template) Update() error {
	err := db.Get(t, "SELECT * FROM templates WHERE name = ?", t.Name)

	if err != nil {
		return errors.Wrap(err, "failed to get template")
	}

	if _, err := db.NamedExec("UPDATE templates SET createdAt = current_timestamp() WHERE name = :name", t); err != nil {
		return errors.Wrap(err, "failed to update template")
	}

	return nil
}

func GetTemplate(id int64) (*Template, error) {
	var template Template

	if err := db.Get(&template, "SELECT * FROM templates WHERE id = ?", id); err != nil {
		return nil, errors.Wrapf(err, "failed to get template with id %d", id)
	}

	return &template, nil
}

func ListTemplates() ([]Template, error) {
	var templates []Template

	if err := db.Select(&templates, "SELECT * FROM templates ORDER BY created_at DESC"); err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}

	return templates, nil
}

func GetTemplateByName(name string) (*Template, error) {
	var template Template

	if err := db.Get(&template, "SELECT * FROM templates WHERE name = ?", name); err != nil {
		return nil, errors.Wrapf(err, "failed to get template with name %s", name)
	}

	return &template, nil
}

func DefaultTemplates() (*Template, error) {
	d := &Template{}

	if err := db.Get(d, "SELECT * FROM templates where name = 'sign'"); err != nil {
		return nil, err
	}

	return d, nil
}

func (t Template) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id        int64     `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
	}{
		Id:        t.Id,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	})
}
