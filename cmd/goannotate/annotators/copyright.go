package annotators

import (
	"context"

	"gopkg.in/yaml.v2"
)

type EnsureCopyright struct {
	Type      string `annotator:"name of annotator type."`
	Name      string `annotator:"name of annotator configuration."`
	Copyright string `annotator:"copyright notice to inserted."`
}

func init() {
	Register(&EnsureCopyright{})
}

// New implements annotators.T.
func (ec *EnsureCopyright) New(name string) T {
	return &EnsureCopyright{Name: name}
}

// Unmarshal implements annotators.T.
func (ec *EnsureCopyright) Unmarshal(buf []byte) error {
	return yaml.Unmarshal(buf, ec)
}

// Describe implements annotators.T.
func (ec *EnsureCopyright) Describe() string {
	return MustDescribe(ec,
		`an annotator that ensures that a copyright notice is 
present at the top of all files. It will not remove existing notices.`,
	)
}

// Do implements annotators.T.
func (ec *EnsureCopyright) Do(ctx context.Context, pkgs []string) error {
	return nil
}
