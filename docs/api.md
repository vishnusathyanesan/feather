# API Reference

Base URL: `/api/v1`

## Authentication

### Register
```
POST /auth/register
Body: { "email": "user@example.com", "name": "User", "password": "password123" }
Response: { "user": {...}, "access_token": "...", "refresh_token": "..." }
```

### Login
```
POST /auth/login
Body: { "email": "user@example.com", "password": "password123" }
Response: { "user": {...}, "access_token": "...", "refresh_token": "..." }
```

### Refresh Token
```
POST /auth/refresh
Body: { "refresh_token": "..." }
Response: { "user": {...}, "access_token": "...", "refresh_token": "..." }
```

### Logout
```
POST /auth/logout (requires auth)
Body: { "refresh_token": "..." }
Response: 204
```

### Get Current User
```
GET /auth/me (requires auth)
Response: { "id": "...", "email": "...", "name": "...", ... }
```

## Channels

### Create Channel
```
POST /channels (requires auth)
Body: { "name": "general", "type": "public", "topic": "..." }
Response: Channel object
```

### List Channels
```
GET /channels (requires auth)
Response: [Channel, ...]
```

### Get/Update/Delete Channel
```
GET    /channels/{channelID}
PATCH  /channels/{channelID}  Body: { "name": "...", "topic": "..." }
DELETE /channels/{channelID}  (admin only)
```

### Membership
```
POST /channels/{channelID}/join
POST /channels/{channelID}/leave
POST /channels/{channelID}/members  Body: { "user_id": "..." }
GET  /channels/{channelID}/members
POST /channels/{channelID}/read
```

## Messages

### Send Message
```
POST /channels/{channelID}/messages (requires auth)
Body: { "content": "Hello!", "parent_id": null }
Response: Message object
```

### List Messages
```
GET /channels/{channelID}/messages?before={id}&limit=50
Response: [Message, ...]
```

### Edit/Delete Message
```
PATCH  /channels/{channelID}/messages/{messageID}  Body: { "content": "..." }
DELETE /channels/{channelID}/messages/{messageID}
```

### Thread
```
GET /messages/{messageID}/thread
Response: [Message, ...] (parent + replies)
```

## Reactions
```
POST   /messages/{messageID}/reactions       Body: { "emoji": "👍" }
DELETE /messages/{messageID}/reactions/{emoji}
```

## Search
```
GET /search?q=keyword&channel_id=...&user_id=...&has=link&limit=20&offset=0
Response: { "messages": [...], "total_count": 42 }
```

## Webhooks
```
POST   /webhooks              Body: { "channel_id": "...", "name": "..." }
GET    /webhooks
DELETE /webhooks/{webhookID}
```

### Incoming Webhook
```
POST /hooks/{token}
Body: { "title": "Alert", "severity": "critical", "message": "...", "metadata": {...} }
Response: 204
```

## WebSocket
```
GET /ws?token={jwt}
```

### Event Types
- `message.new` / `message.updated` / `message.deleted`
- `reaction.added` / `reaction.removed`
- `typing`
- `presence.update`
- `channel.created` / `channel.updated` / `channel.deleted`
- `member.joined` / `member.left`

## File Uploads
```
POST /channels/{channelID}/files  (multipart/form-data, field: "file", max 20MB)
GET  /files/{fileID}/download     (redirects to presigned URL)
```
