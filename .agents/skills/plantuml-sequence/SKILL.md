---
name: plantuml-sequence
description: Generate a PlantUML sequence diagram for an API or system flow. Use when documenting request-response lifecycles, authentication flows, inter-service communication, or any step-by-step interaction between actors. Activates on "plantuml-sequence", "sequence diagram", "flow diagram", "interaction diagram".
argument-hint: <flow name, actors, steps, alt/error paths>
user-invocable: true
---

# PlantUML Sequence Diagram

Generate a PlantUML sequence diagram for: **$ARGUMENTS**

Read `.claude/resources/plantuml-syntax.md` — section **2. Sequence Diagrams** — for syntax reference.

---

## Template

```plantuml
@startuml [flow-name]
skinparam shadowing false
skinparam defaultFontName Helvetica
skinparam defaultFontSize 13
skinparam sequence {
  ParticipantBackgroundColor #EEF5FF
  ParticipantBorderColor #4A90D9
  ArrowColor #444444
  LifeLineBorderColor #AAAAAA
  LifeLineBackgroundColor #FAFAFA
  NoteBackgroundColor #FFFDE7
  NoteBorderColor #F9A825
}

title [Flow Title]
autonumber

actor "Client" as CL
participant "Controller" as CTRL #EEF5FF
participant "UseCase" as UC #F0F4FF
participant "Repository" as REPO #F1F8E9
database "Database" as DB #FFF8E1

CL -> CTRL : POST /endpoint {payload}
activate CTRL

CTRL -> CTRL : validate with Zod
note right : Returns 400 if invalid

CTRL -> UC : execute(dto)
activate UC

UC -> REPO : findBy(criteria)
activate REPO

REPO -> DB : SELECT query
DB --> REPO : row(s)
deactivate REPO

alt Success path
  UC -> REPO : save(entity)
  REPO -> DB : INSERT / UPDATE
  DB --> REPO : ok
  REPO --> UC : savedEntity
  UC --> CTRL : result DTO
  deactivate UC
  CTRL --> CL : 201 Created {data}
  deactivate CTRL
else Entity not found
  UC --> CTRL : NotFoundError
  deactivate UC
  CTRL --> CL : 404 Not Found
  deactivate CTRL
else Conflict / validation
  UC --> CTRL : DomainError
  deactivate UC
  CTRL --> CL : 409 Conflict
  deactivate CTRL
end
@enduml
```

---

## Rules

1. **Always use `autonumber`** — step numbers make flows readable.
2. **`activate` / `deactivate` every participant** — shows lifeline scope clearly.
3. **Name every arrow** — `->` and `-->` must always have a label describing the message.
4. **Solid arrows for calls, dashed for replies** — `A -> B : call` / `B --> A : reply`.
5. **Use `alt / else / end`** for all branching paths — never omit error paths.
6. **Use `loop` for repeated operations** — e.g., retry, batch processing.
7. **Add `note` for non-obvious decisions** — middleware checks, hashing, token rotation.
8. **Keep participants minimal** — only show the actors/participants relevant to this specific flow.
9. **Activate scope matches work** — activate on entry, deactivate on response returned.
10. **Use `box` grouping** for multi-service flows:
    ```plantuml
    box "API Server" #EEF5FF
      participant Controller
      participant UseCase
    end box
    box "Storage" #F1F8E9
      participant Repository
      database DB
    end box
    ```

---

## Auth-Specific Additions

For authentication/JWT flows, include these participants and notes:

```plantuml
participant "JwtMiddleware" as JWT #FFE0B2
participant "HashService" as HASH #F3E5F5
participant "TokenService" as TOKEN #E8F5E9

note over JWT : Verifies Bearer token\nRejects with 401 if invalid
note over TOKEN : signAccessToken (15 min)\nsignRefreshToken (7 days, jti=UUID)
note over HASH : bcrypt.compare for passwords\nSHA-256 for refresh token DB lookup
```

---

## Output

Produce a complete, renderable `.puml` file starting with `@startuml` and ending with `@enduml`.

State the suggested save path: `diagrams/flows/[flow-name].puml`

Then write the file to that path.
