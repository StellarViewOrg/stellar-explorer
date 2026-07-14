# Contributing to StellarView Explorer

Thank you for your interest in contributing. This guide covers the essentials for getting your changes merged.

---

## Development setup

1. **Fork and clone** the repository.
2. Install dependencies with Bun:
   ```bash
   bun install
   ```
3. Start the frontend dev server:
   ```bash
   bun run dev:web   # http://localhost:3000
   ```

This repository holds the **web app** only. The indexer, TUI, and docs site each live in their own repo with their own contributing guide:

- [`StellarViewOrg/indexer`](https://github.com/StellarViewOrg/indexer)
- [`StellarViewOrg/tui`](https://github.com/StellarViewOrg/tui)
- [`StellarViewOrg/docs`](https://github.com/StellarViewOrg/docs)

---

## Project structure

```text
apps/explorer-web/     # Next.js frontend (the only app in this repo)
```

---

## Making changes

### Code style

```bash
bun run lint           # ESLint
bun run format         # Prettier (auto-fix)
bun run format:check   # Prettier (CI check)
```

Prettier config: double quotes, semicolons, 100-char line width, trailing commas (ES5), Tailwind class sorting.

### Tests

```bash
bun run test           # Run all frontend tests (Vitest)
```

Tests use the `happy-dom` environment. Place test files alongside source as `*.test.{ts,tsx}`.

### Internationalization

The UI supports 9 locales (EN, ES, PT, FR, DE, ZH, JA, KO, RU). CI validates that all translation files have **identical key counts**. When adding UI text:

1. Add the key to `messages/en.json`.
2. Add the same key (translated or as a placeholder) to all other `messages/*.json` files.

---

## Pull request checklist

Before opening a PR:

- [ ] `bun run lint` passes with no errors
- [ ] `bun run format:check` passes
- [ ] `bun run test` passes
- [ ] All 9 locale files updated if UI text was added
- [ ] The change is described clearly in the PR description

---

## Reporting issues

Open an issue on GitHub with:

- A clear description of the problem
- Steps to reproduce
- Expected vs. actual behavior
- Network (public / testnet / futurenet) and browser/OS if relevant

---

## License

By contributing, you agree that your changes will be licensed under the [MIT License](./LICENSE).
