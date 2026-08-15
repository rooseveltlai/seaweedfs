package erasure_coding

import (
	"time"

	ecstorage "github.com/seaweedfs/seaweedfs/weed/storage/erasure_coding"
)

const maxAlignedSmallTailPercent uint64 = 1

// isLargeBlockAligned reports whether at most one percent of the volume would
// be encoded with small blocks. This is intentionally one-sided: a volume just
// below a large-row boundary has almost a full row of small blocks and is not
// aligned, while a volume just above it has only a small tail.
func isLargeBlockAligned(volumeSize uint64) bool {
	if volumeSize == 0 {
		return false
	}

	largeRowSize := uint64(ecstorage.DataShardsCount) * uint64(ecstorage.ErasureCodingLargeBlockSize)
	smallTailSize := volumeSize % largeRowSize
	return smallTailSize*100 <= volumeSize*maxAlignedSmallTailPercent
}

func requiredQuietPeriod(volumeSize uint64, fullnessRatio float64, readOnly bool, config *Config) (time.Duration, bool) {
	quietPeriod := time.Duration(config.QuietForSeconds) * time.Second
	aligned := isLargeBlockAligned(volumeSize)
	// Waiting can only improve alignment while the volume can still receive
	// writes. Full and explicitly read-only volumes have no such opportunity.
	if aligned || fullnessRatio >= 1 || readOnly {
		return quietPeriod, aligned
	}

	unalignedQuietPeriod := time.Duration(config.UnalignedQuietForSeconds) * time.Second
	if unalignedQuietPeriod < quietPeriod {
		unalignedQuietPeriod = quietPeriod
	}
	return unalignedQuietPeriod, false
}
