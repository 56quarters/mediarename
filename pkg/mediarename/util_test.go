package mediarename

import (
	"errors"
	"testing"
)

func RequireNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error\n%+v", err)
	}
}

func RequireErrorIs(t *testing.T, err, target error) {
	if !errors.Is(err, target) {
		t.Fatalf("expected error %q to contain target type %q", err, target)
	}
}

func RequireEqual[T comparable](t *testing.T, expected T, actual T) {
	if expected != actual {
		t.Fatalf("%+v is not equal to %+v", expected, actual)
	}
}
