package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const initSchema = `
CREATE TABLE IF NOT EXISTS interface_samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sample_time TEXT NOT NULL,
    interface_name TEXT NOT NULL,
    rx_bytes INTEGER NOT NULL,
    tx_bytes INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_interface_samples_time
    ON interface_samples (sample_time);

CREATE INDEX IF NOT EXISTS idx_interface_samples_iface_time
    ON interface_samples (interface_name, sample_time);
`

type Sample struct {
	Timestamp time.Time `json:"timestamp"`
	Interface string    `json:"interface"`
	RxBytes   int64     `json:"rx_bytes"`
	TxBytes   int64     `json:"tx_bytes"`
}

type ReportRow struct {
	Bucket     string `json:"bucket"`
	Interface  string `json:"interface"`
	TotalBytes int64  `json:"total_bytes"`
	RxBytes    int64  `json:"rx_bytes"`
	TxBytes    int64  `json:"tx_bytes"`
}

type UsageRow struct {
	Bucket     string `json:"bucket"`
	TotalBytes int64  `json:"total_bytes"`
	RxBytes    int64  `json:"rx_bytes"`
	TxBytes    int64  `json:"tx_bytes"`
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.Migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate() error {
	if _, err := s.db.Exec(initSchema); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}

func (s *Store) InsertSample(sample Sample) error {
	_, err := s.db.Exec(`
		INSERT INTO interface_samples (sample_time, interface_name, rx_bytes, tx_bytes)
		VALUES (?, ?, ?, ?)`,
		sample.Timestamp.UTC().Format(time.RFC3339),
		sample.Interface,
		sample.RxBytes,
		sample.TxBytes,
	)
	return err
}

func (s *Store) LatestSamples() ([]Sample, error) {
	rows, err := s.db.Query(`
		SELECT sample_time, interface_name, rx_bytes, tx_bytes
		FROM interface_samples
		WHERE id IN (
			SELECT MAX(id) FROM interface_samples GROUP BY interface_name
		)
		ORDER BY interface_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []Sample
	for rows.Next() {
		var sampleTime string
		var sample Sample
		if err := rows.Scan(&sampleTime, &sample.Interface, &sample.RxBytes, &sample.TxBytes); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339, sampleTime)
		if err != nil {
			return nil, err
		}
		sample.Timestamp = parsed
		samples = append(samples, sample)
	}
	return samples, rows.Err()
}

func (s *Store) AggregateDaily(limit int) ([]ReportRow, error) {
	return s.aggregate("%Y-%m-%d", limit)
}

func (s *Store) AggregateMonthly(limit int) ([]ReportRow, error) {
	return s.aggregate("%Y-%m", limit)
}

func (s *Store) AggregateDailyTotals(limit int) ([]UsageRow, error) {
	return s.aggregateTotals("%Y-%m-%d", limit)
}

func (s *Store) AggregateMonthlyTotals(limit int) ([]UsageRow, error) {
	return s.aggregateTotals("%Y-%m", limit)
}

func (s *Store) aggregate(format string, limit int) ([]ReportRow, error) {
	rows, err := s.db.Query(`
		SELECT
			strftime(?, sample_time) AS bucket,
			interface_name,
			MAX(rx_bytes) - MIN(rx_bytes) AS rx_delta,
			MAX(tx_bytes) - MIN(tx_bytes) AS tx_delta
		FROM interface_samples
		GROUP BY bucket, interface_name
		ORDER BY bucket DESC
		LIMIT ?`, format, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []ReportRow
	for rows.Next() {
		var row ReportRow
		if err := rows.Scan(&row.Bucket, &row.Interface, &row.RxBytes, &row.TxBytes); err != nil {
			return nil, err
		}
		row.TotalBytes = row.RxBytes + row.TxBytes
		reports = append(reports, row)
	}
	return reports, rows.Err()
}

func (s *Store) aggregateTotals(format string, limit int) ([]UsageRow, error) {
	rows, err := s.db.Query(`
		WITH per_interface AS (
			SELECT
				strftime(?, sample_time) AS bucket,
				interface_name,
				MAX(rx_bytes) - MIN(rx_bytes) AS rx_delta,
				MAX(tx_bytes) - MIN(tx_bytes) AS tx_delta
			FROM interface_samples
			GROUP BY bucket, interface_name
		)
		SELECT
			bucket,
			COALESCE(SUM(rx_delta), 0) AS rx_delta,
			COALESCE(SUM(tx_delta), 0) AS tx_delta
		FROM per_interface
		GROUP BY bucket
		ORDER BY bucket DESC
		LIMIT ?`, format, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []UsageRow
	for rows.Next() {
		var row UsageRow
		if err := rows.Scan(&row.Bucket, &row.RxBytes, &row.TxBytes); err != nil {
			return nil, err
		}
		row.TotalBytes = row.RxBytes + row.TxBytes
		reports = append(reports, row)
	}
	return reports, rows.Err()
}
