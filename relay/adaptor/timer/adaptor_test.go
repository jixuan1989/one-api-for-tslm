package timer

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/meta"
)

func TestGetRequestURL_Forecast(t *testing.T) {
	a := &Adaptor{}
	m := &meta.Meta{
		BaseURL:        "http://127.0.0.1:10810",
		RequestURLPath: "/api/v1/forecast",
	}
	url, err := a.GetRequestURL(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "http://127.0.0.1:10810/timer/api/v1/forecast" {
		t.Errorf("got %q, want %q", url, "http://127.0.0.1:10810/timer/api/v1/forecast")
	}
}

func TestGetRequestURL_HelloTimer(t *testing.T) {
	a := &Adaptor{}
	m := &meta.Meta{
		BaseURL:        "http://127.0.0.1:10810",
		RequestURLPath: "/api/v1/hello_timer?name=test",
	}
	url, err := a.GetRequestURL(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "http://127.0.0.1:10810/timer/api/v1/hello_timer" {
		t.Errorf("got %q, want %q", url, "http://127.0.0.1:10810/timer/api/v1/hello_timer")
	}
}

func TestGetRequestURL_V1Forecast(t *testing.T) {
	a := &Adaptor{}
	m := &meta.Meta{
		BaseURL:        "http://localhost:3000",
		RequestURLPath: "/v1/forecast",
	}
	url, err := a.GetRequestURL(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "http://localhost:3000/timer/api/v1/forecast" {
		t.Errorf("got %q, want %q", url, "http://localhost:3000/timer/api/v1/forecast")
	}
}

func TestGetModelList(t *testing.T) {
	a := &Adaptor{}
	models := a.GetModelList()
	if len(models) == 0 {
		t.Fatal("model list is empty")
	}
	found := false
	for _, m := range models {
		if m == "sundial" {
			found = true
		}
	}
	if !found {
		t.Error("sundial not found in model list")
	}
}

func TestGetChannelName(t *testing.T) {
	a := &Adaptor{}
	if a.GetChannelName() != "timer" {
		t.Errorf("got %q, want %q", a.GetChannelName(), "timer")
	}
}
