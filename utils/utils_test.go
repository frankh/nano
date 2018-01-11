package utils

import (
	"bytes"
	"testing"
)

func TestEmpty(t *testing.T) {
	str := []byte{}

	if !bytes.Equal(Reversed(str), []byte{}) {
		t.Errorf("Failed on empty slice")
	}
}

func TestSingle(t *testing.T) {
	str := []byte{1}

	if !bytes.Equal(Reversed(str), []byte{1}) {
		t.Errorf("Failed on slice with single element")
	}
}

func TestReverseOrder(t *testing.T) {
	str := []byte{1, 2, 3}

	if !bytes.Equal(Reversed(str), []byte{3, 2, 1}) {
		t.Errorf("Failed on slice with single element")
	}
}

func TestReverseUnordered(t *testing.T) {
	str := []byte{1, 2, 1, 3, 1}

	if !bytes.Equal(Reversed(str), []byte{1, 3, 1, 2, 1}) {
		t.Errorf("Failed on slice with single element")
	}
}
