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
