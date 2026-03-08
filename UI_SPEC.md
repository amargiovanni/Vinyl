# UI_SPEC.md — Vinyl Visual Design & Frontend Specification

## 1. Aesthetic Direction: Vintage Analogico

The UI evokes the warmth of a 1970s hi-fi listening room. Think: aged paper, warm amber light, vinyl textures, analog meters, hand-lettered labels. Every element should feel like it has weight, history, and warmth — the opposite of flat, clinical digital interfaces.

### Color Palette

```css
:root {
    /* Primary — Warm browns and ambers */
    --vinyl-black:       #1A1410;      /* Deep warm black, main background */
    --groove-dark:       #2A2018;      /* Vinyl groove dark */
    --cardboard:         #3D3024;      /* Card/container background */
    --leather:           #5C4A3A;      /* Secondary surfaces */
    --worn-wood:         #7A6450;      /* Borders, dividers */
    
    /* Accent — Golds and ambers */
    --amber-glow:        #C49A6C;      /* Primary accent, active states */
    --gold-bright:       #E8C496;      /* Highlights, emphasis */
    --cream:             #F2E6D0;      /* Primary text */
    --warm-white:        #FAF3E8;      /* Bright text, headings */
    
    /* Semantic — Muted vintage tones */
    --mood-happy:        #D4A96A;      /* Warm gold */
    --mood-calm:         #8BA48B;      /* Sage green */
    --mood-energetic:    #C47A5A;      /* Terracotta */
    --mood-sad:          #6B7E9E;      /* Dusty blue */
    --mood-nostalgic:    #9E7A8E;      /* Mauve */
    
    /* Heatmap scale */
    --heat-0:            #2A2018;
    --heat-1:            #5C4A3A;
    --heat-2:            #8B6D4F;
    --heat-3:            #C49A6C;
    --heat-4:            #E8C496;
    
    /* Utility */
    --spotify-green:     #7BA67B;      /* Desaturated Spotify green */
    --apple-pink:        #B87A8E;      /* Desaturated Apple pink */
    --text-muted:        #8E7E6E;      /* Secondary text */
    --text-faint:        #6E5E4E;      /* Tertiary text */
}
```

### Typography

```css
/* Import via Google Fonts */
@import url('https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,400;0,700;1,400&family=Source+Serif+4:ital,wght@0,300;0,400;0,600;1,400&family=JetBrains+Mono:wght@400&display=swap');

:root {
    --font-display:   'Playfair Display', Georgia, serif;   /* Headings, track names */
    --font-body:      'Source Serif 4', 'Palatino', serif;  /* Body text, labels */
    --font-mono:      'JetBrains Mono', monospace;          /* Timestamps, stats */
    
    --text-xs:    0.7rem;     /* 11px — timestamps, meta */
    --text-sm:    0.8rem;     /* 13px — secondary info */
    --text-base:  0.9rem;     /* 14px — body text */
    --text-lg:    1.05rem;    /* 17px — track titles */
    --text-xl:    1.3rem;     /* 21px — section headers */
    --text-2xl:   1.7rem;     /* 27px — view titles */
}
```

### Texture & Effects

```css
/* Noise overlay — apply to main container */
.vinyl-noise::after {
    content: '';
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background-image: url('data:image/svg+xml,...'); /* inline SVG noise pattern */
    opacity: 0.03;
    pointer-events: none;
    z-index: 9999;
}

/* Vignette — subtle darkening at edges */
.vinyl-vignette::before {
    content: '';
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: radial-gradient(ellipse at center, transparent 50%, rgba(15,10,5,0.4) 100%);
    pointer-events: none;
    z-index: 9998;
}

/* Card surface — slightly raised, warm */
.vinyl-card {
    background: var(--cardboard);
    border: 1px solid var(--worn-wood);
    border-radius: 8px;
    box-shadow: 
        0 1px 3px rgba(0,0,0,0.3),
        inset 0 1px 0 rgba(255,255,255,0.03);
}

/* Album art — Polaroid-style frame */
.album-art-frame {
    background: var(--cream);
    padding: 6px 6px 20px 6px;
    border-radius: 2px;
    box-shadow: 
        2px 3px 8px rgba(0,0,0,0.4),
        0 0 0 1px rgba(0,0,0,0.1);
    transform: rotate(-1.5deg);
}
```

---

## 2. Component Specifications

### 2.1 Vinyl Disc Animation

The centerpiece. A vinyl record that spins when music plays.

```
Dimensions: 120 × 120px in the popup
Structure:
├── Outer disc (120px circle, dark with groove lines)
│   ├── Concentric groove lines (subtle, 1px, semi-transparent)
│   ├── Label area (40px circle, centered)
│   │   ├── Album art (36px circle, clipped)
│   │   └── Center hole (4px circle, black)
│   └── Light reflection (gradient overlay, rotates independently)
└── Shadow (ellipse below, blurred)
```

**CSS Animation**:
```css
@keyframes spin {
    from { transform: rotate(0deg); }
    to   { transform: rotate(360deg); }
}

.vinyl-disc {
    animation: spin 1.8s linear infinite; /* ~33 RPM */
}

.vinyl-disc.paused {
    animation-play-state: paused;
}

/* Groove lines via repeating radial gradient */
.vinyl-grooves {
    background: repeating-radial-gradient(
        circle at center,
        transparent 20px,
        rgba(255,255,255,0.03) 20.5px,
        transparent 21px,
        transparent 23px
    );
}

/* Light reflection — slow independent rotation */
.vinyl-reflection {
    background: conic-gradient(
        from 0deg,
        transparent 0deg,
        rgba(255,255,255,0.06) 30deg,
        transparent 60deg,
        transparent 180deg,
        rgba(255,255,255,0.04) 210deg,
        transparent 240deg
    );
    animation: spin 8s linear infinite;
}
```

### 2.2 Now Playing View

```
┌─────────────────────────────────────┐
│                                      │
│        ┌──────────────────┐         │
│        │  [Album Art in   │         │
│        │   Polaroid frame │         │
│        │   with vinyl     │         │  ← Album art area
│        │   disc behind]   │         │     Disc spins, art overlaid
│        └──────────────────┘         │
│                                      │
│   Track Title                        │  ← Playfair Display, 17px, cream
│   Artist Name — Album Name           │  ← Source Serif 4, 13px, muted
│                                      │
│   ▶ ━━━━━━━━━━━━━━●━━━━━━━  3:42    │  ← Progress bar (amber glow)
│                                      │
│   ┌─ Mood ─────────────────────┐    │
│   │  😊  😌  🔥  😢  🤔  😤   │    │  ← Mood picker, horizontal scroll
│   │  🥰  🌙  💪  🎉  😴  🌊   │    │     Selected = amber glow ring
│   └────────────────────────────┘    │
│                                      │
│   ☁️ 18°C · Clouds · Pescara        │  ← Weather bar, small, mono font
│   Session: 47m · 8 tracks            │  ← Session info, mono font
│                                      │
│  ┌──┐  ┌──┐  ┌──┐  ┌──┐            │
│  │🎵│  │📖│  │📊│  │⚙️│            │  ← Tab bar (icons only)
│  └──┘  └──┘  └──┘  └──┘            │
└─────────────────────────────────────┘
```

**Key interactions**:
- Vinyl disc + album art: decorative, no interaction
- Progress bar: read-only (not a seek bar)
- Mood emoji: tap to select, tap again to deselect. Selected state = amber ring + slight scale up
- Tab bar: active tab underlined with amber

### 2.3 Diary View

```
┌─────────────────────────────────────┐
│  📖 Diary                 ◀ ▶ month │  ← Header with month navigation
│─────────────────────────────────────│
│                                      │
│  ┌─ Heatmap ───────────────────┐    │
│  │  [52-week calendar grid]     │    │  ← Compact, 280px wide
│  │  M · · · · · · · · · ·      │    │
│  │  T · · · · · · · · · ·      │    │
│  │  W · · · · · · · · · ·      │    │
│  │  T · · · · · · · · · ·      │    │
│  │  F · · · · · · · · · ·      │    │
│  │  S · · · · · · · · · ·      │    │
│  │  S · · · · · · · · · ·      │    │
│  └─────────────────────────────┘    │
│                                      │
│  Today · March 8, 2026              │  ← Date header, Playfair
│  ─────────────────────────────      │
│  ┌─ Session ───────────────────┐    │
│  │  14:30 — 16:45  ·  2h 15m   │    │  ← Time range + duration
│  │  12 tracks  ·  🌊  ·  ☁️ 17° │   │  ← Count + mood + weather
│  │  ♫ Top: "Karma Police"       │    │  ← Top track
│  │  ○ Spotify                    │    │  ← Source badge
│  └──────────────────────────────┘   │
│                                      │
│  ┌─ Session ───────────────────┐    │
│  │  09:00 — 10:15  ·  1h 15m   │    │
│  │  7 tracks  ·  😌  ·  ☀️ 14°  │   │
│  │  ♫ Top: "Clair de Lune"      │    │
│  │  ○ Apple Music                │    │
│  └──────────────────────────────┘   │
│                                      │
│  Yesterday · March 7, 2026          │
│  ─────────────────────────────      │
│  ...                                 │
│                                      │
│  ┌──┐  ┌──┐  ┌──┐  ┌──┐            │
│  │🎵│  │📖│  │📊│  │⚙️│            │
│  └──┘  └──┘  └──┘  └──┘            │
└─────────────────────────────────────┘
```

**Session card expand** (on click):
Shows full track list below the card summary:
```
│  1. Karma Police — Radiohead          4:21  │
│  2. Lucky — Radiohead                 4:19  │
│  3. No Surprises — Radiohead          3:48  │
│  ...                                         │
```

### 2.4 Insights View (P1, if time)

```
┌─────────────────────────────────────┐
│  📊 Insights              2026      │
│─────────────────────────────────────│
│                                      │
│  ┌─ Insight Card ──────────────┐    │
│  │  🌧️ "You listen to jazz      │    │
│  │      when it rains"           │    │
│  │  ─────────────────────        │    │
│  │  32 rainy sessions            │    │
│  │  Jazz: 68% · Other: 32%      │    │
│  └──────────────────────────────┘   │
│                                      │
│  ┌─ Your Week ─────────────────┐    │
│  │  Mon ████████░░  4.2h         │    │
│  │  Tue ██████░░░░  3.1h         │    │
│  │  Wed █████████░  4.8h         │    │  ← Horizontal bars
│  │  Thu ███░░░░░░░  1.5h         │    │
│  │  Fri ████████████  6.0h       │    │
│  │  Sat ██████████░  5.2h        │    │
│  │  Sun ███████░░░  3.8h         │    │
│  └──────────────────────────────┘   │
│                                      │
│  ┌─ Mood Journey ──────────────┐    │
│  │  [Sparkline showing dominant  │    │
│  │   mood per week over time]    │    │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

### 2.5 Settings View

```
┌─────────────────────────────────────┐
│  ⚙️ Settings                         │
│─────────────────────────────────────│
│                                      │
│  Location                            │
│  ┌──────────────────────────────┐   │
│  │  Pescara, IT                  │   │  ← Editable text field
│  │  42.4612° N, 14.2111° E      │   │  ← Auto-resolved coords
│  └──────────────────────────────┘   │
│                                      │
│  Weather API Key                     │
│  ┌──────────────────────────────┐   │
│  │  ●●●●●●●●●●●●dk4f           │   │  ← Masked input
│  └──────────────────────────────┘   │
│  ℹ️ Get free key at openweathermap   │
│                                      │
│  Spotify                             │
│  ┌──────────────────────────────┐   │
│  │  ✓ Connected as andrea.m      │   │  ← Green checkmark
│  │  [Disconnect]                 │   │  ← Danger button
│  └──────────────────────────────┘   │
│     — or —                           │
│  ┌──────────────────────────────┐   │
│  │  Not connected                │   │
│  │  [Connect Spotify]            │   │  ← Primary button
│  └──────────────────────────────┘   │
│                                      │
│  Apple Music                         │
│  ┌──────────────────────────────┐   │
│  │  ✓ Detected automatically     │   │  ← Always enabled
│  └──────────────────────────────┘   │
│                                      │
│  Export                              │
│  [Export 2025 Annual Book]           │  ← Button, opens ExportWizard
│  [Export 2026 Annual Book]           │
│                                      │
│  Data                                │
│  Database: ~/Library/.../vinyl.db    │
│  Size: 12.4 MB · 847 sessions       │
│  [Open Database Folder]             │
│                                      │
│  ─────────────────────────────      │
│  Vinyl v1.0.0                        │
│  Made with ♫ in Pescara              │
│                                      │
│  ┌──┐  ┌──┐  ┌──┐  ┌──┐            │
│  │🎵│  │📖│  │📊│  │⚙️│            │
│  └──┘  └──┘  └──┘  └──┘            │
└─────────────────────────────────────┘
```

---

## 3. Animations & Micro-interactions

| Element | Animation | Duration | Easing |
|---------|-----------|----------|--------|
| Vinyl disc spin | Continuous rotation | 1.8s per revolution | linear |
| View transitions | Slide left/right | 200ms | ease-out |
| Session card expand | Height expand + fade in tracks | 250ms | ease-in-out |
| Mood select | Scale 1.0→1.15 + amber ring fade in | 150ms | ease-out |
| Heatmap cell hover | Slight brightness increase + tooltip | 100ms | ease-out |
| New session toast | Slide down from top | 300ms | ease-out, then auto-dismiss 3s |
| Tab switch | Underline slide | 200ms | ease-in-out |
| Album art load | Fade in from 0 opacity | 400ms | ease-in |
| Progress bar | Smooth width transition | 1000ms | linear |

---

## 4. Responsive Considerations

The popup is fixed at **380 × 520px**. No responsive breakpoints needed. But:
- Heatmap may need horizontal scroll on smaller year ranges
- Track lists in session expand are scrollable (max-height: 200px)
- Long track/artist names: truncate with ellipsis
- Session list: virtual scrolling if >100 sessions visible (use Svelte `each` with keyed blocks)

---

## 5. Accessibility

- All interactive elements have focus states (amber outline)
- Mood emojis have `aria-label` with mood name
- Heatmap cells have `aria-label` with date and listen time
- Color is never the only indicator (shapes/labels supplement)
- Tab navigation works throughout the popup
- Minimum touch/click target: 32×32px
- Reduced motion: respect `prefers-reduced-motion` — disable vinyl spin, use fade instead of slide

---

## 6. Assets Needed

| Asset | Format | Size | Notes |
|-------|--------|------|-------|
| Tray icon (idle) | PNG template | 16×16 @1x, 32×32 @2x | macOS template image (black, system handles dark/light) |
| Tray icon frames (playing) | PNG template × 4 | 16×16 @1x each | 4 rotation frames for animation |
| App icon | PNG | 512×512, 1024×1024 | Vinyl record, warm tones, for About/Dock |
| Noise texture | SVG (inline) | — | Procedural, embedded in CSS |
| Weather icons | SVG set | 24×24 | Custom vintage-style, or use emoji |
| Source badges | SVG | 16×16 | Simplified Spotify/Apple Music marks |
| Mood emojis | System emoji | — | Use native system emoji rendering |

---

## 7. Annual Book Export — Visual Spec

The exported HTML book should feel like a beautifully printed music yearbook.

### Paper feel:
- Background: `#FAF3E8` (warm white, like aged paper)
- Text: `#2A2018` (warm black)
- Accents: `#C49A6C` (amber)
- Page breaks between months via `page-break-before: always`
- Print margins: 2cm all around

### Cover page:
```
┌─────────────────────────────────┐
│                                  │
│                                  │
│          V I N Y L               │  ← Playfair Display, 48pt, spaced
│                                  │
│           2025                   │  ← Playfair Display, 72pt, amber
│                                  │
│      ┌──────────────┐           │
│      │  [Decorative  │           │  ← Generated pattern based on
│      │   circular    │           │     top genres/moods (SVG)
│      │   pattern]    │           │
│      └──────────────┘           │
│                                  │
│    1,247 hours · 3,891 sessions  │  ← Source Serif, 14pt, muted
│    892 artists · 4,567 tracks    │
│                                  │
│                                  │
│    Top Artists                   │
│    1. Radiohead                  │
│    2. Thelonious Monk            │
│    3. Bon Iver                   │
│                                  │
│                                  │
│                                  │
└─────────────────────────────────┘
```

### Month pages:
- Clean, editorial layout
- Stats in a grid: 3 columns
- Top lists with subtle numbering
- Mood distribution as simple horizontal bars (CSS only, no JS charts)
- Mini heatmap for the month (7 × ~5 grid)
- Pull quotes from notable sessions: "March 15 — 4 hours of Miles Davis during a thunderstorm"
