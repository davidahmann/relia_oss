import { getDocContent } from '@/lib/docs';
import { markdownToHtml } from '@/lib/markdown';
import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import MarkdownRenderer from '@/components/MarkdownRenderer';

export const metadata: Metadata = {
  title: 'Documentation - Relia',
  description: 'Relia documentation',
};

export default function DocsIndexPage() {
  const doc = getDocContent('index');
  if (!doc) notFound();
  const html = markdownToHtml(doc.content);
  return <MarkdownRenderer html={html} />;
}

