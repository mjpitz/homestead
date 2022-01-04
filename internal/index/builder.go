package index

import (
	"context"

	"github.com/blevesearch/bleve/v2"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	agentv1 "github.com/mjpitz/aetherfs/api/aetherfs/agent/v1"
	"github.com/mjpitz/myago/vfs"
	"github.com/mjpitz/myago/zaputil"
)

type Config struct {
	Path   string           `json:"path"   usage:"specify the location of the index"`
	Target string           `json:"target" usage:"specify the address of the target AetherFS instance"`
	Tags   *cli.StringSlice `json:"tags"   usage:"tags to publish for the index"`
}

type Builder struct {
	Action func(ctx context.Context, index bleve.Index) error
}

func (b Builder) performAction(ctx context.Context, cfg Config) error {
	var index bleve.Index

	afs := vfs.Extract(ctx)
	exists, err := afero.Exists(afs, cfg.Path)
	switch {
	case exists:
		index, err = bleve.Open(cfg.Path)
	case err == nil:
		_ = afs.MkdirAll(cfg.Path, 0755)
		index, err = bleve.New(cfg.Path, bleve.NewIndexMapping())
	}

	if err != nil {
		return err
	}
	defer index.Close()

	return b.Action(ctx, index)
}

func (b Builder) Run(ctx context.Context, cfg Config) error {
	log := zaputil.Extract(ctx)

	var agentAPI agentv1.AgentAPIClient

	if cfg.Target != "" {
		conn, err := grpc.Dial(cfg.Target, grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer conn.Close()

		agentAPI = agentv1.NewAgentAPIClient(conn)

		defer func() {
			log.Info("triggering graceful shutdown")
			_, _ = agentAPI.GracefulShutdown(ctx, &agentv1.GracefulShutdownRequest{})
			log.Info("shutdown complete")
		}()
	}

	log.Info("updating index")
	err := b.performAction(ctx, cfg)
	if err != nil {
		return err
	}

	if cfg.Target != "" {
		log.Info("publishing index")
		_, err = agentAPI.Publish(ctx, &agentv1.PublishRequest{
			Sync:      true,
			Path:      cfg.Path,
			Tags:      cfg.Tags.Value(),
			BlockSize: 64 * (1 << 20),
		})

		if err != nil {
			return err
		}
	}

	return nil
}
