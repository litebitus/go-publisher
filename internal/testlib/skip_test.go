package testlib

import (
	"testing"

	"github.com/litebitus/go-publisher/internal/pipe"
)

func TestAssertSkipped(t *testing.T) {
	AssertSkipped(t, pipe.Skip("skip"))
}
