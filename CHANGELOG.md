# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2026-04-01

### Changed

- `RequestLogger` now reads the clog logger from the request context instead of a constructor argument, enabling per-request logging context (trace IDs, etc.)

## [0.2.0] - 2026-04-01

### Added

- `Middleware` type and `App.With()` for registering HTTP middleware
- `RequestLogger` middleware for structured request logging via clog

### Fixed

- Code generator (`hwgen`) now produces deterministic output by sorting definitions alphabetically

## [0.1.1] - 2026-04-01

### Fixed

- Empty or malformed request bodies now return 400 instead of silently producing an empty 200 response
- Deserialization errors are properly wrapped as `HeartwoodError` with status 400

## [0.1.0] - 2026-04-01

### Added

- Core handler framework with generic, type-safe request/response processing
- `Serializable` and `Deserializable` interfaces for pluggable encoding
- `Handler` generic type for strongly-typed endpoint functions
- `App` for registering and dispatching handlers
- HTTP integration via `NewServeMux` and `ListenAndServe`
- `HeartwoodError` for structured errors with HTTP status codes
- `ClientError` for client-side error deserialization
- Schema definition package (`pkg/schema`) with typed field builders and validation constraints
- Code generator (`cmd/hwgen`) that produces server handlers and typed HTTP clients from schema definitions
- CI pipeline with linting and 85%+ test coverage enforcement
