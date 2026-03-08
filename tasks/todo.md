# Vinyl — Implementation Plan

## Phase 1: Project Scaffold
- [x] Task 1: Initialize Wails project, Go module, frontend with Svelte 4 + Tailwind 3
- [x] Task 2: Configuration system (`internal/config/config.go`)
- [x] Task 3: Database setup with SQLite + migrations (`internal/diary/`)
- [x] Task 4: Frontend shell — tab navigation, vintage theme, placeholder views

## Phase 2: Music Detection
- [x] Task 5: Apple Music detection via AppleScript
- [x] Task 6: Spotify detection via AppleScript (fallback)
- [x] Task 7: Player detection orchestrator with polling
- [x] Task 8: Now Playing UI (vinyl disc, track info, progress bar)

## Phase 3: Session Tracking
- [x] Task 9: Session lifecycle engine (state machine)
- [x] Task 10: Weather integration (OpenWeatherMap + browser geolocation)
- [x] Task 11: Mood tagging
- [x] Task 12: Spotify OAuth PKCE + rich metadata

## Phase 4: Diary & History
- [x] Task 13: Daily stats aggregation
- [x] Task 14: Heatmap calendar component
- [x] Task 15: Session history / Diary view
- [x] Task 16: Artist stats

## Phase 5: Export & Polish
- [x] Task 17: Annual book — data gathering
- [x] Task 18: Annual book — HTML generation
- [x] Task 19: Export UI
- [x] Task 20: Settings view complete

## Phase 6: System Integration
- [x] Task 23: First run & onboarding

## Phase 7: Quality
- [x] Task 24: Error handling & structured logging (slog throughout)
- [x] Task 26: Tests (config, diary models, store CRUD, player types)

## Remaining (nice-to-have for v1.1)
- [ ] Task 21: Context menu (right-click tray) — requires Wails systray API
- [ ] Task 22: Tray icon animation — requires custom tray icon frames
- [ ] Task 25: Album art cache with local file storage + pruning
