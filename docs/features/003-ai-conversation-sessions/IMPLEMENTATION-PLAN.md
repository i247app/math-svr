# Implementation Plan: AI conversation sessions

Feature: `003-ai-conversation-sessions`
Spec: `./FEATURE-SPEC.md`
Status: Ready for implementation
Last updated: 2026-06-28
Owners: system-architect (build order), database-architect (schema),
golang-specialist (code templates), security-engineer (PII review).

---

## Overview

New aggregate `conversation` with a sub-entity `Message`. A `send` endpoint
appends a user turn, re-sends the windowed history to the existing `bot`
adapter, and appends the assistant turn — all DB-backed so the server stays
stateless. History window + message length are env-driven. No new adapter
kind.

**Scope estimate.**
- Files to create: ~16
- Files to modify: ~10
- Tests: 3 (two command handlers, one module validator)
- Effort: ~2 developer-days

**Build order (enforced, do not deviate):**
Migration + seq seed → seq names → domain → status codes → MySQL repos →
transaction wiring → DTO → application (command/query) → config → module →
bootstrap (PAUSE) → tests → verification.

**Conventions reminders (load-bearing):**
- Migrations are **forward-only** — no down-migrations; filenames are
  sequential `NNN_*.sql` (next free: **021**, **022**). `database.Migrate` is
  disabled at boot → apply manually on each env.
- All writes go through `transaction.UnitOfWork`. IDs are `int64` minted via
  `repos.Seq.Next(ctx, seq.NameX)` inside the UoW. No UUIDs.
- Repo reads translate `sql.ErrNoRows` → `(nil, nil)`; wrap every error
  `fmt.Errorf("conversation repo <op>: %w", err)`; reads filter dual-status.
- **Never log message content** (PII).

---

## Phase 1 — Migration + seq seed

- [ ] **1.1** Create `migrations/021_ma_ai_conversations_table.sql`.
  - `ma_ai_conversations` per spec §8: `id` PK, `conversation_id BIGINT
    UNSIGNED UNIQUE`, `user_id BIGINT UNSIGNED NOT NULL`, `profile_id BIGINT
    UNSIGNED NULL`, `title VARCHAR(255) NULL`, `purpose VARCHAR(32) NULL`,
    `message_count BIGINT UNSIGNED NOT NULL DEFAULT 0`,
    `conversation_status VARCHAR(32) DEFAULT 'ACTIVE'`,
    `status VARCHAR(32) DEFAULT 'ACTIVE'`, audit + `*_dt` + `deleted_dt`.
  - Index `idx_ma_ai_conversations_user_id` on `(user_id, modify_dt)`.
  - `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`.
- [ ] **1.2** Create `migrations/022_ma_ai_messages_table.sql`.
  - `ma_ai_messages` per spec §8: `id` PK, `message_id BIGINT UNSIGNED
    UNIQUE`, `conversation_id BIGINT UNSIGNED NOT NULL`,
    `role VARCHAR(16) NOT NULL`, `content LONGTEXT NOT NULL`,
    `seq_no BIGINT UNSIGNED NOT NULL`, `status VARCHAR(32) DEFAULT 'ACTIVE'`,
    audit + `*_dt`.
  - Index `idx_ma_ai_messages_conversation_seq` on `(conversation_id, seq_no)`.
- [ ] **1.3** Edit `migrations/000_ma_seqs_table.sql`.
  - Add two seed `INSERT` rows (kept commented per repo convention, applied
    manually): `ai_conversation`, `ai_message`.

> ⚠️ `Edit(migrations/**)` is `ask`-gated. Apply 021/022 + the two seq rows
> manually on every environment before the routes go live.

---

## Phase 2 — Seq names

- [ ] **2.1** Edit `internal/domain/seq/names.go`.
  - Add `NameAiConversation = "ai_conversation"` and
    `NameAiMessage = "ai_message"` (strings must match the 1.3 seed rows).

---

## Phase 3 — Domain

- [ ] **3.1** Create `internal/domain/conversation/conversation.go`.
  - Entity `Conversation` (private fields per spec §7), `NewConversation()`,
    getters/setters. `createDt/modifyDt` are `mtime.MathTime`; nullable
    fields are pointers (`profileId *int64`, `title/purpose/conversationStatus
    *string`).
- [ ] **3.2** Create `internal/domain/conversation/message.go`.
  - Entity `Message` (`id, messageId, conversationId, role, content, seqNo,
    status, createDt, modifyDt`), `NewMessage()`, getters/setters.
- [ ] **3.3** Create `internal/domain/conversation/repository.go`.
  - `type ListConversationsParams struct { UserID int64; Page, Limit int64 }`.
  - `IRepository`:
    - `FindByConversationId(ctx, conversationId int64) (*Conversation, error)`
    - `ListByUserId(ctx, *ListConversationsParams) ([]*Conversation, *pagination.Pagination, error)`
    - `Create(ctx, *Conversation) (*Conversation, error)`
    - `IncMessageCount(ctx, conversationId int64, delta int64) error`
    - `SoftDeleteByConversationId(ctx, conversationId int64) error`
- [ ] **3.4** Create `internal/domain/conversation/message_repository.go`.
  - `IMessageRepository`:
    - `Create(ctx, *Message) (*Message, error)`
    - `ListRecentByConversationId(ctx, conversationId int64, limit int64) ([]*Message, error)` — last `limit` rows, returned ascending by `seq_no`.
- [ ] **3.5** Edit `internal/shared/enum/` (new file `conversation.go` or
  existing enum file).
  - `Role` type with `RoleSystem/RoleUser/RoleAssistant` + `IsValid()`.
  - `ConversationStatusType` with `ACTIVE/DELETED` (mirror classroom status).

---

## Phase 4 — Status codes (lockstep, 3 files)

- [ ] **4.1** Edit `internal/domain/shared/status/code.go`.
  - New block `13200–13299`: `AI_CONVERSATION_NOT_FOUND = 13201`,
    `AI_CONVERSATION_NOT_OWNED = 13202`,
    `AI_CONVERSATION_MISSING_MESSAGE = 13203`,
    `AI_CONVERSATION_MESSAGE_TOO_LONG = 13204`,
    `AI_CONVERSATION_GENERATION_FAILED = 13205`,
    `AI_CONVERSATION_DISABLED = 13206`. (Grep first; highest in use is 13104.)
- [ ] **4.2** Edit `internal/domain/shared/status/message_en.go` — EN strings (spec §9).
- [ ] **4.3** Edit `internal/domain/shared/status/message_vn.go` — VN strings (spec §9).

---

## Phase 5 — MySQL (models + repos)

- [ ] **5.1** Create `internal/infrastructure/persistence/mysql/models/conversation.go`.
  - `ConversationModel` + `MessageModel` (raw columns, `time.Time`,
    `sql.Null*`/pointers for nullable). `ModelToDomain` mappers for both.
- [ ] **5.2** Create `internal/infrastructure/persistence/mysql/repositories/conversation_repository.go`.
  - Pattern from `school_repository.go`: `conversationColumns` constant
    (lockstep with `scanConversation`), `conversationActiveWhere` +
    `conversationActiveArgs()` (filter `status` + `conversation_status !=
    DELETED` + `deleted_dt IS NULL`), `findOneBy` helper.
  - `NewConversationRepository(db database.Executor) conversation.IRepository`.
  - `Create` → INSERT then `findBareById` hydrate.
  - `ListByUserId` → `WHERE user_id = ? AND <active>` ORDER BY `modify_dt
    DESC` with pagination (`pagination.NewPagination`).
  - `IncMessageCount` → `UPDATE ... SET message_count = message_count + ?`
    (row-locked; this is what makes `seq_no` race-safe).
  - `SoftDeleteByConversationId` → set `conversation_status = DELETED`,
    `deleted_dt = NOW(6)`.
- [ ] **5.3** Create `internal/infrastructure/persistence/mysql/repositories/conversation_message_repository.go`.
  - `messageColumns` + `scanMessage`, `NewConversationMessageRepository(db) conversation.IMessageRepository`.
  - `Create` → INSERT then hydrate.
  - `ListRecentByConversationId` → `WHERE conversation_id = ? AND status = ?
    ORDER BY seq_no DESC LIMIT ?`, then reverse to ascending before return.

---

## Phase 6 — Transaction wiring

- [ ] **6.1** Edit `internal/application/transaction/unit_of_work.go`.
  - Add `Conversation conversation.IRepository` and
    `ConversationMessage conversation.IMessageRepository` to `Repositories`.
- [ ] **6.2** Edit `internal/infrastructure/persistence/mysql/unit_of_work.go`.
  - In `SqlUnitOfWork.Do`, construct
    `NewConversationRepository(loggedTx)` and
    `NewConversationMessageRepository(loggedTx)` into the bundle.

---

## Phase 7 — DTO

- [ ] **7.1** Create `internal/application/dto/conversation/conversation_dto.go`.
  - `SendMessageReq { ConversationID *int64; Message string; UserID *int64 (session-injected) }`.
  - `SendMessageRes { ConversationID int64; Reply string; Message MessageResponse }`.
  - `ListConversationsReq { Page, Size int; UserID *int64 }`,
    `ListConversationsRes { Conversations []ConversationResponse; Pagination ... }`.
  - `GetConversationByIdRes { Conversation ConversationResponse; Messages []MessageResponse }`.
  - `ConversationResponse`, `MessageResponse` (spec §5) + `DomainToResponse` /
    `DomainListToResponse` / `MessageDomainToResponse` mappers. Snake-case tags.

---

## Phase 8 — Application (command + query)

- [ ] **8.1** Create `internal/application/command/conversation/append_user_message_command.go`.
  - `AppendUserMessageCommand { ConversationID *int64; UserID int64; ProfileID *int64; Content string }`.
  - `Handle` inside `uow.Do`:
    1. If `ConversationID == nil` → mint via `repos.Seq.Next(ctx,
       seq.NameAiConversation)`, build + `repos.Conversation.Create`
       (`message_count = 0`, owner `UserID`).
    2. Else `repos.Conversation.FindByConversationId`; nil → `AI_CONVERSATION_NOT_FOUND`;
       `UserId() != UserID` → `AI_CONVERSATION_NOT_OWNED`.
    3. Mint message id `seq.NameAiMessage`; `seq_no = conversation.MessageCount()`;
       persist user `Message` (role `user`).
    4. `repos.Conversation.IncMessageCount(+1)`.
    5. Return the (possibly new) `*Conversation` + the persisted user `*Message`.
- [ ] **8.2** Create `internal/application/command/conversation/append_assistant_message_command.go`.
  - `AppendAssistantMessageCommand { ConversationID int64; Content string }`.
  - `Handle` inside `uow.Do`: re-read conversation for current
    `message_count`, mint message id, `seq_no = message_count`, persist
    assistant `Message`, `IncMessageCount(+1)`; return the `*Message`.
- [ ] **8.3** Create `internal/application/command/conversation/soft_delete_conversation_command.go`.
  - Ownership re-checked, then `repos.Conversation.SoftDeleteByConversationId`.
- [ ] **8.4** Create `internal/application/query/conversation/get_conversation_by_id_query.go`.
  - Loads conversation (repo direct) + `ListRecentByConversationId` for its
    messages. Returns `(nil, nil)` when not found (ownership enforced in module).
- [ ] **8.5** Create `internal/application/query/conversation/list_conversations_query.go`.
  - `ListByUserId` with pagination.

---

## Phase 9 — Config (env)

- [ ] **9.1** Edit `internal/infrastructure/config/type.go`.
  - Add `ConversationConfig` struct: `HistoryWindowEnabled bool`,
    `HistoryWindowSize int`, `MaxMessageChars int`. Add
    `ConversationConfig ConversationConfig` field to `Env`.
- [ ] **9.2** Edit `internal/infrastructure/config/config.go`.
  - In `NewEnv`, assemble:
    `HistoryWindowEnabled: getBoolConfigWithDefault("CONVERSATION_HISTORY_WINDOW_ENABLED", true)`,
    `HistoryWindowSize: getIntConfigOptionalWithDefault("CONVERSATION_HISTORY_WINDOW_SIZE", 20)`,
    `MaxMessageChars: getIntConfigOptionalWithDefault("CONVERSATION_MAX_MESSAGE_CHARS", 4000)`.
- [ ] **9.3** Edit `.env.example`.
  - Add the three keys with default values + one-line comments.

---

## Phase 10 — Module (presentation + orchestration)

- [ ] **10.1** Create `internal/module/conversation/errors.go`.
  - Sentinel errors (`ErrConversationNotFound`, `ErrConversationNotOwned`,
    `ErrMessageRequired`, `ErrMessageTooLong`, `ErrBotDisabled`,
    `ErrUidNotFoundFromSession`).
- [ ] **10.2** Create `internal/module/conversation/validator.go`.
  - `ValidateSend(ctx, req, maxMessageChars)`: trim non-empty →
    `AI_CONVERSATION_MISSING_MESSAGE`; length ≤ `maxMessageChars` →
    `AI_CONVERSATION_MESSAGE_TOO_LONG`; `ConversationID` (if set) > 0.
- [ ] **10.3** Create `internal/module/conversation/prompt.go`.
  - `systemPrompt(lang enum.LanguageType) string` — fixed tutor prompt,
    VN + EN variants; chosen by `metadata.GetClientLanguage(ctx)` (fallback VN).
- [ ] **10.4** Create `internal/module/conversation/service.go`.
  - `Service` holds the two command handlers, two query handlers, the two
    repos (for direct reads), `bot *botAdapter.Adapter`, and a sanitized
    `cfg` (enabled / windowSize-clamped-[0,200] / maxChars-clamped-(0,60000]).
  - `NewService(uow, convRepo, msgRepo, bot, cfg)`.
  - `SendMessage(ctx, req)`:
    1. `nil`-guard `bot` → `AI_CONVERSATION_DISABLED`.
    2. `ValidateSend(..., cfg.MaxMessageChars)`.
    3. `AppendUserMessageCommand.Handle` (UoW#1) → conversation + user msg.
    4. Build `[]botAdapter.Message`: `{system}` +, **only if
       `cfg.HistoryWindowEnabled`**, `ListRecentByConversationId(convId,
       cfg.HistoryWindowSize)` mapped to messages (already ends with the new
       user turn); when disabled, just `{system, user}`.
    5. `bot.Chat(...)` **outside any UoW** → on error
       `AI_CONVERSATION_GENERATION_FAILED` (user turn already persisted).
    6. `AppendAssistantMessageCommand.Handle` (UoW#2).
    7. Map → `SendMessageRes`.
  - `ListConversations`, `GetConversationById` (enforce ownership after load
    → `AI_CONVERSATION_NOT_OWNED`), `SoftDeleteConversation` (ownership).
- [ ] **10.5** Create `internal/module/conversation/handler.go`.
  - `Handler` + `NewHandler(res *resource.Resource, svc *Service)`.
  - `HandleSend`, `HandleList`, `HandleGet` (path id via
    `strconv.ParseInt(r.PathValue("id"),10,64)`), `HandleSoftDelete`.
  - Each: decode → pull uid from `res.GetRequestSession(r)` (inject into req)
    → call service → `response.WriteJson`. Never log `Message`/`reply`.

---

## Phase 11 — Bootstrap (wiring)

**PAUSE here. A human reviews these edits before tests** (highest blast radius).

- [ ] **11.1** Edit `internal/bootstrap/container/type.go`.
  - Import `module/conversation`; add `ConversationSvc *conversation.Service`
    to `ServiceContainer`; add `ConversationRepository` +
    `ConversationMessageRepository` to `RepositoryContainer`.
- [ ] **11.2** Edit `internal/bootstrap/container/repositories.go`.
  - Construct the two repos from `res.DB` (for the module's direct reads).
- [ ] **11.3** Edit `internal/bootstrap/container/services.go`.
  - `conversationService := conversation.NewService(uow,
    repos.ConversationRepository, repos.ConversationMessageRepository,
    res.BotProvider, res.Env.ConversationConfig)`; add to the returned
    `ServiceContainer`.
- [ ] **11.4** Edit `internal/bootstrap/routes/routes.go`.
  - Import `module/conversation`; new block (all `authMiddleware`):
    - `POST /ai/conversations/send` → `HandleSend`
    - `POST /ai/conversations/list` → `HandleList`
    - `GET  /ai/conversations/{id}` → `HandleGet`
    - `POST /ai/conversations/soft-delete` → `HandleSoftDelete`
- [ ] **11.5** Edit `internal/bootstrap/middleware/log_request_middleware.go`.
  - Add `message`, `content`, `reply` to the JSON-body redaction list (PII).

---

## Phase 12 — Tests

- [ ] **12.1** Create `internal/application/command/conversation/append_user_message_command_test.go`.
  - Table-driven, hand-rolled UoW/repo/seq fakes. Cases: new conversation
    (mints id), existing owned, not found, not owned, seq_no = prior count.
- [ ] **12.2** Create `internal/application/command/conversation/append_assistant_message_command_test.go`.
  - Cases: appends with next seq_no, inc count called once.
- [ ] **12.3** Create `internal/module/conversation/validator_test.go`.
  - Cases: empty message, over-length (boundary at `maxMessageChars`), valid,
    bad `conversation_id`.
- [ ] **12.4** (optional, `//go:build integration`) repo tests against a real
  MySQL DSN for ordering + `IncMessageCount` concurrency.

---

## Phase 13 — Verification

- [ ] **13.1** `gofmt -w` changed files (PostToolUse hook also does this).
- [ ] **13.2** `go vet ./...`.
- [ ] **13.3** `go build ./...`.
- [ ] **13.4** `go test ./...` (short).
- [ ] **13.5** Manual smoke (`make run`): `send` with no id → get
    `conversation_id` + reply; `send` again with that id → reply reflects
    prior turn; flip `CONVERSATION_HISTORY_WINDOW_ENABLED=false` → reply no
    longer uses context; `list`/`{id}`/`soft-delete` behave per spec.

---

## Dependencies

| Phase | Depends on | Notes |
|---|---|---|
| 2 | 1 | seq names must match seed rows |
| 3 | — | pure domain |
| 4 | — | status codes |
| 5 | 3 | repos implement domain interfaces |
| 6 | 5 | UoW bundles the repos |
| 7 | 3 | DTO maps domain |
| 8 | 4, 6, 7 | commands/queries use repos + status + dto |
| 9 | — | config independent |
| 10 | 8, 9 | service orchestrates handlers + config + adapter |
| 11 | 10 | bootstrap wires module (PAUSE) |
| 12 | 8, 10 | tests target handlers + validator |
| 13 | 12 | final verification |

## Risk log during implementation

- **`seq_no` race.** Two concurrent `send` on one conversation must not
  collide — `seq_no` is derived from `message_count` advanced by
  `IncMessageCount` inside the UoW (row-locked `UPDATE`), never `MAX()+1`.
- **PII in logs.** Confirm 11.5 redaction + no `logger` call ever receives
  `Message`/`content`/`reply`. Flag to `security-engineer`.
- **Env misconfig.** Service clamps window/length; non-positive → default.
- **Bot disabled mid-flight.** `nil`-guard returns `AI_CONVERSATION_DISABLED`
  before any DB write.

## Follow-ups (out of this plan)

- `004-ai-conversation-rolling-summary` (phase 2 — compress old turns).
- `005-ai-conversation-streaming` (SSE).
- `006-ai-conversation-rate-limit` (per-user `send` policy — required before
  wide public rollout).

## Sign-off checkboxes

- [ ] `architecture-guardian` — no layer violations.
- [ ] `database-architect` — migrations 021/022 + seq seeds + repos reviewed.
- [ ] `security-engineer` — PII logging + ownership isolation OK.
- [ ] `code-reviewer` — final APPROVE.
- [ ] `go build ./... && go vet ./... && go test ./...` green.
