package annotators

import (
	"context"

	"gopkg.in/yaml.v2"
)

type RmLogCall struct {
	Type        string
	Import      string
	Function    string
	Concurrency int
}

func init() {
	Register(&RmLogCall{})
}

func (lc *RmLogCall) Unmarshal(buf []byte) error {
	return yaml.Unmarshal(buf, lc)
}

func (lc *RmLogCall) Describe() string {
	return "rmlogcall:\n"
}

func (lc *RmLogCall) Do(ctx context.Context, pkgs []string) error {
	return nil
}
