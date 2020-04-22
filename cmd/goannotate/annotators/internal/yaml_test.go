package internal_test

import (
	"testing"

	"cloudeng.io/go/cmd/goannotate/annotators/internal"
	"gopkg.in/yaml.v2"
)

type packed struct {
	MapSlice yaml.MapSlice
	Type     string `yaml:"type"`
	Num      int    `yaml:"num"`
}

func (p *packed) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return internal.DelegatedYAML(p, unmarshal)
}

func TestYAML(t *testing.T) {
	p := &packed{}
	err := yaml.Unmarshal([]byte(`type: typeName
o1: A
o2: B
num: 123
`), p)
	if err != nil {
		t.Errorf("Unmarshal: %v", err)
	}
	if got, want := p.Type, "typeName"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := p.Num, 123; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
