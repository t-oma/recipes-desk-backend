# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **rabbitmq**: E2E integration tests for publish→consume round-trip flow
- **rabbitmq**: DLQ integration test verifying failed messages route to dead-letter queue

## [v0.0.0] - 2026-03-21

### Fixed

- **rabbitmq**: Race condition in `handleReconnect` - connection reference now captured under lock before use (#17)
- **rabbitmq**: Double reconnect bug - `handleReconnect` now starts once in `NewConnection` instead of `connect()` (#17)
- **rabbitmq**: Added `NotifyBlocked` handler for server blocked/unblocked events (#17)
- **rabbitmq**: Removed infinite retry loop - failed messages now go to DLQ instead of requeue (#19)
- **rabbitmq**: Added graceful shutdown with `WaitGroup` - `Close()` waits for in-flight messages (#19)
- **rabbitmq**: Added `SetHandlerTimeout()` for optional handler execution timeout (#19)
- **rabbitmq**: Removed duplicate `QueueDeclare` from Consumer - topology handles queue declaration (#19)

### Added

- **rabbitmq**: `Topology.SetupDLQ()` - creates DLQ infrastructure (exchange + queue + binding) with naming convention `{queue}.dlx`/`{queue}.dlq`
- **rabbitmq**: `Topology.SetupTopologyWithDLQ()` - convenience method for setting up topology with DLQ in one call
- **rabbitmq**: `ProducerConfig` - configuration struct for Producer with `ConfirmMode`, `Mandatory`, `ConfirmTimeout`
- **rabbitmq**: `DefaultProducerConfig()` - default producer configuration
- **rabbitmq**: Auto-generated message ID using nanoid if not provided
- **rabbitmq**: Auto-set message timestamp if zero
- **rabbitmq**: Proper mandatory flag handling with `NotifyReturn` - returns error if message cannot be routed

### Changed

- **rabbitmq**: `NewProducer` now accepts `ProducerConfig` instead of individual parameters
- **rabbitmq**: `NotifyPublish` is now registered before `Publish` to avoid race condition
- **rabbitmq**: Refactored `Producer.Publish()` - extracted helper methods for cleaner code
