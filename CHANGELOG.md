# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### New Features

#### Message status

- **`createdAt` in message status** — message state responses now report when the message was created, as an RFC 3339 timestamp. Returned by `POST /api/3rdparty/v1/messages`, `GET /api/3rdparty/v1/messages`, `GET /api/3rdparty/v1/messages/{id}`, and `DELETE /api/3rdparty/v1/messages/{id}`.

#### Phone number validation

- **Phone number length limit raised to 512 characters** — recipients in `phoneNumbers` on send requests are no longer rejected past 128 characters at request validation. Note that `message_recipients.phone_number` is `varchar(128)`, so un-hashed recipients longer than 128 characters (encrypted requests, or requests with phone validation skipped) are still rejected by the database.

## [1.49.1] - 2026-10-09

### Bug Fixes

- **Gateway mode validation** — `gateway.mode` now accepts only `public` or `private`; any other value fails at startup with a clear error instead of silently defaulting to public-mode FCM push.

## [1.48.0] - 2026-09-25

### New Features

- **Encrypted inbox ingestion** — Devices can upload batches of encrypted SMS, data SMS, and MMS messages, including attachments, to `POST /api/mobile/v1/inbox`.
- **Inbox listing and attachment access** — API users can filter and paginate `GET /api/3rdparty/v1/inbox` with the `inbox:list` scope, request MMS attachment metadata, and download individual attachments with `inbox:read`.
- **Inbox refresh and retention** — Users can request inbox synchronization with `POST /api/3rdparty/v1/inbox/refresh`, while the background worker removes inbox messages and attachments after the configured `TASKS__INBOX_CLEANUP__MAX_AGE` period, which defaults to 30 days.