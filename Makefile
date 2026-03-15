APP_NAME=PiNetMonitor
PREFIX?=/opt/pinetmonitor

.PHONY: frontend-build backend-build build clean

frontend-build:
	cd web && npm run build

backend-build:
	go build -o bin/pinetmonitord ./cmd/pinetmonitord
	go build -o bin/pinetmonitor ./cmd/pinetmonitor

build: frontend-build backend-build

clean:
	rm -rf bin web/dist
