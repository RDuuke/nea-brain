# Hexagonal Architecture Diagram Spec

This diagram describes the ports and adapters layout for NeaBrain.

## Mermaid
```mermaid
flowchart LR
  subgraph Inbound_Adapters
    CLI[CLI]
    HTTP[HTTP]
    MCP[MCP]
    TUI[TUI]
  end

  subgraph Inbound_Ports
    ObservationService[ObservationService]
    SearchService[SearchService]
    TopicService[TopicService]
    SessionService[SessionService]
    ConfigService[ConfigService]
  end

  subgraph Core_Domain
    Observation[Observation]
    Topic[Topic]
    Session[Session]
    Duplicate[Duplicate]
  end

  subgraph Outbound_Ports
    ObservationRepo[ObservationRepository]
    TopicRepo[TopicRepository]
    SessionRepo[SessionRepository]
    DuplicateRepo[DuplicateRepository]
    SearchIndex[SearchIndex]
    Clock[Clock]
    ConfigStore[ConfigStore]
  end

  subgraph Outbound_Adapters
    SqliteRepo[SQLite Repositories]
    SqliteFTS[SQLite FTS]
    ConfigAdapter[Config Adapter]
    ClockAdapter[Clock Adapter]
  end

  CLI --> Inbound_Ports
  HTTP --> Inbound_Ports
  MCP --> Inbound_Ports
  TUI --> Inbound_Ports

  Inbound_Ports --> Core_Domain
  Core_Domain --> Outbound_Ports

  ObservationRepo --> SqliteRepo
  TopicRepo --> SqliteRepo
  SessionRepo --> SqliteRepo
  DuplicateRepo --> SqliteRepo
  SearchIndex --> SqliteFTS
  ConfigStore --> ConfigAdapter
  Clock --> ClockAdapter
```

## Pencil MCP nodes
Nodes:
- Inbound adapters: CLI, HTTP, MCP, TUI
- Inbound ports: ObservationService, SearchService, TopicService, SessionService, ConfigService
- Core entities: Observation, Topic, Session, Duplicate
- Outbound ports: ObservationRepository, TopicRepository, SessionRepository, DuplicateRepository, SearchIndex, Clock, ConfigStore
- Outbound adapters: SQLite Repositories, SQLite FTS, Config Adapter, Clock Adapter

Connections:
- Inbound adapters -> inbound ports -> core entities -> outbound ports -> outbound adapters
