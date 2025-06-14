package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type State string

func StateFromString(s string) State {
	switch s {
	case "init":
		return StateInit
	default:
		return State(s)
	}
}

func (s State) String() string {
	return string(s)
}

const (
	StateInit State = "init" // 正常
)

var MovieCreationSchema = `
CREATE TABLE IF NOT EXISTS movies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	tpl_name TEXT NOT NULL, -- 模板类型
	idea TEXT, -- 电影创意
	state TEXT NOT NULL, -- 状态
	title TEXT, -- 标题
	footer TEXT, -- 底部
	icon TEXT, -- 图标
	script TEXT, -- 内容
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- 创建时间
);
`

type Movie struct {
	Id        int64          `db:"id"`                   // 电影ID
	TplName   string         `db:"tpl_name"`             // 模板类型
	State     string         `db:"state"`                // 状态
	Idea      sql.NullString `db:"idea" json:"idea"`     // 电影创意
	Title     sql.NullString `db:"title" json:"title"`   // 标题
	Footer    sql.NullString `db:"footer" json:"footer"` // 底部
	Icon      sql.NullString `db:"icon" json:"icon"`     // 图标
	Script    sql.NullString `db:"script" json:"script"` // 内容
	CreatedAt time.Time      `db:"created_at"`           // 创建时间

}

type MovieScript struct {
	Title       string        `json:"title"`        // 视频标题
	ScriptItems []*ScriptItem `json:"script_items"` // 视频脚本内容
}

type ScriptItem struct {
	ZhSubtitle  string `json:"cn"`                   // Chinese subtitle
	EnSubtitle  string `json:"en"`                   // English subtitle
	VoicePath   string `json:"voice_path,omitempty"` // Path to the voice file
	ImagePrompt string `json:"image_prompt"`         // Image generation prompt
	ImagePath   string `json:"image_path,omitempty"` // Path to the generated image
}

func NewMovie() *Movie {
	return &Movie{
		TplName: string(Sign),
		State:   StateInit.String(),
		Idea:    sql.NullString{},
		Title:   sql.NullString{},
		Footer:  sql.NullString{},
		Icon:    sql.NullString{},
		Script:  sql.NullString{},
	}
}

func (m *Movie) Create() error {
	result, err := db.NamedExec("INSERT INTO movies (tpl_name, state, idea, title, footer, icon, script)"+
		"VALUES (:tpl_name, :state, :idea, :title, :footer, :icon, :script)", m)
	if err != nil {
		return errors.Wrap(err, "failed to create movie")
	}

	m.Id, _ = result.LastInsertId()
	return nil
}

func (m *Movie) Update() error {
	if _, err := db.NamedExec("UPDATE movies SET state = :state, idea = :idea, title = :title, footer = :footer, icon = :icon, script = :script WHERE id = :id", m); err != nil {
		return errors.Wrap(err, "failed to update movie")
	}

	fmt.Println("11111111111111")

	return nil
}

func (m *Movie) GetScript() (*MovieScript, error) {
	var script MovieScript
	script.ScriptItems = make([]*ScriptItem, 0)

	if !m.Script.Valid {
		return &script, nil // No script content available
	}

	if err := json.Unmarshal([]byte(m.Script.String), &script); err != nil {
		return &script, errors.Wrap(err, "failed to unmarshal script content")
	}

	return &script, nil
}

func ListMovies() ([]*Movie, error) {
	var movies []*Movie

	if err := db.Select(&movies, "SELECT * FROM movies ORDER BY created_at DESC"); err != nil {
		return nil, errors.Wrap(err, "failed to list movies")
	}

	return movies, nil
}

func MoviesCount() (int64, error) {
	var count int64

	err := db.Get(&count, "SELECT COUNT(*) FROM movies")
	if err != nil {
		return 0, errors.Wrap(err, "failed to count movies")
	}

	return count, nil
}

func GetMovie(id int64) (*Movie, error) {
	var movie Movie
	if err := db.Get(&movie, "SELECT * FROM movies WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &movie, nil
}

func (m *Movie) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id        int64     `json:"id"`
		TplName   string    `json:"tpl_name"`
		State     string    `json:"state"`
		Idea      string    `json:"idea"`
		Title     string    `json:"title"`
		Footer    string    `json:"footer"`
		Icon      string    `json:"icon"`
		Script    string    `json:"script"`
		CreatedAt time.Time `json:"created_at"`
	}{
		Id:        m.Id,
		TplName:   m.TplName,
		State:     m.State,
		Idea:      m.Idea.String,
		Title:     m.Title.String,
		Footer:    m.Footer.String,
		Icon:      m.Icon.String,
		Script:    m.Script.String,
		CreatedAt: m.CreatedAt,
	})
}
