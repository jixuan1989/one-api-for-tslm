package relaymode

import "testing"

func TestGetByPath_Forecast(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{"/timer/api/v1/forecast", Forecast},
		{"/timer/api/v1/hello_timer", Forecast},
		{"/timer/api/v1/hello_timer?name=test", Forecast},
		{"/v1/chat/completions", ChatCompletions},
		{"/v1/embeddings", Embeddings},
		{"/v1/images/generations", ImagesGenerations},
		{"/v1/completions", Completions},
		{"/v1/unknown", Unknown},
	}
	for _, tt := range tests {
		got := GetByPath(tt.path)
		if got != tt.want {
			t.Errorf("GetByPath(%q) = %d, want %d", tt.path, got, tt.want)
		}
	}
}
