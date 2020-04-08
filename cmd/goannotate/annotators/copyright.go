package annotators

import (
	"context"

	"gopkg.in/yaml.v2"
)

type EnsureCopyright struct {
	Type      string `annotator:"name of annotator."`
	Copyright string `annotator:"copyright notice to inserted."`
}

func init() {
	Register(&EnsureCopyright{})
}

func (ec *EnsureCopyright) Unmarshal(buf []byte) error {
	return yaml.Unmarshal(buf, ec)
}

func (ec *EnsureCopyright) Describe() string {
	return MustDescribe(ec,
		`an annotator that ensures that a copyright notice is 
present at the top of all files. It will not remove existing notices.`,
	)
}

func (ec *EnsureCopyright) Do(ctx context.Context, pkgs []string) error {
	return nil
}
