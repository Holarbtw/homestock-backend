### Inventory contract

The public API prefix is `/api/v1/inventory`. The routes below are relative to that prefix. When calling the inventory service directly, use `/items` or `/items/{item_id}`. Send JSON bodies with `Content-Type: application/json`.

All item operations require `Authorization: Bearer <access_token>`. Use the access token issued by auth. Inventory resolves the personal space from the authenticated user. Create lazily opens a space if needed. A user who has no space gets an empty list. An item from another space is treated as not found.

### Swagger access

Open `/api/v1/inventory/swagger/index.html` through the gateway. The VPS gateway requires HTTP Basic credentials for the Swagger UI, its assets and `doc.json`. In Swagger, use **Authorize** to set `Bearer <access_token>` for API requests. The local Nginx gateway serves Swagger without Basic authentication.

### Requests and successful responses

| Operation | Request | Successful response |
| --- | --- | --- |
| `POST /items` | `Idempotency-Key` header and a JSON object with `name`, `category`, `unit` and optional item fields or `first_batch` | `201` with `data.item`, `data.batches` and optional `data.operation` |
| `GET /items` | Optional `limit` and `offset` query parameters | `200` with a `data` array of items and a `meta` object |
| `GET /items/{item_id}` | Item UUID in the path | `200` with `data.item` and `data.batches` |
| `PATCH /items/{item_id}` | Item UUID and a JSON object with any of `name`, `brand`, `category`, `minimum_quantity` | `200` with the updated item in `data` |
| `DELETE /items/{item_id}` | Item UUID in the path | `204` with no body |

Deleting an item removes it from active inventory, even if stock remains. Its batches and operation history are preserved. Getting, editing or deleting that item again returns `404`.

An item response always includes `id`, `space_id`, `name`, `category`, `unit`, `total_quantity`, `is_low_stock`, `created_at` and `updated_at`. Optional fields such as `brand`, `minimum_quantity` and `image_url` are omitted when unset. The total is the sum of the remaining quantities across all batches, including expired batches. `is_low_stock` is true when a threshold is set and the total is at or below it.

When present, `batches` contains full batch objects, including any with zero remaining quantity. Empty arrays appear as `[]`. Create returns the first replenishment in `operation` only when the request included a first batch. Each operation line identifies the batch and the amount added to it.

### Input rules

- `name` and `brand` are trimmed. A provided string must be nonblank and contain no NUL characters. The limits are 200 Unicode code points for the name and 100 for the brand.
- `category` accepts `FOOD`, `MEDICINE` or `HOUSEHOLD`. `unit` accepts `ml`, `g` or `pcs`. The card's unit also applies to its batches and operations.
- Send quantities as nonnegative decimal strings with at most three fractional digits, for example `"930.000"`. `package_quantity` represents the size of one package. It must be positive when provided. `minimum_quantity` is an optional low-stock threshold and can be zero.
- `first_batch` is optional. When provided, it needs a positive `quantity`. Its `purchased_on` and `expires_on` dates are optional. The purchase date cannot be later than the current calendar date on the server. The expiry date can be in the past or the future.
- An optional `gtin` accepts 8, 12, 13 or 14 digits and is returned as 14 digits with leading zeroes if needed. An optional `catalog_product_id` must be a nonzero UUID. An optional `external_image_url` must be an absolute HTTP or HTTPS URL with a hostname.

In PATCH, a missing field keeps its current value. `null` clears `brand` or `minimum_quantity`. If `name` or `category` is included, it must contain a valid string. The following request clears the brand and sets an explicit zero threshold.

```json
{
  "brand": null,
  "minimum_quantity": "0"
}
```

### Dates and timestamps

`purchased_on` and `expires_on` are calendar dates in `YYYY-MM-DD` with years from 0001 to 9999. Keep them as dates when displaying them. Server timestamps such as `created_at`, `updated_at`, `added_at` and `occurred_at` use RFC 3339 UTC and end in `Z`. Fractional seconds may be present. The client can convert these timestamps to local time for display.

### Idempotent creation

Generate one `Idempotency-Key` for each new create action. A UUID string works. The key must be nonblank UTF-8 text of at most 200 bytes and must contain no NUL characters. After a timeout or a lost response, retry the same request with the same key.

The key is scoped to the user's personal space. A successful replay returns `201` with the result saved by the original request, including the original item ID. Reusing the key with different valid input returns `409 IDEMPOTENCY_KEY_CONFLICT`.

### Pagination

`limit` defaults to `20` and must be a positive integer. Values above `100` are capped at `100`. `offset` defaults to `0` and must be a nonnegative integer. An empty or invalid query value returns `400`. Items are sorted by `created_at` descending, then by `id` descending.

`meta.limit` and `meta.offset` show the values the server applied. Use `meta.next_offset` in the next request when `meta.has_more` is true. On the last page, `has_more` is false and `next_offset` is `null`. The list response does not include the total number of cards.

The following response represents an empty page.

```json
{
  "data": [],
  "meta": {
    "limit": 20,
    "offset": 0,
    "has_more": false,
    "next_offset": null
  }
}
```

### Gateway rate limits (VPS)

The VPS Caddy gateway applies two independent limits per client IP over a sliding five-second window.

| Operations | Requests per five seconds |
| --- | --- |
| `GET /items` and `GET /items/{item_id}` | 30 combined |
| `POST /items`, `PATCH /items/{item_id}` and `DELETE /items/{item_id}` | 10 combined |

Swagger and health checks are outside these limits. The local Nginx gateway and direct service calls are also outside these limits.

When the limit is reached, Caddy returns **HTTP 429** with `Retry-After` in seconds. Wait at least that long before retrying and use the same `Idempotency-Key` and body when retrying a create request. Handle this response by HTTP status because the gateway may return an empty or non-JSON body.

### Error format

Application errors include `code`, `status_code` and `message`. Some validation errors include `fields`, an array of field paths and codes. The array is omitted when empty. Use the codes to select localized messages and fall back to the HTTP status for unknown codes. Internal causes are kept out of the response.

The following error comes from a PATCH request with an invalid name and category.

```json
{
  "code": "INVALID_ITEM_UPDATE",
  "status_code": 400,
  "message": "Invalid item update data",
  "fields": [
    { "field": "name", "code": "INVALID_ITEM_NAME" },
    { "field": "category", "code": "INVALID_ITEM_CATEGORY" }
  ]
}
```

| HTTP | Code | Meaning |
| --- | --- | --- |
| 400 | `INVALID_REQUEST` | The body, path UUID or pagination parameters could not be read in the expected format |
| 400 | `INVALID_ITEM_CREATE` | Creation input failed validation. Check `fields` for details such as `INVALID_IDEMPOTENCY_KEY`, `INVALID_DATE` or `INVALID_BATCH_QUANTITY` |
| 400 | `INVALID_ITEM_UPDATE` | One or more editable fields failed validation. Check `fields` |
| 400 | `INVALID_QUANTITY_FORMAT`, `QUANTITY_PRECISION_EXCEEDED`, `QUANTITY_OUT_OF_RANGE` | A quantity could not be parsed. On create these codes can appear within `fields` |
| 401 | `UNAUTHORIZED` | The Bearer header is missing or malformed |
| 401 | `INVALID_TOKEN` | The access token is invalid or expired |
| 404 | `ITEM_NOT_FOUND` | The item is absent, deleted or outside the user's space |
| 404 | `SPACE_NOT_FOUND` | The user has no personal space for the requested get, update or delete operation |
| 409 | `IDEMPOTENCY_KEY_CONFLICT` | The key was already used for a different request |
| 415 | `UNSUPPORTED_MEDIA_TYPE` | The create or update request must use `Content-Type: application/json` |
| 500 | `INTERNAL_SERVER_ERROR` | An internal operation failed |

Errors generated by the router or gateway can use another format. Clients should also handle an empty or non-JSON error body.
