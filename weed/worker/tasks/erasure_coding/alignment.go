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

func requiredQuietPeriod(volumeSize uint64, config *Config) (time.Duration, bool) {
	quietPeriod := time.Duration(config.QuietForSeconds) * time.Second
	aligned := isLargeBlockAligned(volumeSize)
	if aligned {
		return quietPeriod, true
	}

	unalignedQuietPeriod := time.Duration(config.UnalignedQuietForSeconds) * time.Second
	if unalignedQuietPeriod < quietPeriod {
		unalignedQuietPeriod = quietPeriod
	}
	return unalignedQuietPeriod, false
}
