import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const docsDirectory = path.join(process.cwd(), '..', 'docs');

export interface DocContent {
  content: string;
  frontmatter: {
    title?: string;
    description?: string;
    keywords?: string;
  };
  resolvedSlug: string;
}

function buildDocIndex(): Map<string, string> {
  const map = new Map<string, string>();

  function walkDir(dir: string, prefix = '') {
    const files = fs.readdirSync(dir);
    for (const file of files) {
      if (file.startsWith('.')) continue;
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);
      if (stat.isDirectory()) {
        if (file.toLowerCase() === 'blog') continue;
        walkDir(filePath, path.join(prefix, file));
        continue;
      }
      if (!file.endsWith('.md') || file.startsWith('_')) continue;

      const rel = path.join(prefix, file.replace(/\.md$/, ''));
      const normalized = rel.split(path.sep).join('/').toLowerCase();
      const actual = rel.split(path.sep).join('/');
      map.set(normalized, actual);

      // Also index a hyphenated alias (e.g. AWS_OIDC -> aws-oidc) for nicer URLs.
      const hyphen = normalized.replace(/_/g, '-');
      map.set(hyphen, actual);

      // And the reverse alias for files that use hyphens.
      const underscore = normalized.replace(/-/g, '_');
      map.set(underscore, actual);
    }
  }

  if (fs.existsSync(docsDirectory)) {
    walkDir(docsDirectory);
  }
  return map;
}

function listActualDocSlugs(): string[] {
  const results: string[] = [];

  function walkDir(dir: string, prefix = '') {
    const files = fs.readdirSync(dir);
    for (const file of files) {
      if (file.startsWith('.')) continue;
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);
      if (stat.isDirectory()) {
        if (file.toLowerCase() === 'blog') continue;
        walkDir(filePath, path.join(prefix, file));
        continue;
      }
      if (!file.endsWith('.md') || file.startsWith('_')) continue;
      const rel = path.join(prefix, file.replace(/\.md$/, ''));
      results.push(rel.split(path.sep).join('/'));
    }
  }

  if (fs.existsSync(docsDirectory)) {
    walkDir(docsDirectory);
  }

  return results;
}

function resolveSlug(slug: string): { normalized: string; actual: string | null } {
  const cleaned = (slug || '').replace(/^\/+|\/+$/g, '');
  const normalized = cleaned.toLowerCase();

  const index = buildDocIndex();
  const actual = index.get(normalized) ?? null;

  // Fallback: also try "index" within a folder
  if (!actual) {
    const alt = normalized.endsWith('/index') ? normalized : `${normalized}/index`;
    return { normalized, actual: index.get(alt) ?? null };
  }

  return { normalized, actual };
}

export function getDocContent(slug: string): DocContent | null {
  const { actual, normalized } = resolveSlug(slug);
  if (!actual) return null;

  const filePath = path.join(docsDirectory, `${actual}.md`);
  if (!fs.existsSync(filePath)) return null;

  const fileContents = fs.readFileSync(filePath, 'utf8');
  const { data, content } = matter(fileContents);

  const cleanContent = content.replace(/\{%.*?%\}/g, '').replace(/\{\{.*?\}\}/g, '');

  return {
    content: cleanContent,
    frontmatter: data as DocContent['frontmatter'],
    resolvedSlug: normalized,
  };
}

export function getAllDocSlugs(): string[] {
  const actuals = listActualDocSlugs();
  const canonical = actuals.map((s) => s.toLowerCase().replace(/_/g, '-'));
  return Array.from(new Set(canonical)).sort();
}
