# Authentication Service

Kabletown authentication service - replicates Jellyfin's API key and user authentication system.

## Architecture

### Auth Flow

1. **Device Registration** (`POST /Devices`)
   - Client sends device name + app name
   - Server generates API token + device ID
   - Returns token for future requests

2. **User Authentication** (`POST /Users/AuthenticateByName`)
   - Client sends username + password hash
   - Server validates credentials against database
   - Returns user profile + API token (or validates existing token)

3. **Middleware Validation** (all protected endpoints)
   - Extract `X-Emby-Authorization` header
   - Parse Token and DeviceId
   - Lookup API key in database
   - Populate context with user_id, device_id, is_admin

### Header Format

```
X-Emby-Authorization: MediaBrowser Client="Jellyfin Web", Device="Firefox", DeviceId="abc123", Version="10.9.0", Token="xyz789"
```

### API Token Lifecycle

- Generated on device registration
- Stored in APIKeys table: `{Token, UserId, DeviceId, DateCreated, DateLastUsed}`
- Used for all authenticated requests
- No expiration (unless revoked)

## Required Database Tables

See `docs/auth_schema.md`

## Endpoints

### Auth Service (Port 8081)

#### POST /Devices
```
Request:
{
  "name": "My Device",
  "appName": "Jellyfin Web",
  "appVersion": "10.9.0"
}

Response: 200
{
  "Id": "device-uuid",
  "Name": "My Device",
  "AccessToken": "api-token-uuid",
  "AppId": "app-uuid"
}
```

#### POST /Devices/Authenticate
```
Request:
{
  "Username": "user",
  "Pw": "password"
}

Response: 200
{
  "Id": "user-uuid",
  "Name": "user",
  "AccessToken": "api-token-uuid"
}
```

### Item Service (Port 8082)

#### GET /Items
Uses same auth middleware. User ID from context passed to query:
```
GET /Items?TopParentId=xxx&UserId=uid
```

### Streaming Service (Port 8083)

#### GET /Videos/{itemId}/master.m3u8
Uses same auth middleware. Token used for transcode job registration.

## Integration with Other Services

**Auth Middleware** (`shared/auth/middleware.go`) provides:
- `AuthMiddleware(DeviceLookupFunc)` - middleware factory
- `AuthHelpers.GetUserID(r)` - get current user
- `AuthHelpers.GetDeviceID(r)` - get current device
- `AuthHelpers.GetIsAdmin(r)` - admin check

Services call lookup with their database:
```go
// In item-service main.go
lookup := func(token string) (string, bool, error) {
    var userID string
    var isAdmin bool
    err := db.QueryRow(
        "SELECT UserId, IsAdmin FROM ApiKeys WHERE Token = ?",
        token,
    ).Scan(&userID, &isAdmin)
    return userID, isAdmin, err
}

auth := shared.AuthMiddleware(lookup)
```

## Development

```bash
cd Kabletown/auth-service
go run cmd/server/main.go
```
