package errors

import (
	"errors"
	"testing"
)

func TestItCreatesACodedError(t *testing.T) {
	err, ok := Code(errors.New("source"), "test").(*CodedError)
	if !ok {
		t.Fatal("function did not return an instance of coded error")
	}

	if err.Code != "test" {
		t.Error("expected error code test, got", err.Code)
	}

	if err.Error() != "source" {
		t.Error("expected error source, got", err.Error())
	}
}

func TestItReturnsNilWhenANilErrorIsPassed(t *testing.T) {
	if Code(nil, "noop") != nil {
		t.Error("non-nil value returned")
	}
}
