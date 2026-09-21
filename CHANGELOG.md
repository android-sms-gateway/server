# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Bug Fixes

- **Gateway mode validation** — `gateway.mode` now accepts only `public` or `private`; any other value fails at startup with a clear error instead of silently defaulting to public-mode FCM push.