package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertMonthNameFormatsIndonesianMonthWithoutExtraSpace(t *testing.T) {
	result := ConvertMonthName("2026-05-22")

	assert.Equal(t, "22 Mei", result)
}
