.DEFAULT_GOAL := build

fmt:
	cargo fmt
.PHONY:fmt

# `rustup component add clippy`
lint:
	cargo clippy
.PHONY: lint

build: fmt lint
	cargo build
.PHONY: build

release: fmt lint
	cargo build --release
.PHONY: release

run: fmt lint
	# it requires `cargo-watch` via `make deps`
	cargo watch -x 'run'
.PHONY: run

clean:
	cargo clean
.PHONY: clean

doc:
	cargo rustdoc
.PHONY: doc

test:
	RUST_BACKTRACE=full cargo test
.PHONY: test

# add tests coverage
# check https://doc.rust-lang.org/rustc/instrument-coverage.html

deps:
	rustup update
	rustup component add clippy
	rustup component add rustfmt
	cargo update
	cargo install cargo-watch
.PHONY: deps

deps-ci:
	rustup component add clippy
	rustup component add rustfmt
.PHONY: deps-ci