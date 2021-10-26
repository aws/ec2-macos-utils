package diskutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinimumGrowSpaceError_Error(t *testing.T) {
	const expectedSize uint64 = 0

	e := FreeSpaceError{
		freeSpaceBytes: expectedSize,
	}

	expectedErrorMessage := fmt.Sprintf("%d bytes available", 0)

	actualErrorMessage := e.Error()

	assert.Equal(t, expectedErrorMessage, actualErrorMessage, "expected message to include metadata")
}
