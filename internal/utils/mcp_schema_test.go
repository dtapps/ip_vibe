package utils

import (
	"testing"
	"vibe_ip/internal/types"
)

func TestAutoSchema(t *testing.T) {
	t.Log("types.GetCdnRecordsRequest{}", AutoSchema(types.GetCdnRecordsRequest{}))
	t.Log("types.GetCdnRecordResponse{}", AutoSchema(types.GetCdnRecordResponse{}))
	t.Log("types.GetCdnRecordRequest{}", AutoSchema(types.GetCdnRecordRequest{}))
	t.Log("types.CreateCdnRecordRequest{}", AutoSchema(types.CreateCdnRecordRequest{}))
	t.Log("types.UpdateCdnRecordRequest{}", AutoSchema(types.UpdateCdnRecordRequest{}))
	t.Log("types.UpdateCdnRecordActiveRequest{}", AutoSchema(types.UpdateCdnRecordActiveRequest{}))
	t.Log("types.DeleteCdnRecordRequest{}", AutoSchema(types.DeleteCdnRecordRequest{}))
	t.Log("types.HealthCheckCdnRecordRequest{}", AutoSchema(types.HealthCheckCdnRecordRequest{}))
}
