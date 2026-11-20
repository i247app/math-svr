-- migration up
-- ma_llm_routes — which LLM provider (and optionally model) each kind of
-- LLM call goes through, editable at runtime by /llm-routes/* (admin).
--
-- One row per route_code. The codes are a catalog in code
-- (internal/domain/llmroute/catalog.go): rows are seeded
-- (seed/<ts>_ma_llm_routes.sql), never created or deleted through the API.
-- Read on every model call, no in-memory cache, so every instance sees an
-- update on its next call.
--
-- Resolution (ResolveLLMRouteQuery): the code's row → the DEFAULT row → env
-- LLM_PROVIDER. llm_provider NULL = fall through to the next step;
-- llm_model NULL = the provider's model from env (LLM_<X>_MODEL). A provider
-- not registered on the instance (no API key) also falls back to env.
-- llm_provider holds a provider name (openai / openrouter / gemini /
-- deepseek), checked against the registry by the API, not by the schema.
--
-- route_code is unique among live rows only (like ma_roles.role_code).
-- Config data: not a clear-data target.

CREATE TABLE IF NOT EXISTS ma_llm_routes (
  llm_route_id     BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  route_code       VARCHAR(64)  NOT NULL,
  llm_provider     VARCHAR(32)  DEFAULT NULL,
  llm_model        VARCHAR(128) DEFAULT NULL,
  description      VARCHAR(255) DEFAULT NULL,
  llm_route_status VARCHAR(32)  DEFAULT NULL,
  rpt_flg          VARCHAR(16)  DEFAULT NULL,
  kwords           VARCHAR(255) DEFAULT NULL,
  note             VARCHAR(500) DEFAULT NULL,
  status           VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id        BIGINT UNSIGNED DEFAULT NULL,
  create_dt        DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id        BIGINT UNSIGNED DEFAULT NULL,
  modify_dt        DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt       DATETIME(6)  DEFAULT NULL,
  UNIQUE KEY uk_live_route_code ((IF(deleted_dt IS NULL, route_code, NULL)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_llm_routes ADD PRIMARY KEY (llm_route_id);
--   CREATE UNIQUE INDEX uk_live_route_code ON ma_llm_routes ((IF(deleted_dt IS NULL, route_code, NULL)));
--
-- Drop:
--   ALTER TABLE ma_llm_routes DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_live_route_code ON ma_llm_routes;
