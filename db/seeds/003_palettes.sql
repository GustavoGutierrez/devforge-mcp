-- Seed: Color palettes for common use cases
-- Run once on a fresh database: sqlite3 db/ui_patterns.db < db/seeds/003_palettes.sql

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-001', 'Fintech Calm Blue', 'saas-dashboard', 'serious',
    '{"background":"#0b1220","surface":"#020617","primary":"#22d3ee","primary-soft":"#0f172a","accent":"#facc15","text":"#e2e8f0","muted":"#475569"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-002', 'SaaS Indigo Light', 'saas-dashboard', 'professional',
    '{"background":"#ffffff","surface":"#f8fafc","primary":"#6366f1","primary-soft":"#eef2ff","accent":"#f59e0b","text":"#1e293b","muted":"#64748b"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-003', 'Marketing Vibrant Red', 'marketing-site', 'bold',
    '{"background":"#fafafa","surface":"#f4f4f5","primary":"#ef4444","primary-soft":"#fef2f2","accent":"#8b5cf6","text":"#18181b","muted":"#71717a"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-004', 'Editorial Minimal', 'blog', 'minimal',
    '{"background":"#ffffff","surface":"#f9fafb","primary":"#111827","primary-soft":"#f3f4f6","accent":"#3b82f6","text":"#111827","muted":"#9ca3af"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-005', 'Developer Dark', 'developer-tool', 'focused',
    '{"background":"#1a1b26","surface":"#24283b","primary":"#7aa2f7","primary-soft":"#2d3f6c","accent":"#9ece6a","text":"#c0caf5","muted":"#565f89"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-006', 'E-commerce Warm Amber', 'ecommerce', 'warm',
    '{"background":"#fffbf5","surface":"#fef3c7","primary":"#d97706","primary-soft":"#fef9ee","accent":"#dc2626","text":"#292524","muted":"#78716c"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-007', 'Health & Wellness Green', 'health-app', 'calm',
    '{"background":"#f0fdf4","surface":"#dcfce7","primary":"#16a34a","primary-soft":"#f0fdf4","accent":"#0891b2","text":"#14532d","muted":"#6b7280"}'
);

INSERT OR IGNORE INTO palettes (id, name, use_case, mood, tokens_json) VALUES (
    'pal-008', 'Premium Dark Violet', 'premium-product', 'premium',
    '{"background":"#09090b","surface":"#18181b","primary":"#a78bfa","primary-soft":"#1c1c27","accent":"#34d399","text":"#fafafa","muted":"#71717a"}'
);
