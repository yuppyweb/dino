package dino_test

import (
	"testing"

	"github.com/yuppyweb/dino"
)

func TestContainer_Inject(t *testing.T) {
	t.Run("injects dependencies into struct fields with inject tag", func(t *testing.T) {
		c := dino.New()

		expected := "test value"
		_ = c.Singleton(func() string { return expected })

		type target struct {
			Value string `inject:""`
		}

		tgt := &target{}
		err := c.Inject(tgt)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tgt.Value != expected {
			t.Errorf("expected %q, got %q", expected, tgt.Value)
		}
	})

	t.Run("injects dependencies with matching tag", func(t *testing.T) {
		c := dino.New()

		_ = c.Factory("tag1", func() string { return "tagged" })
		_ = c.Singleton(func() string { return "untagged" })

		type target struct {
			Tagged   string `inject:"tag1"`
			Untagged string `inject:""`
		}

		tgt := &target{}
		err := c.Inject(tgt)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tgt.Tagged != "tagged" {
			t.Errorf("expected %q, got %q", "tagged", tgt.Tagged)
		}
		if tgt.Untagged != "untagged" {
			t.Errorf("expected %q, got %q", "untagged", tgt.Untagged)
		}
	})

	t.Run("injects nested struct dependencies", func(t *testing.T) {
		c := dino.New()

		_ = c.Singleton(func() string { return "value" })

		type nested struct {
			Value string `inject:""`
		}

		type target struct {
			Nested nested `inject:""`
		}

		tgt := &target{}
		err := c.Inject(tgt)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tgt.Nested.Value != "value" {
			t.Errorf("expected %q, got %q", "value", tgt.Nested.Value)
		}
	})

	t.Run("returns nil for non-struct types", func(t *testing.T) {
		c := dino.New()

		var str string
		err := c.Inject(&str)

		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("skips unexported fields", func(t *testing.T) {
		c := dino.New()

		_ = c.Singleton(func() string { return "value" })

		type target struct {
			exported   string `inject:""`
			Unexported string `inject:""`
		}

		tgt := &target{}
		err := c.Inject(tgt)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tgt.Unexported != "value" {
			t.Errorf("expected %q, got %q", "value", tgt.Unexported)
		}
		if tgt.exported != "" {
			t.Errorf("unexported field should remain empty")
		}
	})

	t.Run("returns error when factory returns error", func(t *testing.T) {
		c := dino.New()

		_ = c.Singleton(func() (string, error) {
			return "", &testError{msg: "factory error"}
		})

		type target struct {
			Value string `inject:""`
		}

		tgt := &target{}
		err := c.Inject(tgt)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
