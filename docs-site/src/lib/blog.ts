import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const blogDirectory = path.join(process.cwd(), '..', 'docs', 'blog');

export interface BlogPost {
  content: string;
  frontmatter: {
    title?: string;
    description?: string;
    keywords?: string;
    date?: string;
  };
  slug: string;
}

function listMarkdownFiles(dir: string, prefix = ''): string[] {
  if (!fs.existsSync(dir)) return [];
  const entries = fs.readdirSync(dir);
  const results: string[] = [];

  for (const entry of entries) {
    if (entry.startsWith('.')) continue;
    const entryPath = path.join(dir, entry);
    const stat = fs.statSync(entryPath);
    if (stat.isDirectory()) {
      results.push(...listMarkdownFiles(entryPath, path.join(prefix, entry)));
      continue;
    }
    if (!entry.endsWith('.md') || entry.startsWith('_')) continue;
    const rel = path.join(prefix, entry.replace(/\.md$/, ''));
    results.push(rel.split(path.sep).join('/'));
  }

  return results;
}

export function getAllBlogSlugs(): string[] {
  return listMarkdownFiles(blogDirectory)
    .map((s) => s.toLowerCase())
    .sort();
}

export function getBlogPost(slug: string): BlogPost | null {
  const cleaned = (slug || '').replace(/^\/+|\/+$/g, '');
  const normalized = cleaned.toLowerCase();

  // Case-insensitive resolution so slugs can be lowercase even if filenames aren't.
  const index = new Map<string, string>();
  for (const rel of listMarkdownFiles(blogDirectory)) {
    index.set(rel.toLowerCase(), rel);
  }

  const actual = index.get(normalized);
  if (!actual) return null;

  const filePath = path.join(blogDirectory, `${actual}.md`);
  if (!fs.existsSync(filePath)) return null;

  const fileContents = fs.readFileSync(filePath, 'utf8');
  const { data, content } = matter(fileContents);

  const cleanContent = content.replace(/\{%.*?%\}/g, '').replace(/\{\{.*?\}\}/g, '');

  return {
    content: cleanContent,
    frontmatter: data as BlogPost['frontmatter'],
    slug: normalized,
  };
}

export function getBlogIndex(): Array<Pick<BlogPost, 'slug' | 'frontmatter'>> {
  const slugs = getAllBlogSlugs();
  return slugs
    .map((slug) => {
      const post = getBlogPost(slug);
      if (!post) return null;
      return { slug: post.slug, frontmatter: post.frontmatter };
    })
    .filter(Boolean) as Array<Pick<BlogPost, 'slug' | 'frontmatter'>>;
}

