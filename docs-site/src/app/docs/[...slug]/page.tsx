import { getAllDocSlugs, getDocContent } from '@/lib/docs';
import { markdownToHtml } from '@/lib/markdown';
import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import MarkdownRenderer from '@/components/MarkdownRenderer';

interface PageProps {
  params: Promise<{ slug: string[] }>;
}

export async function generateStaticParams() {
  const slugs = getAllDocSlugs();
  return slugs
    .filter((slug) => slug !== 'index')
    .map((slug) => ({ slug: slug.split('/') }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const slugPath = slug.join('/');
  const doc = getDocContent(slugPath);
  const title = doc?.frontmatter?.title || slugPath.split('/').pop()?.replace(/-|_/g, ' ') || 'Documentation';
  return {
    title: `${title} - Relia`,
    description: doc?.frontmatter?.description || 'Relia documentation',
  };
}

export default async function DocPage({ params }: PageProps) {
  const { slug } = await params;
  const slugPath = slug.join('/');
  const doc = getDocContent(slugPath);
  if (!doc) notFound();
  const html = markdownToHtml(doc.content);
  return <MarkdownRenderer html={html} />;
}

