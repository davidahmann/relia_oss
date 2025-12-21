import type { Metadata } from 'next';
import './globals.css';
import Sidebar from '@/components/Sidebar';
import Header from '@/components/Header';

export const metadata: Metadata = {
  title: 'Relia - Policy-gated automation',
  description: 'Policy-gated automation with zero standing secrets. Signed receipts you can hand to security/audit.',
  openGraph: {
    title: 'Relia - Policy-gated automation',
    description: 'Policy-gated automation with zero standing secrets. Signed receipts you can hand to security/audit.',
    url: 'https://davidahmann.github.io/relia',
    siteName: 'Relia',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Relia - Policy-gated automation',
    description: 'Policy-gated automation with zero standing secrets. Signed receipts you can hand to security/audit.',
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <body className="antialiased">
        <Header />
        <div className="flex max-w-7xl mx-auto px-4 lg:px-8">
          <Sidebar />
          <main className="flex-1 min-w-0 py-8 lg:pl-8">
            <article className="prose prose-invert max-w-none">{children}</article>
          </main>
        </div>
      </body>
    </html>
  );
}

