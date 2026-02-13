package config

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type Config struct {
	RootPath string `json:"root_path"`
	Folders  struct {
		Fleeting  string `json:"fleeting"`
		Permanent string `json:"permanent"`
		Tmp       string `json:"tmp"`
	} `json:"folders"`
}

func Load() (*Config, error) {
	ctx := cuecontext.New()
	val := ctx.CompileString(DefaultSchema).LookupPath(cue.ParsePath("#Config"))

	// Unify with user file logic here...

	var cfg Config
	err := val.Decode(&cfg)
	return &cfg, err
}
