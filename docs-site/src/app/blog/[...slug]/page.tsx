import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import MarkdownRenderer from '@/components/MarkdownRenderer';
import { markdownToHtml } from '@/lib/markdown';
import { getAllBlogSlugs, getBlogPost } from '@/lib/blog';

interface PageProps {
  params: Promise<{ slug: string[] }>;
}

export async function generateStaticParams() {
  const slugs = getAllBlogSlugs();
  return slugs.map((slug) => ({ slug: slug.split('/') }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const slugPath = slug.join('/');
  const post = getBlogPost(slugPath);

  const title = post?.frontmatter?.title || slugPath.split('/').pop()?.replace(/-|_/g, ' ') || 'Blog';
  const description = post?.frontmatter?.description || 'Relia blog post';

  return {
    title: `${title} - Relia Blog`,
    description,
  };
}

export default async function BlogPostPage({ params }: PageProps) {
  const { slug } = await params;
  const slugPath = slug.join('/');
  const post = getBlogPost(slugPath);
  if (!post) notFound();

  const html = markdownToHtml(post.content);
  return <MarkdownRenderer html={html} />;
}

