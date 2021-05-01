package main

import (
	_ "embed"
	"fmt"
	"github.com/spongeprojects/magicconch"
	"gopkg.in/yaml.v3"
)

//go:embed trans.yaml
var transYAML []byte

var trans map[string]string

func t(messageID string, args ...interface{}) string {
	return fmt.Sprintf(trans[messageID], args...)
}

func init() {
	magicconch.Must(yaml.Unmarshal(transYAML, &trans))
}
