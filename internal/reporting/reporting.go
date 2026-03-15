package reporting

import "github.com/agraja38/PiNetMonitor/internal/store"

type Summary struct {
	Daily   []store.ReportRow `json:"daily"`
	Monthly []store.ReportRow `json:"monthly"`
}

func Build(db *store.Store) (Summary, error) {
	daily, err := db.AggregateDaily(14)
	if err != nil {
		return Summary{}, err
	}
	monthly, err := db.AggregateMonthly(12)
	if err != nil {
		return Summary{}, err
	}
	return Summary{Daily: daily, Monthly: monthly}, nil
}
