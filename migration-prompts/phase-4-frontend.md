# Phase 4: SvelteKit Frontend

**Duration:** 10-14 days
**Goal:** Beautiful web UI with SvelteKit
**Dependencies:** Phase 3 complete (API running)

## Context

Build modern web interface using SvelteKit, Tailwind CSS, and Skeleton UI.

## Pages to Build

- `/` - Dashboard (idea list with filters)
- `/ideas/[id]` - Idea detail view
- `/ideas/[id]/edit` - Edit idea
- `/analytics` - Charts and statistics
- `/settings` - Configuration

## Setup

```bash
cd web
npm create svelte@latest .
npm install -D tailwindcss @skeletonlabs/skeleton
npm install lucide-svelte @tanstack/svelte-query
npm install -D @playwright/test
```

## Components

1. IdeaCard - Display idea with score
2. IdeaForm - Create/edit ideas
3. FilterBar - Filter by score/status
4. ScoreChart - Visualize analytics
5. PatternBadge - Display detected patterns

## TDD Approach

Component tests with Vitest:
```typescript
test('IdeaCard renders correctly', () => {
	const idea = { id: '1', title: 'Test', score: 8.5 };
	const { getByText } = render(IdeaCard, { props: { idea } });
	expect(getByText('Test')).toBeInTheDocument();
	expect(getByText('8.5')).toBeInTheDocument();
});
```

E2E tests with Playwright:
```typescript
test('create idea flow', async ({ page }) => {
	await page.goto('/');
	await page.fill('input[placeholder="Your idea"]', 'Test Idea');
	await page.click('button:has-text("Create")');
	await expect(page.locator('text=Test Idea')).toBeVisible();
});
```

## Deliverables

- [ ] All pages implemented
- [ ] Components tested
- [ ] E2E tests passing
- [ ] Responsive design
- [ ] Type-safe API client

## Validation

```bash
cd web
npm run dev       # Dev server
npm test          # Component tests  
npx playwright test  # E2E tests
npm run build     # Production build
```

## Success Criteria

✅ Beautiful, functional UI
✅ All features working
✅ Tests passing
✅ Mobile responsive
✅ Production build succeeds
