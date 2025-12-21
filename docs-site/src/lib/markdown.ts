import { marked } from 'marked';

marked.setOptions({
  gfm: true,
  breaks: false,
});

const basePath = process.env.NODE_ENV === 'production' ? '/relia' : '';

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim();
}

const renderer = new marked.Renderer();

renderer.link = function ({ href, title, text }) {
  if (href && href.endsWith('.md')) {
    let p = href.replace(/\.md$/, '');
    p = p.replace(/^\.\//, '');
    p = p.replace(/^\.\.\//, '');
    p = p.replace(/^docs\//i, '');
    p = p.replace(/^blog\//i, 'blog/');

    const normalized = p.split('#')[0].toLowerCase();
    const anchor = p.includes('#') ? '#' + p.split('#').slice(1).join('#') : '';
    if (normalized.startsWith('blog/')) {
      href = basePath + '/blog/' + normalized.replace(/^blog\//, '') + anchor;
    } else {
      href = basePath + '/docs/' + normalized.replace(/_/g, '-') + anchor;
    }
  }

  if (href && (href.startsWith('http://') || href.startsWith('https://'))) {
    return `<a href="${href}" target="_blank" rel="noopener noreferrer"${title ? ` title="${title}"` : ''}>${text}</a>`;
  }

  return `<a href="${href}"${title ? ` title="${title}"` : ''}>${text}</a>`;
};

renderer.code = function ({ text, lang }) {
  const language = lang || '';
  if (language === 'mermaid') {
    return `<pre><code class="language-mermaid">${text}</code></pre>`;
  }
  const escapedCode = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  return `<pre><code class="language-${language}">${escapedCode}</code></pre>`;
};

renderer.codespan = function ({ text }) {
  return `<code class="inline-code">${text}</code>`;
};

renderer.heading = function ({ text, depth }) {
  const slug = slugify(text);
  return `<h${depth} id="${slug}">${text}</h${depth}>`;
};

marked.use({ renderer });

export function markdownToHtml(markdown: string): string {
  const cleanContent = markdown.replace(/\{%.*?%\}/g, '').replace(/\{\{.*?\}\}/g, '');
  return marked.parse(cleanContent) as string;
}
