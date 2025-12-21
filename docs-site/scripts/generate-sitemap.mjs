import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const SITE_URL = 'https://davidahmann.github.io/relia';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const siteDir = path.resolve(scriptDir, '..');
const docsDir = path.resolve(siteDir, '..', 'docs');
const blogDir = path.join(docsDir, 'blog');
const publicDir = path.join(siteDir, 'public');

function listMarkdown(dir, options = {}) {
  const { skipDirs = new Set(), prefix = '' } = options;
  if (!fs.existsSync(dir)) return [];
  const entries = fs.readdirSync(dir);
  const results = [];

  for (const entry of entries) {
    if (entry.startsWith('.')) continue;
    const entryPath = path.join(dir, entry);
    const stat = fs.statSync(entryPath);
    if (stat.isDirectory()) {
      if (skipDirs.has(entry.toLowerCase())) continue;
      results.push(...listMarkdown(entryPath, { ...options, prefix: path.join(prefix, entry) }));
      continue;
    }
    if (!entry.endsWith('.md') || entry.startsWith('_')) continue;
    const rel = path.join(prefix, entry.replace(/\.md$/, ''));
    results.push(rel.split(path.sep).join('/'));
  }

  return results;
}

function canonicalizeSlug(slug) {
  return slug.toLowerCase().replace(/_/g, '-');
}

function toUrl(pathname) {
  const p = pathname.endsWith('/') ? pathname : `${pathname}/`;
  return `${SITE_URL}${p}`;
}

const urls = new Set();
urls.add(toUrl('/'));
urls.add(toUrl('/docs'));
urls.add(toUrl('/blog'));

for (const doc of listMarkdown(docsDir, { skipDirs: new Set(['blog']) })) {
  const slug = canonicalizeSlug(doc);
  if (slug === 'index') continue;
  urls.add(toUrl(`/docs/${slug}`));
}

for (const post of listMarkdown(blogDir)) {
  const slug = canonicalizeSlug(post);
  urls.add(toUrl(`/blog/${slug}`));
}

const lastmod = new Date().toISOString().slice(0, 10);

const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${Array.from(urls)
  .sort()
  .map((loc) => `  <url><loc>${loc}</loc><lastmod>${lastmod}</lastmod></url>`)
  .join('\n')}
</urlset>
`;

fs.mkdirSync(publicDir, { recursive: true });
fs.writeFileSync(path.join(publicDir, 'sitemap.xml'), xml, 'utf8');
console.log(`wrote public/sitemap.xml with ${urls.size} urls`);
