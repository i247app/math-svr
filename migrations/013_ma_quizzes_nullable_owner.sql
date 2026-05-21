-- migration up
-- Allow anonymous / profile-less quiz generation. When the client omits
-- profile_id the service can persist the quiz without an owner — only
-- the quiz_id is needed to fetch it back. user_id follows the same rule
-- since today it is derived from the profile.
ALTER TABLE ma_quizzes
  MODIFY COLUMN user_id    CHAR(36) DEFAULT NULL,
  MODIFY COLUMN profile_id CHAR(36) DEFAULT NULL;
