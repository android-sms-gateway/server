# Changelog

All notable changes to SMSGate Server are documented in this file.

## [1.48.0] - 2026-09-25

### New Features

- **Encrypted inbox ingestion** — Devices can upload batches of encrypted SMS, data SMS, and MMS messages, including attachments, to `POST /api/mobile/v1/inbox`.
- **Inbox listing and attachment access** — API users can filter and paginate `GET /api/3rdparty/v1/inbox` with the `inbox:list` scope, request MMS attachment metadata, and download individual attachments with `inbox:read`.
- **Inbox refresh and retention** — Users can request inbox synchronization with `POST /api/3rdparty/v1/inbox/refresh`, while the background worker removes inbox messages and attachments after the configured `TASKS__INBOX_CLEANUP__MAX_AGE` period, which defaults to 30 days.
