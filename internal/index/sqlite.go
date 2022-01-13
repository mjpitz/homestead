package index

import (
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenSQLite(dsn string, readonly bool) (*SQLite, error) {
	hasOptions := strings.LastIndex(dsn, "?") >= 0
	if hasOptions {
		dsn += "&_journal=wal"
	} else {
		dsn += "?_journal=wal"
	}

	if readonly {
		dsn += "&mode=ro&immutable=true"
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		return nil, err
	}

	return &SQLite{db}, nil
}

type SQLite struct {
	db *gorm.DB
}

func (sqlite *SQLite) Index(docs ...interface{}) error {
	if len(docs) == 0 {
		return nil
	}

	err := sqlite.db.AutoMigrate(docs[0])
	if err != nil {
		return err
	}

	return sqlite.db.Transaction(func(tx *gorm.DB) error {
		for _, doc := range docs {
			if err := tx.Create(doc).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
