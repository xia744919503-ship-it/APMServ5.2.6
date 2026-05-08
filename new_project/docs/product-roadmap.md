# Product Roadmap

## Product Goal

Target: rebuild the legacy RXSG single-player game as a playable product that matches the original shell, navigation rhythm, and core gameplay loop as closely as practical on the modern stack.

This means the project is no longer judged by "data can be read from the legacy database" alone.
It is judged by whether the rebuilt client behaves like the original game.

## Working Principles

- Reuse the current Go legacy-adapter work where it already matches real game data.
- Stop adding management-dashboard style pages that do not exist in the original product flow.
- Build and verify one playable phase at a time.
- Save a memory archive entry at the end of every phase.

## Reuse vs Rebuild

### Keep and build on

- Legacy database access in `backend/internal/legacy`
- Session bridge and city ownership checks
- World map, city detail, and production/tax adapter code where it matches real data
- Smoke script under `scripts/smoke-api.ps1`

### Rebuild or heavily reshape

- Frontend shell and navigation model
- Root overview/main scene
- City interaction flow
- World-to-city-to-battle transition flow
- Module coverage versus the original PHP command surface

## Phase Plan

### Phase 0 - Stable Baseline

Goal: restore a clean build and establish a product-facing execution plan.

Success criteria:

- Frontend source builds successfully
- Current work direction is documented in this file
- `docs/project-memory.md` has a stage entry for the reset

### Phase 1 - Legacy Shell Recreation

Goal: replace the current SPA/dashboard feel with a stage-like main shell closer to the original client.

Success criteria:

- Root screen acts as a game home scene, not a dashboard
- Navigation is centered on the current city and world transitions
- Main shell feels like one continuous game client surface

### Phase 2 - City Core Loop

Goal: rebuild the city interior around the original interaction model.

Success criteria:

- Building grid/state is shown in a city-first layout
- Core city actions are available in-place
- Resource, population, morale, and queues update coherently

### Phase 3 - World and Warfare Loop

Goal: make the world map part of a real gameplay chain.

Success criteria:

- World map supports scouting/selection into city and troop actions
- Marching, warnings, and reports are represented in the rebuilt flow
- Map is no longer just a read-only inspection screen

### Phase 4 - System Modules

Goal: cover the major original subsystems needed for a believable full product.

Target modules:

- heroes
- troops
- technology
- reports
- mail
- tasks
- union
- market
- ranking
- items / rewards / buffs

Success criteria:

- The rebuilt client covers the major daily-use modules from the legacy game
- Cross-module navigation follows the original play rhythm

### Phase 5 - Legacy Command Replacement

Goal: progressively replace legacy PHP execution paths with Go implementations without changing player-facing behavior.

Success criteria:

- High-traffic gameplay flows no longer depend on the original PHP command path
- Behavior stays compatible with existing saved data

## Phase Archive Rule

At the end of every completed phase, append a new entry to `docs/project-memory.md` with:

- date and local time
- current model information as exposed by the environment
- phase goal
- files touched
- verification result
- open risks and next step

## Current Execution Order - 2026-04-19

1. Stabilization gate - completed
   - auth path no longer pretends unsupported password login works
   - relation read/write path now matches the live schema
   - hero recruit roster path now survives the live `double` exp column
2. City core loop - next
   - building upgrades and queue visibility
   - troop training / recruit-capacity chain
   - technology research chain
   - city defence / wall operations
3. World and warfare loop - after city core loop
   - city-to-world dispatch rhythm
   - warning / report / callback loop tightening
4. Remaining system depth - after the core loops are stable
   - union
   - tasks
   - market
   - remaining legacy module gaps
