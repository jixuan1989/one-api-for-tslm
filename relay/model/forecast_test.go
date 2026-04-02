package model

import (
	"encoding/json"
	"testing"
)

func TestForecastRequest_Serialize(t *testing.T) {
	modelId := "sundial"
	req := ForecastRequest{
		Targets: []TimeSeriesData{
			{Columns: []string{"value"}, Data: [][]interface{}{{1.0}, {2.0}}},
		},
		ModelID:          &modelId,
		OutputLengthList: []int{5},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if parsed["model_id"] != "sundial" {
		t.Errorf("model_id = %v, want sundial", parsed["model_id"])
	}
	targets := parsed["targets"].([]interface{})
	if len(targets) != 1 {
		t.Errorf("targets len = %d, want 1", len(targets))
	}
}

func TestForecastRequest_OmitEmpty(t *testing.T) {
	req := ForecastRequest{
		Targets: []TimeSeriesData{
			{Columns: []string{"value"}, Data: [][]interface{}{{1.0}}},
		},
	}
	data, _ := json.Marshal(req)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if _, ok := parsed["model_id"]; ok {
		t.Error("model_id should be omitted when nil")
	}
	if _, ok := parsed["history_covs"]; ok {
		t.Error("history_covs should be omitted when empty")
	}
}

func TestForecastResponse_Deserialize(t *testing.T) {
	raw := `{"code":200,"message":"ok","data":{"results":[{"columns":["time","value"],"data":[["2024-01-01",1.5],["2024-01-02",2.5]]}]}}`
	var resp ForecastResponse
	err := json.Unmarshal([]byte(raw), &resp)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if resp.Code != 200 {
		t.Errorf("code = %d, want 200", resp.Code)
	}
	if len(resp.Data.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(resp.Data.Results))
	}
	if len(resp.Data.Results[0].Data) != 2 {
		t.Errorf("data rows = %d, want 2", len(resp.Data.Results[0].Data))
	}
	if resp.Data.Results[0].Columns[0] != "time" {
		t.Errorf("first column = %q, want time", resp.Data.Results[0].Columns[0])
	}
}
