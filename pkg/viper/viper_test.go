package viper

import (
	"fmt"
	"testing"
)

func TestViper(t *testing.T) {
	apiConfig := Init("api")
	var otherServices map[string]string
	if err := apiConfig.Viper.UnmarshalKey("otherService", &otherServices); err != nil {
		t.Fatalf("Unable to unmarshal 'otherService' into map: %v", err)
	}
	fmt.Printf("%v\n", otherServices["comment"])
}
