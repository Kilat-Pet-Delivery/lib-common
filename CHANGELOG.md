# Changelog

All notable changes to lib-common are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `kafka.Producer`: `getWriter` read and wrote the `writers` map with no
  synchronisation, while `Publish` is reached from HTTP handlers. Concurrent
  publishes could trigger Go's `fatal error: concurrent map writes`, which kills
  the process and cannot be recovered. Guarded with a `sync.RWMutex`, with a
  double-check on the miss path so racing goroutines share one writer.
  `Close` takes the same lock. (KPD-66)
- `kafka.Producer`: writers now set `AllowAutoTopicCreation`. Without it kafka-go
  refuses to publish to a topic that does not exist yet, even when the broker has
  auto-creation enabled, because creation is client-driven. Consumers create
  topics on subscribe and producers did not, so only topics somebody happened to
  consume ever existed and everything else was silently dropped. (KPD-64)

  Partial fix: this makes the topic appear, but the publish that triggers
  creation still fails. Pre-creating the topics is the complete answer -- see
  KPD-67.

### Added

- `kafka/producer_test.go`: the package had no tests. Covers the concurrency
  regression (reproduces `concurrent map writes` against the old code without
  needing `-race`, which cannot run on the Windows dev laptop), writer reuse per
  topic, and the `AllowAutoTopicCreation` setting.
- `CHANGELOG.md`: this file. Partially advances KPD-52.
