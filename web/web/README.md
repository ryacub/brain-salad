# Telos Idea Matrix - Web Frontend

Beautiful SvelteKit-based web interface for the Telos Idea Matrix project.

## Features

- **Dashboard** (`/`) - View and manage your ideas with filtering
- **Idea Details** (`/ideas/[id]`) - Detailed view of individual ideas with score breakdown
- **Edit Ideas** (`/ideas/[id]/edit`) - Update existing ideas
- **Analytics** (`/analytics`) - View statistics and insights about your ideas
- **Settings** (`/settings`) - Configure application settings

## Components

- `IdeaCard` - Display idea cards with scores and patterns
- `IdeaForm` - Create and edit ideas
- `FilterBar` - Filter ideas by status and score
- `ScoreChart` - Visualize analytics data
- `PatternBadge` - Display detected patterns

## Setup

```bash
cd web
npm install
```

## Development

```bash
npm run dev
```

The app will be available at `http://localhost:5173`

## Environment Variables

Create a `.env` file:

```
VITE_API_URL=http://localhost:8080
```

## Testing

```bash
# Run component tests
npm test

# Run component tests with UI
npm run test:ui

# Run E2E tests
npm run test:e2e

# Run E2E tests with UI
npm run test:e2e:ui
```

## Building

```bash
npm run build
npm run preview
```

## Tech Stack

- **SvelteKit** - Web framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **TanStack Query** - Data fetching and caching
- **Lucide Svelte** - Icons
- **Vitest** - Component testing
- **Playwright** - E2E testing

## API Integration

The frontend communicates with the Go API server at `localhost:8080` (configurable via `VITE_API_URL`).

### API Endpoints

- `POST /api/v1/ideas` - Create idea
- `GET /api/v1/ideas` - List ideas
- `GET /api/v1/ideas/{id}` - Get idea
- `PUT /api/v1/ideas/{id}` - Update idea
- `DELETE /api/v1/ideas/{id}` - Delete idea
- `POST /api/v1/analyze` - Analyze idea
- `GET /api/v1/analytics/stats` - Get statistics

## Notes

- The UI is styled with Tailwind CSS
- Make sure the Go API server is running before starting the frontend
- All pages are mobile-responsive

## Known Issues

- Skeleton UI has version conflicts with Tailwind CSS - using plain Tailwind for now
- Some styling may need refinement
