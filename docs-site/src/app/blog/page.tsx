import Link from 'next/link';
import type { Metadata } from 'next';
import { getBlogIndex } from '@/lib/blog';

export const metadata: Metadata = {
  title: 'Blog - Relia',
  description: 'Relia blog: receipts, approvals, OIDC, and auditability for production automation.',
};

export default function BlogIndexPage() {
  const posts = getBlogIndex();

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold text-white mb-2">Blog</h1>
      <p className="text-gray-400 mb-8">
        Short, practical notes on policy-gated automation, zero-secrets CI, approvals, and audit artifacts.
      </p>

      <div className="space-y-6">
        {posts.map((post) => (
          <article key={post.slug} className="p-6 rounded-xl border border-gray-800 bg-gray-900/40">
            <h2 className="text-xl font-semibold text-white mb-1">
              <Link className="hover:text-cyan-400" href={`/blog/${post.slug}`}>
                {post.frontmatter.title ?? post.slug}
              </Link>
            </h2>
            {post.frontmatter.date && <div className="text-sm text-gray-500 mb-3">{post.frontmatter.date}</div>}
            <p className="text-gray-300">{post.frontmatter.description ?? ''}</p>
          </article>
        ))}
      </div>
    </div>
  );
}

