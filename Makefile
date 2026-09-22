SHELL := /bin/bash

.PHONY: dev run

run:
	cd backend && go run ./cmd/server

dev:
	@trap 'kill 0' EXIT INT TERM; \
	(cd frontend && npm run dev) & \
	$(MAKE) run & \
	wait