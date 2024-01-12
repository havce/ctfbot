CREATE TABLE IF NOT EXISTS ctfs (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL UNIQUE,
  start       TEXT NOT NULL,
  player_role TEXT NOT NULL UNIQUE,
  can_join    BOOLEAN NOT NULL DEFAULT 1,
  ctftime_url TEXT NOT NULL,
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);
