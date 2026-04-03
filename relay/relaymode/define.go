package relaymode

const (
	Unknown = iota
	ChatCompletions
	Completions
	Embeddings
	Moderations
	ImagesGenerations
	Edits
	AudioSpeech
	AudioTranscription
	AudioTranslation
	// Proxy is a special relay mode for proxying requests to custom upstream
	Proxy
	// Forecast is for time series forecasting via Timer REST Service
	Forecast
	// SaasBusiness is for business requests forwarded to SaaS Backend
	SaasBusiness
)
