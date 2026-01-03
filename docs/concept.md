# Project Concept: wips-cli

## 1. Basic Concept
**wips-cli (wip command)** is a "Local-first Event Logger for Developers".

### Core Value
- Centralized management of developer's daily activities (manual notes + automatic git commit capture) across projects/repositories in chronological order.
- **Local-first & Privacy-focused**: All data is stored locally.
- **"What did I do today?"**: Automatic/Manual recording -> Daily report aid, past search, context restoration.

### Data Structure
- **Storage**: NDJSON for event storage.
- **Context**: Dictionary for context normalization (repo/cwd/env).

### Uniqueness
- The combination of **Git commit automatic capture + Manual note chronological integration + Automatic context attachment** provides a unique value proposition not found in competitors like `journalot` or `git-standup`.

## 2. Monetization Strategy
**Priority: Success as OSS first.**

### Phase 1: OSS Growth
- Release CLI as free OSS.
- Goal: Gather GitHub stars and increase user base.
- Monetization: **GitHub Sponsors** (donations).

### Phase 2: Web Client (Freemium)
- Develop a centralized Web Client (Next.js).
- **Free Plan**: Sync limit (e.g., 1-2 devices).
- **Pro Plan ($30-50/year)**: Unlimited sync, advanced search, AI summary, enhanced export.
- **Team Plan ($10-20/month/user)**: Team dashboard, sharing, admin features.
- **Policy**: Cloud sync is optional; E2EE (End-to-End Encryption) is mandatory. Reference: Obsidian Sync model.

## 3. Positioning & Marketing
- **Strengths**: Solves "git-standup lacks context" and "jrnl doesn't track git".
- **Differentiation**: Automation + Deep Context + Integrated Review.
- **Marketing Channels**: Reddit (r/commandline, r/golang), Hacker News, Zenn/Qiita, X (Demo videos).
- **Key Message**: "Privacy Control Perfect", "Local-first", "git-standup + jrnl Upward Compatibility".

## 4. Data Management Design
- **Default**: Centralized management (`~/Library/Application Support/wips/` or similar).
    - Merits: Strong cross-project search, easy backup/sync, logs persist after repo deletion.
- **Control & Privacy**:
    - **Global Config**: `~/.wip/config.toml` (ignore/only_paths).
    - **Per-Repo**: `.wipignore` for exclusion.
    - **Safety**: Warning/Confirm options on recording.
    - **Git Hooks**: Opt-in (`wip hooks install`).

## 5. Feature Roadmap (OSS Release)
### High Priority (Must-haves)
1.  **Search**: `wip search` (Full text + Date/Repo/Type filters).
2.  **Export**: `wip export --format md` (Markdown daily report).
3.  **Visuals**: Colored output, table formatting.
4.  **UX**: Easy install (Homebrew), refined README (GIFs, Quickstart).

### Medium Priority
- Tag support.
- Statistics (`wip stats`).
- **Compatibility Modes**:
    - `wip standup` (git-standup compatibility).
    - `jrnl` style input.
