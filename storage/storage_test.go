package storage

import (
	"testing"
)

func TestInit(t *testing.T) {
	Init(":memory:")
}
