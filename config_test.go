package main

import (
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	c, _ := getConfigFromFile("config.yaml")
	log.Printf("config %#v\n", c)
}
