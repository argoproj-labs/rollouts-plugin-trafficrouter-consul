package utils

import (
	"os"

	"log/slog"

	pluginTypes "github.com/argoproj/argo-rollouts/utils/plugin/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKubeConfig() (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return nil, pluginTypes.RpcError{ErrorString: err.Error()}
	}
	return config, nil
}

func InitLogger(lvl slog.Level) {
	lvlVar := &slog.LevelVar{}
	lvlVar.Set(lvl)
	opts := slog.HandlerOptions{
		Level: lvlVar,
	}

	attrs := []slog.Attr{
		slog.String("plugin", "trafficrouter"),
		slog.String("vendor", "consul"),
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &opts).WithAttrs(attrs))
	slog.SetDefault(l)
}
