# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Fixed

- **rabbitmq**: Race condition in `handleReconnect` - connection reference now captured under lock before use (#17)
- **rabbitmq**: Double reconnect bug - `handleReconnect` now starts once in `NewConnection` instead of `connect()` (#17)
- **rabbitmq**: Added `NotifyBlocked` handler for server blocked/unblocked events (#17)
