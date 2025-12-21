export interface NavItem {
  title: string;
  href: string;
  children?: NavItem[];
}

export const navigation: NavItem[] = [
  {
    title: 'Getting Started',
    href: '/docs',
    children: [
      { title: 'Introduction', href: '/docs' },
      { title: 'Demo (15 min)', href: '/docs/demo' },
      { title: 'Quickstart', href: '/docs/quickstart' },
      { title: 'Policies', href: '/docs/policies' },
    ],
  },
  {
    title: 'Integrations',
    href: '/docs/aws-oidc',
    children: [
      { title: 'AWS (GitHub OIDC)', href: '/docs/aws-oidc' },
      { title: 'Slack approvals', href: '/docs/slack' },
    ],
  },
  {
    title: 'Reference',
    href: '/docs/testing',
    children: [
      { title: 'Testing', href: '/docs/testing' },
      { title: 'Security', href: '/docs/security' },
      { title: 'Release', href: '/docs/release' },
      { title: 'Roadmap', href: '/docs/roadmap' },
    ],
  },
  {
    title: 'Blog',
    href: '/blog',
    children: [
      { title: 'All Posts', href: '/blog' },
      { title: 'Why receipts beat logs', href: '/blog/receipts-beat-logs' },
      { title: 'Zero secrets with GitHub OIDC', href: '/blog/zero-secrets-github-oidc-aws' },
      { title: 'Slack approvals that don’t flake', href: '/blog/slack-approvals-with-retries' },
      { title: 'Audit packs for incident response', href: '/blog/audit-packs-incident-response' },
      { title: 'A simple “quality grade” for receipts', href: '/blog/receipt-quality-grade' },
      { title: 'Policy simulator: instant clarity', href: '/blog/policy-simulator-instant-clarity' },
    ],
  },
];
