package database

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AppDB struct {
	db *gorm.DB
}

func (a *AppDB) GormDB() *gorm.DB {
	return a.db
}

type Startup struct {
	ID        uint `gorm:"primaryKey"`
	Timestamp time.Time
	Host      string
}

type CommandRevision struct {
	ID       string `gorm:"primaryKey"`
	Revision int
}

func GetDSN(path string) string {
	q := url.Values{}
	q.Set("_journal_mode", "WAL")
	q.Set("_busy_timeout", "5000")
	q.Set("_txlock", "immediate")

	if strings.Contains(path, "?") {
		return path + "&" + q.Encode()
	}
	return path + "?" + q.Encode()
}

func InitDB(path string) (*AppDB, error) {
	db, err := gorm.Open(sqlite.Open(GetDSN(path)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&Startup{},
		&CommandRevision{},
	); err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	db.Create(&Startup{
		Timestamp: time.Now(),
		Host:      hostname,
	})

	return &AppDB{db: db}, nil
}

func (a *AppDB) GetCommandRevision(id string) (int, error) {
	var cr CommandRevision
	err := a.db.First(&cr, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return cr.Revision, nil
}

func (a *AppDB) SetCommandRevision(id string, rev int) error {
	var cr CommandRevision
	cr.ID = id
	cr.Revision = rev
	return a.db.Save(&cr).Error
}
