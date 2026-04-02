package channeltype

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/apitype"
)

func TestToAPIType_TimerRestService(t *testing.T) {
	got := ToAPIType(TimerRestService)
	if got != apitype.TimerRestService {
		t.Errorf("ToAPIType(TimerRestService) = %d, want %d", got, apitype.TimerRestService)
	}
}

func TestToAPIType_ExistingTypes(t *testing.T) {
	// Ensure existing mappings are not broken
	tests := []struct {
		channel int
		want    int
	}{
		{Anthropic, apitype.Anthropic},
		{Gemini, apitype.Gemini},
		{Ollama, apitype.Ollama},
		{OpenAI, apitype.OpenAI}, // default
	}
	for _, tt := range tests {
		got := ToAPIType(tt.channel)
		if got != tt.want {
			t.Errorf("ToAPIType(%d) = %d, want %d", tt.channel, got, tt.want)
		}
	}
}

func TestChannelBaseURLs_TimerRestService(t *testing.T) {
	if TimerRestService >= len(ChannelBaseURLs) {
		t.Fatalf("TimerRestService index %d out of ChannelBaseURLs range %d", TimerRestService, len(ChannelBaseURLs))
	}
	url := ChannelBaseURLs[TimerRestService]
	if url != "http://127.0.0.1:10810" {
		t.Errorf("ChannelBaseURLs[TimerRestService] = %q, want %q", url, "http://127.0.0.1:10810")
	}
}
