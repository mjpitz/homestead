package postgres

import (
	"context"

	"github.com/mjpitz/homestead/internal/index"
	"github.com/mjpitz/myago/zaputil"
)

type Builder struct {
	Action func(ctx context.Context, index index.Index) error
}

func (b Builder) performAction(ctx context.Context, cfg index.Config) error {
	idx, err := Open(cfg.Endpoint)
	if err != nil {
		return err
	}

	defer func() {
		// ensure the index is closed
		err = idx.Close()

		if err != nil {
			zaputil.Extract(ctx).Error(err.Error())
		}
	}()

	return b.Action(ctx, idx)
}

func (b Builder) Run(ctx context.Context, cfg index.Config) error {
	log := zaputil.Extract(ctx)

	log.Info("updating index")
	err := b.performAction(ctx, cfg)
	if err != nil {
		return err
	}

	return nil
}
