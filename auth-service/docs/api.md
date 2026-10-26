### Authentication contract

All paths below are relative to `/api/v1/auth`. Send JSON request bodies with `Content-Type: application/json`.

Protected operations require `Authorization: Bearer <access_token>`. Registration, login, guest login, refresh and logout do not require this header. Refresh and logout identify the session using `refresh_token` in the request body.

### Dates and response fields

`access_expires_at`, `created_at` and `updated_at` are RFC 3339 timestamps in UTC, ending in `Z`, for example `2026-10-25T10:00:00Z` or `2026-10-25T10:00:00.123456Z`. Fractional seconds are optional and may contain up to nine digits. The profile's `timezone` does not change these timestamps; it is a separate user preference.

`required` means a JSON key must be present. It does not mean the key can contain `null`. An optional key may be absent; a nullable value may explicitly be `null`. These are different properties of the contract.

Successful response objects always include the following keys with non-null values:

| Response | Always present |
| --- | --- |
| `AuthResponse` | `access_token`, `refresh_token`, `access_expires_at`, `user` |
| `TokenResponse` | `access_token`, `refresh_token`, `access_expires_at` |
| `UserResponse` | `id`, `nickname`, `timezone`, `is_guest`, `created_at`, `updated_at` |

`UserResponse.email` is a string when available. For a guest, the key is **omitted**, rather than sent as `null` or `""`. Handle the absent key when decoding a guest profile. Responses with status `204` have no body.

Example guest profile (the `user` object inside a guest-login response, or the body of `GET /profile`):

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "nickname": "Гость",
  "timezone": "UTC",
  "is_guest": true,
  "created_at": "2026-10-26T10:00:00Z",
  "updated_at": "2026-10-26T10:00:00Z"
}
```

After linking, the same profile ID is returned with `is_guest: false` and an `email` string. The other profile keys remain present.

### Request fields and defaults

| Operation | Input needed for the operation | Optional input and defaults |
| --- | --- | --- |
| `POST /register` | `email`, `password`, `nickname` | `timezone`: defaults to `UTC` when omitted or blank |
| `POST /login` | `email`, `password` | None |
| `POST /guest-login` | None; the body may be absent | `nickname`: defaults to `Гость` when omitted or blank; `timezone`: defaults to `UTC` when omitted or blank |
| `POST /refresh` | Nonempty `refresh_token` | None |
| `POST /logout` | Nonempty `refresh_token` | None |
| `POST /link-guest` | `email`, `password` | None |
| `PATCH /password` | `old_password`, `new_password` | None |
| `PATCH /email` | `new_email`, `password` for an actual change | Submitting the current normalized email is a no-op; the current implementation returns the profile before checking the password in this case |
| `PATCH /profile` | A JSON body; no individual field is mandatory | Omitted or blank `nickname` and `timezone` keep their current values; `{}` is accepted |
| `GET /profile` | No body | None |

Input rules:

- Email is trimmed, validated as an address without a display name and converted to lowercase. Its trimmed length must be at most 255 UTF-8 bytes.
- A new password, or a password submitted to `/login`, must contain at least eight Unicode code points and at most 72 UTF-8 bytes. Passwords are not trimmed. Multibyte characters count towards the byte limit. The current password supplied to a password/email change is checked against the stored hash.
- Nicknames are trimmed and must contain 2-32 Unicode code points after trimming. Guest login applies its default before this validation.
- Timezones are trimmed and must be supported IANA timezone names, such as `Europe/Moscow`, or `UTC`. Registration and guest login use `UTC` for blank values. A blank timezone in `PATCH /profile` keeps the existing timezone.

The current JSON decoder does not independently enforce schema `required` fields. Missing string fields, and string fields set to JSON `null`, decode as empty strings; service validation or the defaults above then apply. A top-level `null` also decodes as an empty request. This does **not** provide a way to clear a field. Malformed JSON or a number/array/object supplied for a string field returns `400 INVALID_REQUEST`. Unknown JSON fields are currently ignored. Clients should send one JSON object and use the documented fields and types.

### Error format

Errors returned by the auth application's error handler have this shape:

```json
{
  "code": "EMAIL_ALREADY_EXISTS",
  "status_code": 409,
  "message": "User with this email already exists"
}
```

All three keys are present. `status_code` matches the HTTP status. Internal causes and database details are not included in the JSON response.

Use `code` to select a localized UI message or action. Do not compare the English `message` text. Handle an unknown code with a generic fallback based on the HTTP status. Router-generated, proxy-generated and panic-recovery responses can have another body format; the client must also handle an empty or non-JSON error body.

| HTTP | Code | Meaning |
| --- | --- | --- |
| 400 | `INVALID_REQUEST` | The JSON request cannot be decoded, including a wrong value type or a body that exceeds the application's decoder limit |
| 400 | `INVALID_EMAIL` | The email is missing or has an invalid format/length |
| 400 | `INVALID_PASSWORD` | The submitted password does not meet the length rules |
| 400 | `INVALID_NICKNAME` | The nickname does not meet the length rules |
| 400 | `INVALID_TIMEZONE` | The timezone is not supported |
| 400 | `GUEST_ALREADY_LINKED` | The account is already a permanent account |
| 401 | `INVALID_CREDENTIALS` | Login credentials or the current password are incorrect; also returned when the account has no password |
| 401 | `UNAUTHORIZED` | The Bearer authorization header is missing or malformed |
| 401 | `INVALID_TOKEN` | The access token is invalid/expired, or the refresh token is missing, unknown or expired; see endpoint behavior below |
| 401 | `SESSION_REVOKED` | The refresh token has already been revoked, including a token consumed by successful rotation |
| 404 | `NOT_FOUND` | The requested user no longer exists or is deleted |
| 409 | `EMAIL_ALREADY_EXISTS` | Another active account uses this email |
| 400 | `VALIDATION_FAILED` | A database constraint rejected the data; a generic infrastructure fallback |
| 409 | `CONFLICT` | A database uniqueness constraint rejected the write without a more specific mapping |
| 500 | `INTERNAL_SERVER_ERROR` | An internal operation failed |