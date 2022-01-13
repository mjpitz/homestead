package index

import (
	"context"
	"path/filepath"

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
	Action func(ctx context.Context, index Index) error
}

func (b Builder) performAction(ctx context.Context, cfg Config) error {
	afs := vfs.Extract(ctx)
	_ = afs.MkdirAll(cfg.Path, 0755)

	index, err := OpenSQLite(filepath.Join(cfg.Path, "db.sqlite"), false)
	if err != nil {
		return err
	}

	return b.Action(ctx, index)
}

func (b Builder) Run(ctx context.Context, cfg Config) error {
	log := zaputil.Extract(ctx)

	var agentAPI agentv1.AgentAPIClient

	if cfg.Target != "" {
		conn, err := grpc.Dial(cfg.Target, grpc.WithInsecure(), grpc.WithBlock())
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
