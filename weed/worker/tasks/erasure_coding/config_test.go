package erasure_coding

import (
	"testing"

	"github.com/seaweedfs/seaweedfs/weed/pb/worker_pb"
	"github.com/stretchr/testify/require"
)

func TestConfigPersistsUnalignedQuietPeriod(t *testing.T) {
	config := NewDefaultConfig()
	config.UnalignedQuietForSeconds = 4 * 24 * 60 * 60

	policy := config.ToTaskPolicy()
	require.Equal(t, int32(4*24*60*60), policy.GetErasureCodingConfig().GetUnalignedQuietForSeconds())

	loaded := NewDefaultConfig()
	require.NoError(t, loaded.FromTaskPolicy(policy))
	require.Equal(t, config.UnalignedQuietForSeconds, loaded.UnalignedQuietForSeconds)
}

func TestConfigDefaultsUnalignedQuietPeriodForLegacyPolicy(t *testing.T) {
	legacyPolicy := &worker_pb.TaskPolicy{
		TaskConfig: &worker_pb.TaskPolicy_ErasureCodingConfig{
			ErasureCodingConfig: &worker_pb.ErasureCodingTaskConfig{
				QuietForSeconds: 3600,
			},
		},
	}

	var loaded Config
	require.NoError(t, loaded.FromTaskPolicy(legacyPolicy))
	require.Equal(t, DefaultUnalignedQuietForSeconds, loaded.UnalignedQuietForSeconds)
}

func TestConfigPersistsUnalignedQuietPeriodAtLeastNormalQuietPeriod(t *testing.T) {
	config := NewDefaultConfig()
	config.QuietForSeconds = 2 * 60 * 60
	config.UnalignedQuietForSeconds = 0

	policy := config.ToTaskPolicy()
	require.Equal(t, int32(config.QuietForSeconds), policy.GetErasureCodingConfig().GetUnalignedQuietForSeconds())

	loaded := NewDefaultConfig()
	require.NoError(t, loaded.FromTaskPolicy(policy))
	require.Equal(t, config.QuietForSeconds, loaded.UnalignedQuietForSeconds)
}
