-- Seed: UI patterns for each framework/domain combination
-- Run once on a fresh database: sqlite3 db/ui_patterns.db < db/seeds/001_patterns.sql

-- ─── Frontend / Astro / Tailwind v4 ────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-001',
    'Hero Section — Tailwind v4 Centered',
    'frontend', 'landing', 'astro', 'tailwind-v4',
    'hero,landing,centered,tailwind',
    '<section class="flex flex-col items-center py-24 gap-6">
  <h1 class="text-5xl font-bold tracking-tight text-(--color-text)">Build faster, ship smarter</h1>
  <p class="text-xl text-(--color-muted) max-w-2xl text-center">Your AI-powered development toolkit.</p>
  <div class="flex gap-4">
    <a href="#" class="bg-(--color-primary) text-white px-8 py-3 rounded-lg font-semibold">Get Started</a>
    <a href="#" class="border border-(--color-primary) text-(--color-primary) px-8 py-3 rounded-lg font-semibold">Learn More</a>
  </div>
</section>',
    'Full-width hero with centered headline, subtext, and dual CTA buttons. Tailwind v4 tokens.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-001';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-002',
    'Feature Grid — Tailwind v4 Three-Column',
    'frontend', 'landing', 'astro', 'tailwind-v4',
    'features,grid,three-column,tailwind',
    '<section class="py-20 px-4">
  <div class="max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
    <div class="p-8 bg-(--color-surface) rounded-xl">
      <div class="text-3xl mb-4">⚡</div>
      <h3 class="text-xl font-semibold mb-2">Fast</h3>
      <p class="text-(--color-muted)">Optimized for performance from day one.</p>
    </div>
  </div>
</section>',
    'Three-column feature grid with icon, title, and description. Tailwind v4 surface tokens.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-002';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-003',
    'Nav Header — Tailwind v4 Sticky',
    'frontend', 'component', 'astro', 'tailwind-v4',
    'nav,header,sticky,tailwind',
    '<header class="sticky top-0 z-50 bg-(--color-background)/80 backdrop-blur border-b border-(--color-surface)">
  <nav class="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
    <a href="/" class="font-bold text-xl">Brand</a>
    <ul class="hidden md:flex gap-8 text-(--color-muted)">
      <li><a href="#features" class="hover:text-(--color-primary) transition">Features</a></li>
      <li><a href="#pricing" class="hover:text-(--color-primary) transition">Pricing</a></li>
    </ul>
    <a href="#" class="bg-(--color-primary) text-white px-4 py-2 rounded-lg text-sm font-medium">Sign up</a>
  </nav>
</header>',
    'Sticky navigation header with blur backdrop. Tailwind v4 CSS variable tokens.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-003';

-- ─── Frontend / Astro / Plain CSS ────────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, css_snippet, description)
VALUES (
    'pat-004',
    'Hero Section — Plain CSS Centered',
    'frontend', 'landing', 'astro', 'plain-css',
    'hero,landing,plain-css,centered',
    '<section class="hero-section">
  <div class="hero-container">
    <h1 class="hero-heading">Build faster, ship smarter</h1>
    <p class="hero-sub">Your development toolkit.</p>
    <a href="#" class="hero-cta">Get Started</a>
  </div>
</section>',
    '.hero-section { padding: 6rem 1rem; display: flex; justify-content: center; }
.hero-container { max-width: 48rem; text-align: center; }
.hero-heading { font-size: 3rem; font-weight: 700; color: var(--color-text); }
.hero-sub { font-size: 1.25rem; color: var(--color-muted); margin-top: 1rem; }
.hero-cta { display: inline-block; background: var(--color-primary); color: white; padding: 0.75rem 2rem; border-radius: 0.5rem; text-decoration: none; margin-top: 2rem; }',
    'Centered hero section using CSS custom properties. No framework dependency.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-004';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-005',
    'Card Grid — Plain CSS Responsive',
    'frontend', 'component', 'astro', 'plain-css',
    'card,grid,responsive,plain-css',
    '<div class="card-grid">
  <article class="card">
    <img src="/img/feature.png" alt="Feature" class="card-image">
    <div class="card-body">
      <h3 class="card-title">Card Title</h3>
      <p class="card-desc">Brief description here.</p>
    </div>
  </article>
</div>',
    'Responsive CSS grid of cards with image, title, and description. Uses CSS custom properties.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-005';

-- ─── Frontend / Next.js / Tailwind v4 ────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-006',
    'Dashboard Stats — Next.js Tailwind v4',
    'frontend', 'dashboard', 'next', 'tailwind-v4',
    'dashboard,stats,kpi,next,tailwind',
    'export function StatsGrid() {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 p-6">
      {[
        { label: "Total Users", value: "12,450", change: "+8%" },
        { label: "Revenue", value: "$48,290", change: "+12%" },
      ].map((stat) => (
        <div key={stat.label} className="bg-(--color-surface) p-6 rounded-xl">
          <p className="text-(--color-muted) text-sm">{stat.label}</p>
          <p className="text-2xl font-bold mt-1">{stat.value}</p>
          <span className="text-green-500 text-sm">{stat.change}</span>
        </div>
      ))}
    </div>
  );
}',
    'KPI stats grid for dashboards. Four columns on desktop, two on mobile. Tailwind v4 surface tokens.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-006';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-007',
    'Login Form — Next.js Tailwind v4',
    'frontend', 'form', 'next', 'tailwind-v4',
    'form,login,auth,next,tailwind',
    'export function LoginForm() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-(--color-background)">
      <form className="w-full max-w-sm bg-(--color-surface) p-8 rounded-2xl shadow-lg">
        <h2 className="text-2xl font-bold mb-6">Sign in</h2>
        <div className="space-y-4">
          <input type="email" placeholder="Email" className="w-full px-4 py-3 rounded-lg border border-(--color-surface) bg-(--color-background)" />
          <input type="password" placeholder="Password" className="w-full px-4 py-3 rounded-lg border border-(--color-surface) bg-(--color-background)" />
          <button className="w-full bg-(--color-primary) text-white py-3 rounded-lg font-semibold">Sign in</button>
        </div>
      </form>
    </div>
  );
}',
    'Centered login form with email/password fields. Next.js + Tailwind v4.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-007';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-008',
    'Sidebar Nav — Next.js Plain CSS',
    'frontend', 'dashboard', 'next', 'plain-css',
    'sidebar,nav,dashboard,next,plain-css',
    'export function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="sidebar-brand">DevForge</div>
      <nav className="sidebar-nav">
        <a href="/dashboard" className="sidebar-link active">Dashboard</a>
        <a href="/projects" className="sidebar-link">Projects</a>
        <a href="/settings" className="sidebar-link">Settings</a>
      </nav>
    </aside>
  );
}',
    'Left sidebar navigation for dashboards. Next.js with CSS custom properties.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-008';

-- ─── Frontend / SvelteKit ───────────────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-009',
    'Hero — SvelteKit Tailwind v4',
    'frontend', 'landing', 'sveltekit', 'tailwind-v4',
    'hero,landing,sveltekit,tailwind',
    '<script lang="ts">
  let { heading = "Build with SvelteKit" } = $props();
</script>

<section class="py-24 px-4 text-center">
  <h1 class="text-5xl font-bold">{heading}</h1>
  <p class="mt-6 text-xl text-(--color-muted)">Fast, elegant, and reactive.</p>
  <a href="/start" class="mt-8 inline-block bg-(--color-primary) text-white px-8 py-3 rounded-lg">Get Started</a>
</section>',
    'Hero section for SvelteKit landing pages. Uses Svelte 5 $props() rune and Tailwind v4.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-009';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-010',
    'Dashboard Layout — SvelteKit Plain CSS',
    'frontend', 'dashboard', 'sveltekit', 'plain-css',
    'dashboard,layout,sveltekit,plain-css',
    '<div class="app-layout">
  <aside class="sidebar">
    <slot name="sidebar" />
  </aside>
  <main class="main-content">
    <slot />
  </main>
</div>

<style>
  .app-layout { display: grid; grid-template-columns: 16rem 1fr; min-height: 100vh; }
  .sidebar { background: var(--color-surface); border-right: 1px solid var(--color-muted); padding: 1.5rem; }
  .main-content { padding: 2rem; overflow-y: auto; }
</style>',
    'SvelteKit dashboard layout with sidebar slot pattern. Plain CSS custom properties.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-010';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-011',
    'Data Table — SvelteKit Tailwind v4',
    'frontend', 'dashboard', 'sveltekit', 'tailwind-v4',
    'table,data,sveltekit,tailwind,dashboard',
    '<script lang="ts">
  let { rows = [] } = $props();
</script>

<div class="overflow-x-auto rounded-xl border border-(--color-surface)">
  <table class="w-full text-sm">
    <thead class="bg-(--color-surface) text-(--color-muted)">
      <tr>
        <th class="px-4 py-3 text-left">Name</th>
        <th class="px-4 py-3 text-left">Status</th>
        <th class="px-4 py-3 text-right">Actions</th>
      </tr>
    </thead>
    <tbody>
      {#each rows as row}
        <tr class="border-t border-(--color-surface) hover:bg-(--color-surface)/50">
          <td class="px-4 py-3">{row.name}</td>
          <td class="px-4 py-3">{row.status}</td>
          <td class="px-4 py-3 text-right">...</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>',
    'SvelteKit data table component with hover states and Tailwind v4 tokens.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-011';

-- ─── Backend patterns ────────────────────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-012',
    'REST API Endpoint Structure',
    'backend', 'component', 'vanilla', 'plain-css',
    'api,rest,endpoint,backend,structure',
    '// routes/users.ts — Express-style REST endpoint pattern
export const userRoutes = Router();

userRoutes.get("/", authenticate, async (req, res) => {
  try {
    const users = await UserService.list(req.query);
    res.json({ data: users, total: users.length });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

userRoutes.post("/", authenticate, validate(createUserSchema), async (req, res) => {
  const user = await UserService.create(req.body);
  res.status(201).json({ data: user });
});',
    'REST API endpoint structure with authentication middleware and error handling.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-012';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-013',
    'Middleware Chain Pattern',
    'backend', 'component', 'vanilla', 'plain-css',
    'middleware,chain,backend,auth,logging',
    '// middleware/index.ts — Composable middleware chain
export const chain = (...middlewares) =>
  (req, res, next) => {
    const run = (index) => {
      if (index === middlewares.length) return next();
      middlewares[index](req, res, () => run(index + 1));
    };
    run(0);
  };

export const apiMiddleware = chain(
  requestLogger,
  rateLimit({ max: 100, window: "1m" }),
  authenticate,
  parseJSON,
);',
    'Composable middleware chain for Node.js/Express backends.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-013';

-- ─── Fullstack patterns ───────────────────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-014',
    'Auth Flow — Next.js App Router',
    'fullstack', 'component', 'next', 'tailwind-v4',
    'auth,login,session,next,fullstack',
    '// app/api/auth/route.ts — Next.js App Router auth endpoint
export async function POST(request: Request) {
  const { email, password } = await request.json();
  const user = await db.user.findUnique({ where: { email } });
  if (!user || !await bcrypt.compare(password, user.passwordHash)) {
    return Response.json({ error: "Invalid credentials" }, { status: 401 });
  }
  const token = await createSession(user.id);
  return Response.json({ token }, {
    headers: { "Set-Cookie": `session=${token}; HttpOnly; SameSite=Strict` },
  });
}',
    'Full-stack auth flow with Next.js App Router, bcrypt password verification, and session cookie.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-014';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-015',
    'Data Fetching — SvelteKit +page.server.ts',
    'fullstack', 'component', 'sveltekit', 'plain-css',
    'data-fetching,load,sveltekit,fullstack,server',
    '// +page.server.ts — SvelteKit server-side data loading
export const load = async ({ params, locals }) => {
  const { db } = locals;
  const [project, tasks] = await Promise.all([
    db.project.findUnique({ where: { id: params.id } }),
    db.task.findMany({ where: { projectId: params.id }, orderBy: { createdAt: "desc" } }),
  ]);
  if (!project) throw error(404, "Project not found");
  return { project, tasks };
};',
    'SvelteKit server load function with parallel data fetching and 404 error handling.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-015';

-- ─── Dashboard patterns (frontend) ───────────────────────────────────────────
INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-016',
    'Analytics Dashboard — Next.js Tailwind v4',
    'frontend', 'dashboard', 'next', 'tailwind-v4',
    'analytics,dashboard,charts,next,tailwind',
    'export function AnalyticsDashboard({ metrics }) {
  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Analytics</h1>
        <select className="bg-(--color-surface) rounded-lg px-3 py-2 text-sm">
          <option>Last 30 days</option>
          <option>Last 7 days</option>
        </select>
      </div>
      <div className="grid grid-cols-3 gap-4">
        <MetricCard label="Page Views" value={metrics.views} />
        <MetricCard label="Unique Visitors" value={metrics.visitors} />
        <MetricCard label="Bounce Rate" value={metrics.bounceRate} />
      </div>
    </div>
  );
}',
    'Analytics dashboard shell with time range selector and metric cards. Next.js + Tailwind v4.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-016';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-017',
    'Settings Page — Nuxt Tailwind v4',
    'frontend', 'dashboard', 'nuxt', 'tailwind-v4',
    'settings,profile,nuxt,tailwind,dashboard',
    '<template>
  <div class="max-w-2xl mx-auto py-8 px-4 space-y-6">
    <h1 class="text-2xl font-bold">Settings</h1>
    <section class="bg-(--color-surface) rounded-xl p-6 space-y-4">
      <h2 class="font-semibold">Profile</h2>
      <div class="grid grid-cols-2 gap-4">
        <UiInput v-model="form.firstName" label="First name" />
        <UiInput v-model="form.lastName" label="Last name" />
      </div>
      <UiInput v-model="form.email" label="Email" type="email" />
      <UiButton @click="save" variant="primary">Save changes</UiButton>
    </section>
  </div>
</template>',
    'Settings page layout with profile form. Nuxt + Tailwind v4 surface card style.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-017';

INSERT OR IGNORE INTO patterns (id, name, domain, category, framework, css_mode, tags, snippet, description)
VALUES (
    'pat-018',
    'Hero — Nuxt Plain CSS',
    'frontend', 'landing', 'nuxt', 'plain-css',
    'hero,landing,nuxt,plain-css',
    '<template>
  <section class="hero">
    <div class="hero-inner">
      <h1>{{ title }}</h1>
      <p>{{ subtitle }}</p>
      <NuxtLink to="/start" class="btn-primary">{{ cta }}</NuxtLink>
    </div>
  </section>
</template>

<script setup>
defineProps(["title", "subtitle", "cta"]);
</script>

<style scoped>
.hero { display: flex; align-items: center; justify-content: center; padding: 6rem 1rem; }
.hero-inner { text-align: center; max-width: 40rem; }
</style>',
    'Nuxt hero component with scoped CSS custom properties and slot props.'
);
INSERT OR IGNORE INTO patterns_fts (rowid, name, category, tags, description)
SELECT rowid, name, category, tags, description FROM patterns WHERE id = 'pat-018';
