---
name: plantuml-state
description: Generate a PlantUML state diagram for an entity lifecycle, token states, or deployment pipeline. Use when documenting state machines, CI/CD pipelines, token rotation flows, or any system with well-defined states and transitions. Activates on "plantuml-state", "state diagram", "state machine", "lifecycle diagram", "deployment states".
argument-hint: <entity or system name, states, transitions/events, guards>
user-invocable: true
---

# PlantUML State Diagram

Generate a PlantUML state diagram for: **$ARGUMENTS**

Read `.claude/resources/plantuml-syntax.md` — section **4. State Diagrams** — for syntax reference.

---

## Template

```plantuml
@startuml [state-name]
!pragma layout smetana
skinparam shadowing false
skinparam defaultFontName Helvetica
skinparam defaultFontSize 13
skinparam state {
  BackgroundColor #FFF8E1
  BorderColor #F9A825
  ArrowColor #444444
  StartColor #2E7D32
  EndColor #C62828
}

title [Entity/System] — State Lifecycle

[*] --> InitialState : trigger / event

InitialState --> ActiveState : transition label
note right of InitialState : Describe constraint\nor guard condition

ActiveState --> SuccessState : success event
ActiveState --> FailedState : error event
ActiveState --> [*] : terminal condition

FailedState --> InitialState : retry / reset
SuccessState --> [*]

state ActiveState {
  [*] --> SubStateA
  SubStateA --> SubStateB : internal step
  SubStateB --> [*]
}
@enduml
```

---

## CI/CD Pipeline Template

For deployment pipeline state machines:

```plantuml
@startuml deployment-states
!pragma layout smetana
skinparam shadowing false
skinparam defaultFontName Helvetica
skinparam defaultFontSize 13
skinparam state {
  BackgroundColor #E3F2FD
  BorderColor #1565C0
  ArrowColor #444444
  StartColor #2E7D32
  EndColor #C62828
}

title [Service] — Deployment Pipeline States

[*] --> Development : feature branch created

Development --> Building : git push / PR opened
note right of Development : Local dev server\npnpm dev

Building --> BuildFailed : lint / type / test error
BuildFailed --> Development : fix & push

Building --> Staged : pnpm validate passed
note right of Staged : Wrangler dev or\nNode.js local build

Staged --> Deploying : deploy command triggered
Deploying --> Live : deploy succeeded
Deploying --> Rollback : deploy failed

Live --> Updating : new deploy triggered
Updating --> Live : deploy OK
Updating --> Rollback : deploy failed

Rollback --> Staged : wrangler rollback / revert

Live --> [*] : service deprecated

state Building {
  [*] --> Linting
  Linting --> TypeChecking
  TypeChecking --> Testing
  Testing --> Bundling
  Bundling --> [*]
}
@enduml
```

---

## Token Lifecycle Template

For JWT or session token state machines:

```plantuml
@startuml token-lifecycle
!pragma layout smetana
skinparam shadowing false
skinparam defaultFontName Helvetica
skinparam defaultFontSize 13
skinparam state {
  BackgroundColor #F3E5F5
  BorderColor #7B1FA2
  ArrowColor #444444
  StartColor #2E7D32
  EndColor #C62828
}

title JWT Token — Lifecycle

[*] --> Issued : sign (login / refresh)
note right of Issued : AccessToken: HS256, 15 min\nRefreshToken: HS256, 7 days, jti=UUID

Issued --> Active : used in Authorization header
Active --> Expired : TTL elapsed
Issued --> Revoked : explicit logout / changePassword
Active --> Revoked : explicit logout

Expired --> [*] : discarded
Revoked --> [*] : discarded

Active --> Refreshed : POST /auth/refresh (RefreshToken only)
Refreshed --> [*] : old token revoked
note right of Refreshed : New Issued token created\nOld token marked revoked in DB\nRotation guarantees single-use

state Active {
  [*] --> ValidSignature
  ValidSignature --> NotExpired
  NotExpired --> NotRevoked
  NotRevoked --> [*]
}
@enduml
```

---

## Entity Lifecycle Template

For domain entity state machines (order, product, user, etc.):

```plantuml
@startuml entity-lifecycle
!pragma layout smetana
skinparam shadowing false

title [Entity] — Domain Lifecycle

[*] --> Draft : create()

Draft --> Active : activate() / publish()
Draft --> Cancelled : cancel()

Active --> Suspended : suspend() [admin action]
Active --> Completed : complete() [business rule met]
Active --> Cancelled : cancel() [within allowed window]

Suspended --> Active : restore()
Suspended --> Cancelled : cancel()

Completed --> [*]
Cancelled --> [*]

note right of Active : Primary usable state.\nMost operations allowed here.
note right of Cancelled : Soft-delete pattern.\nData retained for audit.
@enduml
```

---

## Rules

0. **Always add `!pragma layout smetana`** as the second line (right after `@startuml`) in every state diagram. State diagrams use Graphviz (`dot`) by default; without this pragma PlantUML crashes with `Cannot run program "dot"` when Graphviz is not installed. Smetana is PlantUML's built-in pure-Java layout engine with no external dependencies.
1. **Always start with `[*]`** — initial pseudo-state must be the entry point.
2. **Keep transition labels short — one line only.** Never embed `\n` or multi-line text in arrow labels (`StateA --> StateB : long text\nmore text`). Long labels cause text overlap and unreadable diagrams with Smetana. Move details to a `note right of StateName` block instead.
3. **Never place `note` blocks inside `state { }` composite blocks.** The syntax `state X { note : text }` is not supported by Smetana and produces garbled output. Always use external notes: `note right of X ... end note`.
2. **Always have at least one `State --> [*]`** — every state machine needs terminal states.
3. **Label every transition** — `StateA --> StateB : event [guard] / action`.
4. **Use composite states for sub-flows** — `state Name { ... }` for states with internal steps.
5. **Add notes for constraints** — `note right of StateName : guard or business rule`.
6. **Color terminal states** — success states green `#LightGreen`, failure/cancelled states red `#LightCoral`.
7. **Keep transitions unidirectional unless explicitly bidirectional** — avoid visual clutter.
8. **Name the diagram after the entity or process** — `@startuml jwt-token-lifecycle`.
9. **Separate concerns** — one state machine per entity/token/pipeline. Don't mix token + deployment.
10. **Parallel regions** for concurrent sub-states:
    ```plantuml
    state Processing {
      [*] --> EmailVerification
      --
      [*] --> PaymentValidation
    }
    ```

---

## Output

Produce a complete, renderable `.puml` file starting with `@startuml` and ending with `@enduml`.

State the suggested save path: `diagrams/states/[state-name].puml`

Then write the file to that path.
