-- Seed: Architecture patterns per domain
-- Run once on a fresh database: sqlite3 db/ui_patterns.db < db/seeds/002_architectures.sql

-- ─── Frontend ────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO architectures (id, name, domain, framework, css_mode, description, decisions, tags)
VALUES (
    'arch-001',
    'Astro Islands Architecture',
    'frontend', 'astro', 'tailwind-v4',
    'Static-first Astro site with selective hydration for interactive islands.',
    'Use Astro components for static content. Hydrate only interactive widgets with client:load or client:visible. Shared state via nanostores. CSS tokens in a single global.css with @theme block.',
    'astro,islands,partial-hydration,performance,tailwind-v4'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-001';

INSERT OR IGNORE INTO architectures (id, name, domain, framework, css_mode, description, decisions, tags)
VALUES (
    'arch-002',
    'Next.js App Router Architecture',
    'frontend', 'next', 'tailwind-v4',
    'Modern Next.js 14+ App Router with React Server Components and streaming.',
    'Prefer Server Components for data fetching. Use Client Components only for interactivity (event handlers, state). Co-locate styles with components. Server Actions for mutations. Tailwind v4 without tailwind.config.js.',
    'next,app-router,rsc,server-components,streaming'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-002';

-- ─── Backend ─────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO architectures (id, name, domain, framework, description, decisions, tags)
VALUES (
    'arch-003',
    'Hexagonal Architecture (Ports & Adapters)',
    'backend', NULL,
    'Domain-driven backend with clear separation between business logic and infrastructure.',
    'Core domain layer has no framework dependencies. Adapters implement ports (interfaces). Use dependency injection. Repository pattern for data access. Application services orchestrate use cases.',
    'hexagonal,clean-architecture,ddd,ports-adapters,backend'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-003';

INSERT OR IGNORE INTO architectures (id, name, domain, framework, description, decisions, tags)
VALUES (
    'arch-004',
    'Event-Driven Microservices',
    'backend', NULL,
    'Loosely coupled services communicating via events (message queue or event bus).',
    'Each service owns its data and schema. Publish domain events on state changes. Consumers react asynchronously. Use sagas for distributed transactions. Prefer eventual consistency over synchronous coupling.',
    'microservices,events,kafka,rabbitmq,saga,backend'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-004';

-- ─── Fullstack ────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO architectures (id, name, domain, framework, css_mode, description, decisions, tags)
VALUES (
    'arch-005',
    'SvelteKit Full-Stack Application',
    'fullstack', 'sveltekit', 'tailwind-v4',
    'SvelteKit monolith with server-side rendering, API routes, and form actions.',
    'Use load() functions for server data fetching. Form actions for mutations (no client fetch needed). +layout.server.ts for shared auth context. Progressive enhancement by default. Deploy to Vercel, Cloudflare, or Node adapter.',
    'sveltekit,fullstack,ssr,form-actions,progressive-enhancement'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-005';

INSERT OR IGNORE INTO architectures (id, name, domain, framework, css_mode, description, decisions, tags)
VALUES (
    'arch-006',
    'Nuxt 3 Full-Stack with Nitro',
    'fullstack', 'nuxt', 'tailwind-v4',
    'Nuxt 3 application with Nitro server engine, server routes, and composables.',
    'Use useAsyncData / useFetch for data fetching with SSR. Server routes in server/api/. Composables in composables/ for shared logic. Pinia for global state. Nuxt modules for auth (sidebase/nuxt-auth).',
    'nuxt,nitro,composables,pinia,fullstack'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-006';

-- ─── DevOps ───────────────────────────────────────────────────────────────────
INSERT OR IGNORE INTO architectures (id, name, domain, description, decisions, tags)
VALUES (
    'arch-007',
    'GitOps with GitHub Actions + Kubernetes',
    'devops',
    'Fully automated CI/CD pipeline with GitOps principles and Kubernetes deployment.',
    'Source of truth is Git. PRs trigger CI (lint, test, build, security scan). Merge to main triggers CD. Use ArgoCD or Flux for cluster sync. Helm charts for deployment configuration. Separate clusters per environment.',
    'gitops,kubernetes,github-actions,argocd,devops,ci-cd'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-007';

INSERT OR IGNORE INTO architectures (id, name, domain, description, decisions, tags)
VALUES (
    'arch-008',
    'Serverless Edge Functions',
    'devops',
    'Deploy functions to the CDN edge for global low-latency without managing servers.',
    'Use Cloudflare Workers or Vercel Edge Functions for API routes. Store state in Durable Objects or KV. Cold starts are near-zero. Limit bundle size to stay within edge constraints. Use streaming responses for AI/LLM output.',
    'serverless,edge,cloudflare,vercel,low-latency,devops'
);
INSERT OR IGNORE INTO architectures_fts (rowid, name, description, tags, decisions)
SELECT rowid, name, description, tags, decisions FROM architectures WHERE id = 'arch-008';
