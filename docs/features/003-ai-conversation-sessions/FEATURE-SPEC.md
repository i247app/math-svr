# Feature: AI conversation sessions (contextual multi-turn chat)

Feature ID: `003-ai-conversation-sessions`
Status: Draft
Last updated: 2026-06-28
Owners: business-analyst (spec), system-architect (design), security-engineer (review)

---

## 1. Summary

Introduce a first-class **conversation** aggregate so a client can hold a
contextual, multi-turn chat with the LLM. Each conversation has its own
external `conversation_id` (the "session" the boss is asking for). History
is persisted in MySQL and the last N turns are re-sent to the model on every
turn via the existing `internal/adapter/bot` adapter — so the model answers
*in context* while the server itself stays **stateless** (DB is the single
source of truth, safe across multiple EC2 instances).

This is the standard "store history + resend" pattern. The LLM vendor does
not remember anything by `conversation_id`; our server owns the history and
re-injects it. The `conversation_id` is what gives the *effect* of a
remembered session.

## 2. Business context

**Problem / opportunity.** Today every AI call (`quiz` generation/grading,
`exercise`) is stateless: each request is built fresh and the model has no
memory of prior turns (see `internal/libs/langchain/client.go:176` —
`Generate` sends one isolated request; no `memory`/`chains` are used). The
business wants to demonstrate a ChatGPT-style "the AI remembers the
conversation" capability for demos and future tutoring features.

**Stakeholders.**
- Parents / students — want a tutor that follows the thread of a discussion.
- Product / leadership — wants to show "session-aware AI" to sponsors.
- Backend platform — owns the new aggregate; must keep the server stateless
  and within token budget.
- Security — message content is user (often child) data → PII handling.

**Success criteria (business terms).**
- A user can ask a follow-up question and the AI answer reflects the prior
  turn, demonstrably, in a live demo.
- No regression to existing quiz/exercise latency or cost.

## 3. User stories & acceptance criteria

### US-1: Start a conversation and get a contextual reply
> As an authenticated user, I want to send a message without a
> `conversation_id` so the server starts a new session and replies.

- **Given** an authenticated (secure) session and a non-empty `message`,
  **when** the user POSTs `/ai/conversations/send` with no `conversation_id`,
  **then** the server returns `SUCCESS` with a newly minted
  `conversation_id`, the assistant `reply`, and persists exactly two
  `ma_ai_messages` rows (role `user`, then `assistant`) plus one
  `ma_ai_conversations` row owned by the session `user_id`.

### US-2: Continue an existing conversation in context
> As a user, I want to send a follow-up with the same `conversation_id` so
> the AI answers using earlier turns.

- **Given** an existing conversation owned by the caller with prior turns,
  **when** the user POSTs `/ai/conversations/send` with that
  `conversation_id` and a new `message`,
  **then** the server loads the last N messages (windowed, in `seq_no`
  order), builds `[]Message = [system] + windowed history + new user
  message`, calls the bot adapter, persists the new user + assistant
  messages with the next `seq_no` values, and returns the reply.
- **Given** a `conversation_id` that does not belong to the caller,
  **when** the user sends to it,
  **then** the server returns `AI_CONVERSATION_NOT_OWNED` and writes
  nothing.
- **Given** a `conversation_id` that does not exist,
  **when** the user sends to it,
  **then** the server returns `AI_CONVERSATION_NOT_FOUND`.

### US-3: List my conversations
> As a user, I want to list my conversations so I can resume one.

- **Given** the caller owns 3 active conversations,
  **when** they POST `/ai/conversations/list`,
  **then** the server returns those 3 (most-recent-first, paginated),
  excluding soft-deleted ones, without message bodies.

### US-4: Read one conversation with its history
> As a user, I want to open a conversation and see its messages.

- **Given** an owned conversation with M messages,
  **when** they GET `/ai/conversations/{id}`,
  **then** the server returns the conversation plus its messages in `seq_no`
  order (paginated/most-recent window), enforcing ownership.

### US-5: Soft-delete a conversation
> As a user, I want to delete a conversation.

- **Given** an owned active conversation,
  **when** they POST `/ai/conversations/soft-delete`,
  **then** its `conversation_status` becomes `DELETED`, it disappears from
  list/get, and the row is retained in DB.

### US-6: Graceful degradation when the bot is disabled
- **Given** `BOT_PROVIDER=""`/`"disabled"` (adapter pointer is `nil`),
  **when** a user sends a message,
  **then** the server returns `AI_CONVERSATION_DISABLED` and persists no
  assistant row (the user row may still be persisted — see §11).

## 4. Scope

**In scope (this feature):**
- New aggregate `conversation` with two entities: `Conversation` + `Message`,
  and two repo interfaces (`IRepository`, `IMessageRepository`).
- Two migrations: `021_ma_ai_conversations_table.sql`,
  `022_ma_ai_messages_table.sql`. Two new `ma_seqs` seed rows.
- Endpoints: `send`, `list`, `{id}`, `soft-delete` under `/ai/conversations`.
- Windowed history (last N messages) re-sent on each turn.
- New status codes + EN/VN messages.
- Wiring into `transaction.Repositories`, `SqlUnitOfWork`, container, routes.

**Out of scope / non-goals:**
- Rolling summary / token-summarisation of old turns (**phase 2**).
- Streaming responses (SSE) — `send` is a single synchronous reply.
- RAG / vector memory (`conversational_retrieval_qa`).
- Editing or deleting individual messages; regenerating a turn.
- Sharing conversations between users; classroom-scoped conversations.
- Per-IP/per-user rate-limiting beyond what auth already implies (note in
  §11 as a follow-up).
- Tool/function-calling, multi-modal (images/audio) messages.
- Reusing this aggregate to refactor existing quiz/exercise flows.

## 5. API design

All endpoints are **authenticated** (require a `secure` session via
`authMiddleware`), unlike the public `/ai/shake` warm-up.

### POST `/ai/conversations/send`  — command
Start (if `conversation_id` omitted) or continue a conversation; returns the
assistant reply.

**Request**
```go
type SendMessageReq struct {
    ConversationID *int64 `json:"conversation_id"` // nil -> create new
    Message        string `json:"message"`
    // UserID is injected from the session, never from the body.
}
```
**Response**
```go
type SendMessageRes struct {
    ConversationID int64           `json:"conversation_id"`
    Reply          string          `json:"reply"`
    Message        MessageResponse `json:"message"` // the persisted assistant turn
}
```

### POST `/ai/conversations/list`  — query
**Request**
```go
type ListConversationsReq struct {
    Page int `json:"page"`
    Size int `json:"size"`
    // UserID injected from session.
}
```
**Response**: `{ "conversations": []ConversationResponse, "pagination": ... }`
(no message bodies).

### GET `/ai/conversations/{id}`  — query
Path id parsed with `strconv.ParseInt(r.PathValue("id"), 10, 64)`.
**Response**: `ConversationResponse` + its `messages []MessageResponse`
(windowed, `seq_no` order). Ownership enforced.

### POST `/ai/conversations/soft-delete`  — command
**Request**: `{ "conversation_id": int64 }` → `WriteJsonNoContent` on success.

### DTO shapes (`internal/application/dto/conversation/`)
```go
type ConversationResponse struct {
    ConversationID int64   `json:"conversation_id"`
    Title          *string `json:"title,omitempty"`
    Purpose        *string `json:"purpose,omitempty"`
    MessageCount   int64   `json:"message_count"`
    CreateDt       string  `json:"create_dt"`
    ModifyDt       string  `json:"modify_dt"`
}
type MessageResponse struct {
    MessageID int64  `json:"message_id"`
    Role      string `json:"role"` // user | assistant | system
    Content   string `json:"content"`
    SeqNo     int64  `json:"seq_no"`
    CreateDt  string `json:"create_dt"`
}
```

### Status codes returned
| Code | When |
|---|---|
| `SUCCESS` (200) | reply produced / list / get |
| `NO_CONTENT` (204) | soft-delete |
| `AI_CONVERSATION_MISSING_MESSAGE` (13203) | empty `message` |
| `AI_CONVERSATION_MESSAGE_TOO_LONG` (13204) | over length cap |
| `AI_CONVERSATION_NOT_FOUND` (13201) | unknown `conversation_id` |
| `AI_CONVERSATION_NOT_OWNED` (13202) | conversation not owned by caller |
| `AI_CONVERSATION_GENERATION_FAILED` (13205) | adapter/LLM error |
| `AI_CONVERSATION_DISABLED` (13206) | bot adapter is `nil` |
| `UNAUTHORIZED` (existing) | no uid in session |
| `INTERNAL_SERVER_ERROR` | unexpected |

## 6. CQRS classification

- **`SendMessageCommand`** — write + LLM side effect.
  Handler `NewSendMessageCommandHandler(uow, botAdapter)`. The module
  service orchestrates two UoW writes around one out-of-tx adapter call
  (see §11 for why the LLM call is outside the tx).
- **`SoftDeleteConversationCommand`** — `NewSoftDeleteConversationCommandHandler(uow)`.
- **`GetConversationByIdQuery`** — read, repo direct
  (`NewGetConversationByIdQueryHandler(convRepo, msgRepo)`).
- **`ListConversationsQuery`** — read, repo direct
  (`NewListConversationsQueryHandler(convRepo)`).

## 7. Domain changes

**New aggregate** `internal/domain/conversation/`:

- `conversation.go` — `Conversation` entity (private fields + getters/setters
  per `rules/conventions.md`):
  - `id int64`, `conversationId int64`, `userId int64`,
    `profileId *int64`, `title *string`, `purpose *string`,
    `messageCount int64`, `conversationStatus *string`, `status string`,
    `createId *int64`, `createDt mtime.MathTime`, `modifyId *int64`,
    `modifyDt mtime.MathTime`.
- `message.go` — `Message` entity:
  - `id int64`, `messageId int64`, `conversationId int64`, `role string`,
    `content string`, `seqNo int64`, `status string`,
    `createDt mtime.MathTime`, `modifyDt mtime.MathTime`.
- `repository.go` — `IRepository`:
  - `FindByConversationId(ctx, conversationId int64) (*Conversation, error)`
  - `ListByUserId(ctx, *ListConversationsParams) ([]*Conversation, *pagination.Pagination, error)`
  - `Create(ctx, *Conversation) (*Conversation, error)`
  - `IncMessageCount(ctx, conversationId int64, delta int64) error`  // advances seq + count in the same tx
  - `SoftDeleteByConversationId(ctx, conversationId int64) error`
- `message.go` repo interface `IMessageRepository`:
  - `Create(ctx, *Message) (*Message, error)`
  - `ListByConversationId(ctx, conversationId int64, limit int64) ([]*Message, error)`  // last N, seq_no order
  - `ListWindow(ctx, conversationId int64, limit int64) ([]*Message, error)`  // alias used by the turn flow

- `enum` (new) `Role`: `system | user | assistant`. Validate at the module
  layer. Reuse `enum.StatusType` for `status`; `conversation_status` mirrors
  classroom's dual-status with `ACTIVE | DELETED`.
- **System prompt** is a fixed constant in `module/conversation`, rendered in
  the language from `metadata.GetClientLanguage(ctx)` (fallback VN). `title`
  and `purpose` stay `nil` in v1.

No changes to `quiz`, `exercise`, or `bot` domain.

## 8. Infrastructure changes

**Migrations** (forward-only; applied manually — boot migrate is disabled):

`021_ma_ai_conversations_table.sql` — `ma_ai_conversations`:
- standard skeleton (`id` PK, `conversation_id BIGINT UNSIGNED UNIQUE`),
  `user_id BIGINT UNSIGNED NOT NULL`, `profile_id BIGINT UNSIGNED NULL`,
  `title VARCHAR(255) NULL`, `purpose VARCHAR(32) NULL`,
  `message_count BIGINT UNSIGNED NOT NULL DEFAULT 0`,
  `conversation_status VARCHAR(32) DEFAULT 'ACTIVE'`,
  `status VARCHAR(32) DEFAULT 'ACTIVE'`, audit + time + `deleted_dt`.
- Index `idx_ma_ai_conversations_user_id` on `(user_id, modify_dt)` — list
  query (most-recent-first per user).

`022_ma_ai_messages_table.sql` — `ma_ai_messages`:
- skeleton (`id` PK, `message_id BIGINT UNSIGNED UNIQUE`),
  `conversation_id BIGINT UNSIGNED NOT NULL`,
  `role VARCHAR(16) NOT NULL`, `content LONGTEXT NOT NULL`,
  `seq_no BIGINT UNSIGNED NOT NULL`, `status VARCHAR(32) DEFAULT 'ACTIVE'`,
  audit + time.
- Index `idx_ma_ai_messages_conversation_seq` on `(conversation_id, seq_no)`
  — windowed history read + uniqueness of ordering.
- No FKs (project convention); integrity enforced inside the UoW.

**`ma_seqs` seed rows** (commented INSERTs in `000`, applied manually):
`ai_conversation`, `ai_message`.

**`internal/domain/seq/names.go`** — add `NameAiConversation = "ai_conversation"`
and `NameAiMessage = "ai_message"`.

**`transaction.Repositories`** — add `Conversation conversation.IRepository`
and `ConversationMessage conversation.IMessageRepository`.
**`SqlUnitOfWork.Do`** — construct both repos bound to the tx.

**New config sub-struct `ConversationConfig`** in
`internal/infrastructure/config/type.go`, assembled in `config.NewEnv`
(`config.go`) and reachable as `res.Env.ConversationConfig` (passed into the
conversation service in `container/services.go`):

| Field | Env var | Helper | Default |
|---|---|---|---|
| `HistoryWindowEnabled bool` | `CONVERSATION_HISTORY_WINDOW_ENABLED` | `getBoolConfigWithDefault` | `true` |
| `HistoryWindowSize int` | `CONVERSATION_HISTORY_WINDOW_SIZE` | `getIntConfigOptionalWithDefault` | `20` |
| `MaxMessageChars int` | `CONVERSATION_MAX_MESSAGE_CHARS` | `getIntConfigOptionalWithDefault` | `4000` |

- Add the three keys to `.env.example` with the defaults above.
- Defensive clamps in the service (a misconfigured env must not break a
  turn): `HistoryWindowSize` clamped to `[0, 200]`; `MaxMessageChars` clamped
  to `(0, 60000]` (stay under the adapter's `maxMessageBytes = 64*1024`); a
  non-positive value falls back to the default.

**No new adapter kind.** Reuses the existing `bot` adapter (`Chat`), keeping
its `MathError` mapping, retry, timeout, JSON-mode, and PII-safe logging.

**`container` + `routes`** — new `ConversationSvc`, new `module/conversation`
handler, four routes under `/ai/conversations/*` (all `authMiddleware`).

## 9. Status codes

New block `13200–13299` (`AI_CONVERSATION_*`). Highest code in use today is
`13104` (grep `code.go` before committing). Add in lockstep to `code.go`,
`message_en.go`, `message_vn.go`:

| Constant | Numeric | EN | VN |
|---|---|---|---|
| `AI_CONVERSATION_NOT_FOUND` | 13201 | Conversation not found. | Không tìm thấy cuộc trò chuyện. |
| `AI_CONVERSATION_NOT_OWNED` | 13202 | You do not own this conversation. | Bạn không sở hữu cuộc trò chuyện này. |
| `AI_CONVERSATION_MISSING_MESSAGE` | 13203 | Message is required. | Nội dung tin nhắn là bắt buộc. |
| `AI_CONVERSATION_MESSAGE_TOO_LONG` | 13204 | Message is too long. | Tin nhắn quá dài. |
| `AI_CONVERSATION_GENERATION_FAILED` | 13205 | Failed to generate a reply. | Không thể tạo câu trả lời. |
| `AI_CONVERSATION_DISABLED` | 13206 | AI chat is currently unavailable. | Tính năng trò chuyện AI hiện không khả dụng. |

## 10. Validation

- **Module validator** (`validator.go`):
  - `Message` trimmed non-empty → `AI_CONVERSATION_MISSING_MESSAGE`.
  - `len(Message)` ≤ `MaxMessageChars` (from
    `CONVERSATION_MAX_MESSAGE_CHARS`, default 4000; clamped under the
    adapter's `maxMessageBytes = 64*1024`) →
    `AI_CONVERSATION_MESSAGE_TOO_LONG`.
  - `ConversationID` (when present) must be > 0.
- **Ownership** check in service before any LLM call: load conversation,
  compare `conversation.UserId()` to session uid → `AI_CONVERSATION_NOT_OWNED`.
- **Domain invariant:** `seq_no` is strictly increasing per conversation;
  it is derived from `message_count` inside the UoW (never client-supplied).
- **Role** is server-assigned (`user` for the request, `assistant` for the
  reply); never taken from the client.

## 11. Non-functional requirements

- **Statelessness.** No per-conversation in-process state. DB is the source
  of truth → any EC2 instance can serve any turn. (Contrast: langchaingo's
  in-memory `memory.*` would not survive restart or scale across instances —
  the reason we self-manage history; see `docs` decision note.)
- **Transaction boundary.** The LLM call is **outside** `uow.Do` (slow I/O
  must not hold a tx open — same rule the `quiz` module follows,
  `module/quiz/service.go:85`). The turn runs as:
  1. **UoW #1:** find-or-create conversation; mint + persist the **user**
     message (`seq_no = message_count`); `IncMessageCount(+1)`; return the
     windowed history.
  2. **Outside tx:** `botAdapter.Chat([]Message{system, ...history, user})`.
  3. **UoW #2:** mint + persist the **assistant** message
     (`seq_no = message_count`); `IncMessageCount(+1)`.
  On LLM failure after step 1, the user message stays persisted (client can
  retry the turn); return `AI_CONVERSATION_GENERATION_FAILED`. (Alternative —
  persist both messages in a single post-call UoW — is rejected: it loses the
  user turn on failure and complicates retries.)
- **Token budget / cost — env-driven.** The history window is configurable
  at deploy time (no rebuild):
  - `CONVERSATION_HISTORY_WINDOW_ENABLED` (default `true`) — master toggle.
    When **`false`**, a turn sends only `[system] + new user message` (no
    prior context re-sent); messages are still persisted, so turning it back
    on resumes context. This is the kill-switch for cost/latency.
  - `CONVERSATION_HISTORY_WINDOW_SIZE` (default `20`) — `N`, how many recent
    messages are re-sent when the window is enabled.
  - Cost grows with history; the window (and the toggle) caps it. Rolling
    summary (phase 2) will compress older turns.
- **Latency.** Dominated by the LLM (seconds). Non-LLM overhead (2 small
  UoW writes + 1 windowed read) target p99 < 100 ms.
- **Idempotency.** `send` is **not** idempotent (each call appends a turn).
  Clients must not auto-retry on timeout without user intent.
- **Consistency.** Strong / read-your-write (single MySQL).

## 12. Security considerations

- **Auth.** All four endpoints require a `secure` session (`authMiddleware`).
  Ownership enforced on every per-conversation operation.
- **PII.** `content` is user-authored, often a child's input. **Never log
  message content** (extends rule 6 / `rules/security.md`). Log only
  `conversation_id`, `user_id`, `role`, `seq_no`, lengths, latency. Add
  `message`, `content`, `reply` to the `LogRequestMiddleware` redaction list.
- **Rate-limit.** Not in scope; flag as a follow-up — an authenticated user
  could burn vendor quota by spamming `send`. Consider an OTP-style policy
  later (`application/command/.../policy.go` precedent).
- **Secret material.** None handled here.
- **Tenant isolation.** A user must never read another user's conversation
  or messages — covered by the ownership check; add a repo-level
  `WHERE user_id = ?` on list and a join/guard on get.
- **Trigger `security-engineer` review: YES** — persistent storage of
  (child) PII + a public-facing AI cost vector.

## 13. Admin / operator UX

Not admin-facing. Observability expectations (content-free):
- `INFO conversation.turn conversation_id=<id> uid=<uid> in_len=<n> reply_len=<n> latency_ms=<n>`
- `WARN conversation.generation_failed conversation_id=<id> uid=<uid> err=<wrapped>`
- `INFO conversation.created conversation_id=<id> uid=<uid>`

## 14. Risks & mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Re-sending history inflates token cost | High | Medium | Window `N=20` now; rolling summary phase 2; per-user rate-limit follow-up. |
| Message content leaks into logs | Medium | High | Redaction list + code review; never pass content to `logger`. |
| Concurrent turns on one conversation race `seq_no` | Low | Medium | `seq_no` derived from `message_count` advanced inside the UoW (row-locked update), not `MAX()+1` outside a tx. |
| LLM failure loses the user turn | Medium | Low | User message persisted in UoW #1 before the call; failure returns a typed error and keeps the turn. |
| Server scaled to N instances breaks "memory" | Low | High | DB-backed history (no in-process state); explicitly chosen over langchaingo in-memory `memory.*`. |
| Unbounded conversation growth | Medium | Low | Pagination on get; future retention/cleanup job. |

## 15. Resolved decisions

All previously-open questions were decided with the product owner on
2026-06-28:

1. **System prompt source — fixed constant.** A tutor system prompt lives as
   a constant in `module/conversation` (not reused from `domain/bot`, whose
   templates are quiz-shaped). Revisit consolidating into `domain/bot` only
   if a second chat surface appears.
2. **`title` / `purpose` — nullable in v1.** Not collected from the client
   and not AI-generated for now. The client may display a placeholder
   ("Cuộc trò chuyện #id" or the first message). AI auto-titling is deferred
   to phase 2 (`004-...`).
3. **Window / length — env-driven (not hardcoded).** Three new env vars,
   read into `ConversationConfig` (see §8):
   `CONVERSATION_HISTORY_WINDOW_ENABLED` (default `true`) to turn the history
   window on/off, `CONVERSATION_HISTORY_WINDOW_SIZE` (default `20`) for `N`,
   and `CONVERSATION_MAX_MESSAGE_CHARS` (default `4000`). All tunable per
   deploy with no rebuild and no schema change.
4. **`send` rate-limiting — deferred for v1.** Acceptable for the controlled
   sponsor demo (authenticated, small audience). **Mandatory before any wide
   public rollout** — tracked as follow-up `006-ai-conversation-rate-limit`.
5. **Prompt language — follow the client.** Resolve via
   `metadata.GetClientLanguage(ctx)` (same as the quiz module), falling back
   to Vietnamese (`enum.LanguageTypeVietnamese`, rule 10) when absent.

---

## Follow-ups triggered by this spec

- `security-engineer` review before merge (PII storage + cost vector).
- `plan-implementation` to convert this spec into a phased, file-by-file
  plan (migration → seq → domain → repo → DTO → status → command/query →
  transaction wiring → module → container → routes).
- Phase 2 features:
  - `004-ai-conversation-rolling-summary` — compress old turns.
  - `005-ai-conversation-streaming` — SSE token streaming.
  - `006-ai-conversation-rate-limit` — per-user send policy.
