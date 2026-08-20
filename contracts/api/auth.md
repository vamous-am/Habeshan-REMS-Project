# Auth Contract

Published by Dev 1 after `feature/dev1-auth-middleware` merges.
This is the source of truth for Dev 2, Dev 3, and Dev 4 — read this
before protecting any route or consuming any JWT claim.

Change any of this only with team agreement — other slices build directly
against it.

---

## Endpoints

### POST /api/v1/auth/register
**Public** — no token required.

Request body:
```json
{
  "org_name":  "string (required, min 2 chars)",
  "full_name": "string (required, min 2 chars)",
  "email":     "string (required, valid email)",
  "password":  "string (required, min 8 chars)",
  "phone":     "string (optional)"
}
```

Response `201`:
```json
{
  "status": "success",
  "data": {
    "token": "<jwt>",
    "user": {
      "id":        "uuid",
      "org_id":    "uuid",
      "email":     "string",
      "full_name": "string",
      "phone":     "string (omitted if empty)",
      "role":      "employee",
      "status":    "active"
    }
  }
}
```

> ⚠️ `role` is **always** `employee` regardless of what the client sends — FR-AUTH-08.
> The request has no `role` field, so there is nothing a client can send to change it.

Errors:
| Status | When |
|--------|------|
| `400`  | Missing/invalid fields (bad email, password < 8 chars, etc.) |
| `409`  | Email already used in the organization created by this same call — only reachable if two requests race on the same email, since each call creates a brand-new org |

**🚧 Open issue — needs a decision before Branch 5/6 (admin endpoints) are usable:**
Since `Register` is the only way an organization gets created, and every user
it creates is `employee`, there is currently **no way for any organization to
get its first `admin`.** Endpoints gated behind `RequireRole("admin")` are
permanently unreachable until this is resolved (seed script, invite flow,
"first user in a new org is admin" exception, or similar). Flag this in
sprint planning — do not silently work around it in a feature branch.

---

### POST /api/v1/auth/login
**Public** — no token required.

Request body:
```json
{
  "email":    "string (required)",
  "password": "string (required)"
}
```

Response `200`: same `user`/`token` shape as the register response above.

> Wrong email and wrong password both return `401` with the same message,
> to prevent user enumeration.

Errors:
| Status | When |
|--------|------|
| `401`  | Email not found, or password incorrect (same message for both) |

**⚠️ Known limitation:** `Login` looks up by email only — **not** scoped to
an organization (`WHERE email = ? AND deleted_at IS NULL`). The schema
enforces `UNIQUE(org_id, email) WHERE deleted_at IS NULL`, which explicitly
allows the same email to exist under two different organizations. If that
ever happens, login returns an arbitrary one of the matching rows — not
necessarily the one the caller intended. Rare today since `Register` always
creates one org+user together in a single call, but it's a landmine for
later (e.g. an invite-based multi-org flow). Track separately from this
branch.

---

## JWT

All protected endpoints require:
```
Authorization: Bearer <token>
```

### Token payload (`auth.Claims`)
```json
{
  "user_id": "uuid string",
  "org_id":  "uuid string",
  "role":    "admin | manager | employee",
  "sub":     "uuid string (same as user_id)",
  "iat":     1234567890,
  "exp":     1234567890,
  "jti":     "uuid string, unique per token"
}
```

Signed **HS256** with `JWT_SECRET`. Fixed **24h** expiry.

---

## How to protect your routes

```go
import "github.com/habeshan-rems/backend/internal/middleware"

// Any authenticated user (all roles):
app.Get("/attendance/me", middleware.JWTAuth(), yourHandler)

// Manager or Admin only:
app.Get("/attendance/team", middleware.JWTAuth(), middleware.RequireRole("admin", "manager"), yourHandler)

// Admin only:
app.Post("/admin/users", middleware.JWTAuth(), middleware.RequireRole("admin"), yourHandler)
```

`RequireRole` must always run **after** `JWTAuth()` — it reads the role
`JWTAuth()` stores in `fiber.Locals`. Using it alone (without `JWTAuth()`
first) returns `401`, not `403`, since there's no role to check yet.

## How to read claims in your handler

Use the exported constants — **do not** use raw string literals like
`"userID"` or `"orgId"` for `c.Locals` keys. A typo in a hand-typed string
key fails silently (`c.Locals` returns `nil`, which casts to an empty
string), and an empty `org_id` used to scope a DB query is a cross-org data
leak that only shows up at runtime.

```go
import "github.com/habeshan-rems/backend/internal/middleware"

func MyHandler(c *fiber.Ctx) error {
    userID := c.Locals(middleware.LocalUserID).(string)
    orgID  := c.Locals(middleware.LocalOrgID).(string)
    role   := c.Locals(middleware.LocalRole).(string)
    // use orgID to scope every query — required by NFR-SEC-03
}
```

| Local key                 | Type   |
|----------------------------|--------|
| `middleware.LocalUserID`   | string |
| `middleware.LocalOrgID`    | string |
| `middleware.LocalRole`     | string |

---

## Error responses

All errors use the standard envelope:
```json
{ "status": "fail", "message": "..." }
```

| Status | Message                          | When |
|--------|-----------------------------------|------|
| `400`  | varies (field-specific)           | Missing/invalid request fields |
| `401`  | `"missing or malformed token"`    | No `Authorization` header, or not `Bearer <token>` format |
| `401`  | `"invalid or expired token"`      | Bad signature, malformed token, or expired |
| `401`  | `"unauthorized"`                  | `RequireRole` used without `JWTAuth()` running first (misconfigured route) |
| `403`  | `"forbidden"`                     | Valid token, but role not in the route's allowed list |
| `409`  | `"resource already exists"`       | Email already registered (register endpoint) |
| `500`  | `"internal server error"`         | Unexpected server error |

Note the split: **request-level checks** (missing header, malformed token,
bad body, missing fields) use hand-written `common.Fail` messages — this
matches the convention already established in `auth/handlers.go`. **Role
authorization** uses the sentinel `common.ErrForbidden` / `common.ErrUnauthorized`
via `common.HandleError`, since there's no per-case nuance worth preserving
there. Keep this split when adding new checks — don't collapse everything
into one generic message, and don't hand-write messages for cases sentinels
already cover.