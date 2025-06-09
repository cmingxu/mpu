package model

var AudioCreationSchema = `
CREATE TABLE IF NOT EXISTS audios (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL, -- 音频名称
	path TEXT NOT NULL -- 音频路径
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_audios_name ON audios(name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_audios_path ON audios(path);
`

type Audio struct {
	Name string `json:"name"` // 音频名称
	Path string `json:"path"` // 音频路径
}

func NewAudio(name, path string) *Audio {
	return &Audio{
		Name: name,
		Path: path,
	}

}

func (a *Audio) Create() error {
	if _, err := db.NamedExec("INSERT INTO audios (name, path) VALUES (:name, :path)", a); err != nil {
		return err
	}
	return nil
}

func (a *Audio) Update() error {
	if _, err := db.NamedExec("UPDATE audios SET path = :path WHERE name = :name", a); err != nil {
		return err
	}
	return nil
}

func (a *Audio) Delete() error {
	if _, err := db.NamedExec("DELETE FROM audios WHERE name = :name", a); err != nil {
		return err
	}

	return nil
}

func (a *Audio) List() ([]*Audio, error) {
	var audios []*Audio
	if err := db.Select(&audios, "SELECT * FROM audios"); err != nil {
		return nil, err
	}
	return audios, nil
}
