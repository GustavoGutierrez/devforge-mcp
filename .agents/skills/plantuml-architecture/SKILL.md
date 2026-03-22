---
name: plantuml-architecture
description: Generate a PlantUML component or deployment diagram (C4-like) for an application. Use when documenting system architecture, layer dependencies, or multi-node deployment topology. Activates on "plantuml-architecture", "architecture diagram", "component diagram", "deployment diagram".
argument-hint: <app description with components, layers, databases, external services>
user-invocable: true
---

# PlantUML Architecture Diagram

Generate a PlantUML architecture diagram for: **$ARGUMENTS**

Read `.claude/resources/plantuml-syntax.md` — sections **1. Architecture / Component Diagrams** and **5. Style Tips** — for syntax reference.

---

## Decision: Component vs Deployment

Choose the diagram type based on the input:

| Input mentions | Use |
|----------------|-----|
| Layers, modules, services, components, data flow | **Component diagram** (`package`, `component`, `-->`) |
| Servers, nodes, runtimes, cloud providers, deploy targets | **Deployment diagram** (`node`, `cloud`, `artifact`) |
| Both | Two separate `@startuml` blocks or composite with both |

---

## Component Diagram Template

Use for Clean Architecture layers, service boundaries, internal module structure.

```plantuml
@startuml architecture-components
!pragma layout smetana
skinparam shadowing false
skinparam componentStyle rectangle
skinparam defaultFontName Helvetica
skinparam defaultFontSize 13
skinparam component {
  BackgroundColor #EEF5FF
  BorderColor #4A90D9
  ArrowColor #444444
}

title Architecture — [App Name]

package "Presentation Layer" #FAFAFA {
  [Routes] as R
  [Controllers] as C
  [Middlewares] as MW
  [Zod Schemas] as ZS
}

package "Application Layer" #F0F4FF {
  [Use Cases] as UC
  [DTOs] as DTO
}

package "Domain Layer" #FFF8E1 {
  [Entities] as E
  [Repository Interfaces] as RI
  [Domain Errors] as DE
  [Value Objects] as VO
}

package "Infrastructure Layer" #F1F8E9 {
  [Repository Impls] as REPO
  [ORM / DB Client] as ORM
  [Services (JWT, Hash)] as SVC
  [DI Container] as DI
}

database "Database" as DB

R --> C : delegates
C --> UC : invokes
MW --> C : guards
UC --> RI : via interface
UC --> E : creates/reads
REPO ..|> RI : implements
REPO --> ORM : uses
ORM --> DB : SQL
DI --> UC : wires
DI --> REPO : wires
DI --> SVC : wires

note bottom of RI : Dependency points inward.\nDomain knows nothing of infra.
@enduml
```

## Deployment Diagram Template

Use for node topology, cloud providers, runtime environments.

```plantuml
@startuml architecture-deployment
!pragma layout smetana
skinparam shadowing false
skinparam defaultFontName Helvetica
skinparam defaultFontSize 13
skinparam node {
  BackgroundColor #E3F2FD
  BorderColor #1565C0
}
skinparam cloud {
  BackgroundColor #F3E5F5
  BorderColor #7B1FA2
}

title Deployment — [App Name]

actor "Client" as CL

node "Local / VPS" {
  artifact "Node.js Server" as NS
  database "SQLite\nbetter-sqlite3" as SQLITE
  NS --> SQLITE : file I/O
}

cloud "Cloudflare Edge" {
  artifact "Workers Bundle" as WB
  database "D1 Database" as D1
  WB --> D1 : D1 binding
}

CL --> NS : HTTPS :3000
CL --> WB : HTTPS (custom domain)

note right of WB : Same src/ codebase.\nDifferent entry point:\nsrc/worker.ts
note right of NS : Entry: src/index.ts\n@hono/node-server
@enduml
```

---

## Rules

0. **Always add `!pragma layout smetana`** as the first line after `@startuml` in every component and deployment diagram. Component diagrams require Graphviz (`dot`) by default; without this pragma PlantUML crashes with `Cannot run program "dot"` when Graphviz is not installed. Smetana is PlantUML's built-in pure-Java layout engine and has no external dependencies.
1. **Group by logical boundary** — use `package` for layers/services, `node` for runtime hosts.
2. **Label arrows** — every `-->` must describe the relationship (uses, calls, implements, queries).
3. **One diagram = one concern** — if both component + deployment are needed, create two files.
4. **No implementation details** — diagram shows structure, not code. Class names, not method names.
5. **Dependency rule visible** — outer layers arrow toward inner (Presentation → Application → Domain). Never reverse.
6. **Add a title** — always include `title` directive.
7. **Note strategic decisions** — use `note` for architectural constraints (e.g., "Domain knows nothing of infra").
8. **Never use `note right of` or inline `note` on components inside a `package` block.** Smetana does not support notes anchored to components that live inside packages — they cause rendering errors or invisible output. Place all notes **outside** any `package` block as standalone `note bottom of [Component] ... end note` directives.

---

## Output

Produce a complete, renderable `.puml` file:

```
@startuml [diagram-id]
' ... diagram content ...
@enduml
```

State the suggested save path:
- Component: `diagrams/architecture/[name].puml`
- Deployment: `diagrams/architecture/deployment.puml`

Then write the file to that path.
