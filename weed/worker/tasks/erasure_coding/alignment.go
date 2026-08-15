package erasure_coding

import (
	"time"

	ecstorage "github.com/seaweedfs/seaweedfs/weed/storage/erasure_coding"
)

const maxAlignedSmallTailPercent uint64 = 1

// isLargeBlockAligned reports whether at most one percent of the volume would
// be encoded with small blocks. This is intentionally one-sided: a volume just
// below a large-row boundary has almost a full row of small blocks and is not
// aligned, while a volume just above it has only a small tail. The threshold
// is relative to the whole volume, so very large volumes can treat even a
// sizable absolute tail as negligible.
func isLargeBlockAligned(volumeSize uint64) bool {
	if volumeSize == 0 {
		return false
	}

	largeRowSize := uint64(ecstorage.DataShardsCount) * uint64(ecstorage.ErasureCodingLargeBlockSize)
	smallTailSize := volumeSize % largeRowSize
	return smallTailSize*100 <= volumeSize*maxAlignedSmallTailPercent
}

// canReachNextLargeBlockRow reports whether the configured volume size limit
// leaves room to cross the next large-block row boundary. An unknown limit
// cannot prove that waiting will help, so it uses the normal quiet period.
func canReachNextLargeBlockRow(volumeSize, volumeSizeLimit uint64) bool {
	if volumeSizeLimit == 0 {
		return false
	}
	largeRowSize := uint64(ecstorage.DataShardsCount) * uint64(ecstorage.ErasureCodingLargeBlockSize)
	return volumeSize/largeRowSize < volumeSizeLimit/largeRowSize
}

func requiredQuietPeriod(volumeSize, volumeSizeLimit uint64, fullnessRatio float64, readOnly, hasExistingECShards bool, config *Config) (time.Duration, bool) {
	quietPeriod := time.Duration(config.QuietForSeconds) * time.Second
	aligned := isLargeBlockAligned(volumeSize)
	// Waiting can only improve alignment while the volume can still receive
	// writes and the next row boundary is below its size limit. Leftover shards
	// need prompt recovery or cleanup rather than more source-volume writes.
	if aligned || fullnessRatio >= 1 || readOnly || hasExistingECShards || !canReachNextLargeBlockRow(volumeSize, volumeSizeLimit) {
		return quietPeriod, aligned
	}

	unalignedQuietPeriod := time.Duration(config.UnalignedQuietForSeconds) * time.Second
	if unalignedQuietPeriod < quietPeriod {
		unalignedQuietPeriod = quietPeriod
	}
	return unalignedQuietPeriod, false
}
