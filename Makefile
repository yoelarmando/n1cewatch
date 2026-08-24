.PHONY: build clean install test docker

BIN=bin/n1cewatch
GOFLAGS=-trimpath
LDFLAGS=-s -w

build:
	@echo "[*] Building N1ceWatch (all-Ubuntu compatible)..."
	@mkdir -p bin
	go vet ./... || echo "[WARN] vet failed, continuing"
	CGO_ENABLED=1 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/agent 2>&1 || \
	( echo "[WARN] CGO build failed (missing linux-headers/libbpf), trying static fallback..." && $(MAKE) build-static )
	@echo "[OK] $(BIN) built"

build-static:
	@echo "[*] Static build (no eBPF, auditd -> /proc fallback for Kali/16.04)..."
	@mkdir -p bin
	CGO_ENABLED=0 go build -tags without_ebpf -o $(BIN) ./cmd/agent
	@echo "[OK] $(BIN) static built (fallback mode)"

install: build
	sudo mkdir -p /opt/n1cewatch/bin /var/log/n1cewatch
	sudo cp $(BIN) /opt/n1cewatch/bin/n1cewatch
	sudo chmod 0750 /opt/n1cewatch/bin/n1cewatch
	sudo cp -r packs /opt/n1cewatch/
	sudo cp deploy/systemd/n1cewatch.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable --now n1cewatch

clean:
	rm -rf bin/ frontend/dist/

test:
	go test ./... -v

# Test on all Ubuntu via Docker (requires Docker)
docker:
	docker build -f deploy/docker/Dockerfile.ubuntu16 -t n1cewatch:16.04 .
	docker build -f deploy/docker/Dockerfile.ubuntu22 -t n1cewatch:22.04 .
	docker run --rm n1cewatch:22.04 --help

fmt:
	go fmt ./...
