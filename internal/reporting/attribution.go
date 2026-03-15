package reporting

type ServiceShareRow struct {
	Service    string  `json:"service"`
	UsageBytes int64   `json:"usage_bytes"`
	Percentage float64 `json:"percentage"`
	Color      string  `json:"color"`
}

type ServiceShareSummary struct {
	Period        string            `json:"period"`
	DataAvailable bool              `json:"data_available"`
	Rows          []ServiceShareRow `json:"rows"`
}

func EmptyServiceShareSummary(period string) ServiceShareSummary {
	return ServiceShareSummary{
		Period:        period,
		DataAvailable: false,
		Rows:          []ServiceShareRow{},
	}
}
