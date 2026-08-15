package erasure_coding

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsLargeBlockAligned(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)

	tests := []struct {
		name string
		size uint64
		want bool
	}{
		{name: "empty", size: 0, want: false},
		{name: "28.5 GiB has an 8.5 GiB tail", size: 28*gib + gib/2, want: false},
		{name: "just below 30 GiB is not close enough", size: 30*gib - 1, want: false},
		{name: "exactly 30 GiB", size: 30 * gib, want: true},
		{name: "30.2 GiB has less than one percent tail", size: 30*gib + gib/5, want: true},
		{name: "30.5 GiB has more than one percent tail", size: 30*gib + gib/2, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isLargeBlockAligned(tt.size))
		})
	}
}

func TestRequiredQuietPeriod(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)
	config := NewDefaultConfig()

	quiet, aligned := requiredQuietPeriod(30*gib, config)
	assert.True(t, aligned)
	assert.Equal(t, time.Hour, quiet)

	quiet, aligned = requiredQuietPeriod(30*gib-1, config)
	assert.False(t, aligned)
	assert.Equal(t, 72*time.Hour, quiet)

	config.QuietForSeconds = 96 * 60 * 60
	quiet, aligned = requiredQuietPeriod(30*gib-1, config)
	assert.False(t, aligned)
	assert.Equal(t, 96*time.Hour, quiet, "unaligned policy must not shorten the normal quiet period")
}
