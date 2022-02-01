package postgres

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) (*Index, error) {
	cfg := &gorm.Config{
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
	}

	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, err
	}

	return &Index{db: db}, nil
}

type Index struct {
	once sync.Once
	db   *gorm.DB
}

func (index *Index) Index(docs ...interface{}) (err error) {
	if len(docs) == 0 {
		return nil
	}

	index.once.Do(func() {
		err = index.db.AutoMigrate(docs[0])
	})

	if err != nil {
		return err
	}

	return index.db.Transaction(func(tx *gorm.DB) error {
		for _, doc := range docs {
			if err := tx.Create(doc).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (index *Index) Close() error {
	db, err := index.db.DB()
	if err != nil {
		return err
	}

	return db.Close()
}
