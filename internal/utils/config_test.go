package nomnom

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config := LoadConfig("../../config.json")
	fmt.Println(config)
}

func TestLoadConfigWithHomeDir(t *testing.T) {
	config := LoadConfig("")
	fmt.Println(config)
}
