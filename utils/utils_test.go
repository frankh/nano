package utils

import (
	"bytes"
	"testing"
)

func TestEmpty(t *testing.T) {
	byteArray := []byte{}

	if !bytes.Equal(Reversed(byteArray), []byte{}) {
		t.Errorf("Failed on empty slice")
	}
}

func TestSingle(t *testing.T) {
	byteArray := []byte{1}

	if !bytes.Equal(Reversed(byteArray), []byte{1}) {
		t.Errorf("Failed on slice with single element")
	}
}

func TestReverseOrder(t *testing.T) {
	byteArray := []byte{1, 2, 3}

	if !bytes.Equal(Reversed(byteArray), []byte{3, 2, 1}) {
		t.Errorf("Failed on slice with single element")
	}
}

func TestReverseUnordered(t *testing.T) {
	byteArray := []byte{1, 2, 1, 3, 1}

	if !bytes.Equal(Reversed(byteArray), []byte{1, 3, 1, 2, 1}) {
		t.Errorf("Failed on slice with single element")
	}
}
