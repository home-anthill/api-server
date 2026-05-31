package utils

import (
	"errors"
	"regexp"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestRandomString(t *testing.T) {
	t.Run("returns base64url random text", func(t *testing.T) {
		got, err := RandomString(32)
		if err != nil {
			t.Fatalf("RandomString returned error: %v", err)
		}
		if len(got) != 43 {
			t.Fatalf("RandomString(32) length = %d, want 43", len(got))
		}
		if !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(got) {
			t.Fatalf("RandomString returned non base64url text: %q", got)
		}

		other, err := RandomString(32)
		if err != nil {
			t.Fatalf("RandomString returned error: %v", err)
		}
		if got == other {
			t.Fatalf("two random strings matched: %q", got)
		}
	})

	t.Run("rejects non positive length", func(t *testing.T) {
		for _, n := range []int{0, -1} {
			if got, err := RandomString(n); err == nil || got != "" {
				t.Fatalf("RandomString(%d) = %q, %v; want empty error", n, got, err)
			}
		}
	})
}

func TestGetErrorMessage(t *testing.T) {
	t.Run("returns validation field list", func(t *testing.T) {
		type request struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}
		err := validator.New().Struct(request{})
		if err == nil {
			t.Fatal("expected validation error")
		}

		got := GetErrorMessage(err)
		if got != " name email" {
			t.Fatalf("GetErrorMessage() = %q, want %q", got, " name email")
		}
	})

	t.Run("returns plain error message for non validation error", func(t *testing.T) {
		err := errors.New("plain failure")
		if got := GetErrorMessage(err); got != "plain failure" {
			t.Fatalf("GetErrorMessage() = %q, want plain failure", got)
		}
	})
}
