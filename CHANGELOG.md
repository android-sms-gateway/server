# Changelog

Notable changes to SMSGate Server.

Release notes for published versions are also available on the
[GitHub Releases page](https://github.com/android-sms-gateway/server/releases).

## [Unreleased]

### New Features

#### Devices

- **Device public key** — device objects returned by `GET /api/3rdparty/v1/devices` now include two
  optional fields, `publicKey` and `keyVersion`. `publicKey` holds a base64-encoded RSA public key and
  `keyVersion` tracks key rotation. Both are absent for devices without a registered key.

  Keys are registered and rotated by the Android app through `POST`/`PATCH /api/mobile/v1/device`.
  Sending a new `publicKey` together with a `keyVersion` overwrites the previously stored key;
  clearing a registered key is not supported.

#### Messages

- **Longer phone numbers** — each entry in `phoneNumbers` now accepts up to 512 characters, up from
  the previous limit of 128. The stored recipient number column was widened to match.

  ```json
  {
    "phoneNumbers": ["<up to 512 characters>"]
  }
  ```