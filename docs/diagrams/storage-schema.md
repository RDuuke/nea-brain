# Storage Schema Diagram Spec

This diagram reflects the SQLite schema and FTS table definitions.

## Mermaid
```mermaid
erDiagram
  SCHEMA_MIGRATIONS {
    INTEGER version PK
    TEXT name
    TEXT applied_at
  }

  OBSERVATIONS {
    TEXT id PK
    TEXT content
    TEXT created_at
    TEXT updated_at
    TEXT deleted_at
    TEXT project
    TEXT topic_key
    TEXT tags
    TEXT source
    TEXT metadata
  }

  TOPICS {
    TEXT id PK
    TEXT topic_key UNIQUE
    TEXT name
    TEXT description
    TEXT metadata
    TEXT created_at
    TEXT updated_at
  }

  SESSIONS {
    TEXT id PK
    TEXT created_at
    TEXT last_seen_at
    TEXT disclosure_level
    TEXT recent_operations
  }

  DUPLICATES {
    TEXT id PK
    TEXT original_observation_id
    TEXT duplicate_observation_id
    TEXT reason
    TEXT created_at
  }

  OBSERVATIONS_FTS {
    TEXT content
    TEXT observation_id
  }

  OBSERVATIONS ||--o{ DUPLICATES : original
  OBSERVATIONS ||--o{ DUPLICATES : duplicate
  OBSERVATIONS ||--o{ OBSERVATIONS_FTS : indexed
  TOPICS ||--o{ OBSERVATIONS : topic_key
```

## Pencil MCP nodes
Nodes:
- schema_migrations
- observations
- topics
- sessions
- duplicates
- observations_fts

Connections:
- observations.id -> duplicates.original_observation_id
- observations.id -> duplicates.duplicate_observation_id
- observations.id -> observations_fts.observation_id
- topics.topic_key -> observations.topic_key
