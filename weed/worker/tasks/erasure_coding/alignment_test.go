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
		{name: "largest tail within one percent", size: 30*gib + 325376310, want: true},
		{name: "smallest tail above one percent", size: 30*gib + 325376311, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isLargeBlockAligned(tt.size))
		})
	}
}

func TestCanReachNextLargeBlockRow(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)
	const mib = uint64(1024 * 1024)

	assert.True(t, canReachNextLargeBlockRow(30*gib-1, 30*gib))
	assert.False(t, canReachNextLargeBlockRow(30*gib-1, 30000*mib), "the old 30,000 MiB limit cannot reach 30 GiB")
	assert.False(t, canReachNextLargeBlockRow(8*gib, 9*gib), "a limit below the first row cannot reach it")
	assert.False(t, canReachNextLargeBlockRow(30*gib-1, 0), "an unknown limit cannot prove that waiting will help")
}

func TestRequiredQuietPeriod(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)
	config := NewDefaultConfig()

	quiet, aligned := requiredQuietPeriod(30*gib, 40*gib, 0.96, false, false, config)
	assert.True(t, aligned)
	assert.Equal(t, time.Hour, quiet)

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 0.96, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, 72*time.Hour, quiet)

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 0.9999, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, 72*time.Hour, quiet, "a volume just below full can still improve its alignment")

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 1, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, time.Hour, quiet, "a full volume cannot improve its alignment")

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 1.0001, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, time.Hour, quiet, "an overfilled volume cannot improve its alignment")

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 0.96, true, false, config)
	assert.False(t, aligned)
	assert.Equal(t, time.Hour, quiet, "a read-only volume cannot improve its alignment")

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 0.96, false, true, config)
	assert.False(t, aligned)
	assert.Equal(t, time.Hour, quiet, "leftover EC shards need recovery rather than more writes")

	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib-1, 0.96, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, time.Hour, quiet, "waiting cannot reach a row beyond the volume limit")

	quiet, aligned = requiredQuietPeriod(30*gib-1, 0, 0.96, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, time.Hour, quiet, "an unknown limit must not introduce a speculative delay")

	config.QuietForSeconds = 96 * 60 * 60
	quiet, aligned = requiredQuietPeriod(30*gib-1, 30*gib, 0.96, false, false, config)
	assert.False(t, aligned)
	assert.Equal(t, 96*time.Hour, quiet, "unaligned policy must not shorten the normal quiet period")
}
