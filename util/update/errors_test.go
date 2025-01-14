package update_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/update"
)

// Just doing this to get coverage up on `errors.go`
func TestErrors(t *testing.T) {
	t.Run("NoTagsError", func(t *testing.T) {
		assert.Equal(t, "sup", update.NoTagsError("sup").Error())
	})

	t.Run("NoValidTagsError", func(t *testing.T) {
		assert.Equal(t, "sup", update.NoValidTagsError("sup").Error())
	})

	t.Run("InvalidSemverError", func(t *testing.T) {
		assert.Equal(t, "invalid semver: sup", update.InvalidSemverError("sup").Error())
		assert.Equal(t, "sup", update.InvalidSemverError("sup").Value())
	})
}
