# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### New Features

#### Message status

- **`scheduleAt` in message status** — message state responses now report when the message is scheduled for delivery, as an RFC 3339 timestamp. The field is returned only for scheduled messages and is omitted for all others. Returned by `POST /api/3rdparty/v1/messages`, `GET /api/3rdparty/v1/messages`, `GET /api/3rdparty/v1/messages/{id}`, and `DELETE /api/3rdparty/v1/messages/{id}`.

### Bug Fixes

- **Message list pagination** — `GET /api/3rdparty/v1/messages` now applies the `offset` query parameter, which was previously accepted but ignored, so the second and later pages of a listing can be fetched. The `limit` parameter is now capped at 100 (default 50).

## [1.51.1] - 2026-10-21

### Maintenance

- **Dependency updates** — the OpenTelemetry Go SDK and its instrumentation packages were upgraded from v1.44.0 to v1.46.0 (contrib packages v0.69.0 to v0.71.0), together with updates to the Firebase, Fiber, Redis and MySQL driver libraries.

## [1.51.0] - 2026-10-19

### New Features

#### Device encryption keys

- **Device public key and key version** — `POST /api/mobile/v1/device` and `PATCH /api/mobile/v1/device` now accept `publicKey` (base64-encoded RSA public key) and `keyVersion`, and both are returned by `GET /api/mobile/v1/device` and `GET /api/3rdparty/v1/devices`. Sending a new key with a version overwrites the stored key, so devices can rotate keys; clearing a stored key is not supported. Requires migration `20261002000000` (`make db-upgrade`), which adds the `public_key` and `key_version` columns.

#### Phone number validation

- **Phone number length limit raised to 512 characters** — recipients in `phoneNumbers` on send requests are no longer rejected past 128 characters at request validation. The same migration widens `message_recipients.phone_number` to `varchar(512)`; until `make db-upgrade` is run, un-hashed recipients longer than 128 characters (encrypted requests, or requests with phone validation skipped) are still rejected by the database.

## [1.50.0] - 2026-10-16

### New Features

#### Message status

- **`createdAt` in message status** — message state responses now report when the message was created, as an RFC 3339 timestamp. Returned by `POST /api/3rdparty/v1/messages`, `GET /api/3rdparty/v1/messages`, `GET /api/3rdparty/v1/messages/{id}`, and `DELETE /api/3rdparty/v1/messages/{id}`.

### Bug Fixes

- **Clearing a device push token** — `PATCH /api/mobile/v1/device` with `"pushToken": ""` now clears the stored FCM token. Previously the empty value was treated as "no change", so the device kept its old token and continued receiving pushes.

## [1.49.0] - 2026-10-02

### New Features

- **Outgoing MMS** — send requests now accept an `mmsMessage` with an optional subject, an optional text body and a list of attachments, each carrying its MIME type, an optional file name and base64-encoded content. The stored message is returned as `mmsMessage` in status responses when `includeContent=true`. The request shape is provided by `client-go` v1.16.0. Requires migration `20260925000000` (`make db-upgrade`), which widens the `messages.content` column and adds the `Mms` message type.

## [1.49.1] - 2026-10-09

### Bug Fixes

- **Gateway mode validation** — `gateway.mode` now accepts only `public` or `private`; any other value fails at startup with a clear error instead of silently defaulting to public-mode FCM push.

## [1.48.0] - 2026-09-25

### New Features

- **Encrypted inbox ingestion** — Devices can upload batches of encrypted SMS, data SMS, and MMS messages, including attachments, to `POST /api/mobile/v1/inbox`.
- **Inbox listing and attachment access** — API users can filter and paginate `GET /api/3rdparty/v1/inbox` with the `inbox:list` scope, request MMS attachment metadata, and download individual attachments with `inbox:read`.
- **Inbox refresh and retention** — Users can request inbox synchronization with `POST /api/3rdparty/v1/inbox/refresh`, while the background worker removes inbox messages and attachments after the configured `TASKS__INBOX_CLEANUP__MAX_AGE` period, which defaults to 30 days.