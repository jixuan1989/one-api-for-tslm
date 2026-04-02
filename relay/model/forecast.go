package model

// TimeSeriesData represents pandas DataFrame split format: {columns: [...], data: [[...], ...]}
type TimeSeriesData struct {
	Columns []string        `json:"columns"`
	Data    [][]interface{} `json:"data"`
}

// ForecastRequest mirrors timer-rest-service POST /api/v1/forecast
type ForecastRequest struct {
	Targets             []TimeSeriesData `json:"targets"`
	HistoryCovs         []TimeSeriesData `json:"history_covs,omitempty"`
	FutureCovs          []TimeSeriesData `json:"future_covs,omitempty"`
	ModelID             *string          `json:"model_id,omitempty"`
	OutputLengthList    []int            `json:"output_length_list,omitempty"`
	OutputStartTimeList []string         `json:"output_start_time_list,omitempty"`
	OutputIntervalList  []string         `json:"output_interval_list,omitempty"`
	TimeColList         []string         `json:"time_col_list,omitempty"`
}

// ForecastResult holds one forecast task result
type ForecastResult struct {
	Results []TimeSeriesData `json:"results"`
}

// ForecastResponse mirrors timer-rest-service forecast response
type ForecastResponse struct {
	Code        int                    `json:"code"`
	Message     string                 `json:"message"`
	ServiceInfo map[string]interface{} `json:"service_info,omitempty"`
	Data        ForecastResult         `json:"data"`
}
