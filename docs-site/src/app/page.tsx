import Link from 'next/link';
import CodeBlock from '@/components/CodeBlock';

const QUICKSTART_CODE = `# Run the gateway locally (dev token)
export RELIA_DEV_TOKEN=dev
go run ./cmd/relia-gateway

# Smoke check
curl -sS http://localhost:8080/healthz

# Run the one-path demo (approval → receipt → pack)
open docs/DEMO.md
`;

export default function Home() {
  return (
    <div className="not-prose">
      <div className="text-center py-12 lg:py-20">
        <h1 className="text-4xl lg:text-6xl font-bold text-white mb-6">
          Policy-gated automation with{' '}
          <span className="bg-gradient-to-r from-cyan-400 to-purple-500 bg-clip-text text-transparent">
            zero standing secrets
          </span>
        </h1>
        <p className="text-xl text-gray-400 max-w-2xl mx-auto mb-8">
          Put a gate in front of production-changing automation. Require approval, mint short-lived AWS creds via GitHub
          OIDC, and emit signed receipts + audit packs.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            href="/docs/demo"
            className="px-6 py-3 bg-cyan-500 hover:bg-cyan-400 text-gray-900 font-semibold rounded-lg transition-colors"
          >
            Run the 15-minute demo
          </Link>
          <Link
            href="/docs"
            className="px-6 py-3 bg-gray-800 hover:bg-gray-700 text-gray-100 font-semibold rounded-lg border border-gray-700 transition-colors"
          >
            Read docs
          </Link>
        </div>
      </div>

      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
        <FeatureCard
          icon="🔐"
          title="Zero standing secrets"
          description="Use GitHub OIDC + AWS STS. No long-lived cloud keys in CI."
          href="/docs/aws-oidc"
        />
        <FeatureCard
          icon="✅"
          title="Human approvals"
          description="Require Slack approval for high-risk actions (optional)."
          href="/docs/slack"
        />
        <FeatureCard
          icon="🧾"
          title="Audit artifacts"
          description="Signed receipts + packs you can attach to an incident or audit ticket."
          href="/docs/demo"
        />
        <FeatureCard
          icon="🧪"
          title="Policy simulator"
          description="See which rule matches and what verdict you get before shipping."
          href="/docs/policies"
        />
        <FeatureCard
          icon="📦"
          title="One-page summary"
          description="Packs include summary.html + summary.json with checksums."
          href="/docs/demo"
        />
        <FeatureCard
          icon="🌐"
          title="Hosted verify page"
          description="Optional human-friendly verify page at /verify/<receipt_id>."
          href="/docs/quickstart"
        />
      </div>

      <div className="mb-16">
        <h2 className="text-2xl font-bold text-white mb-6 text-center">2 minutes to “it works”</h2>
        <CodeBlock code={QUICKSTART_CODE} language="bash" filename="quickstart.sh" />
      </div>

      <div className="text-center py-12 border-t border-gray-800">
        <h2 className="text-2xl font-bold text-white mb-4">Ready to gate prod changes?</h2>
        <p className="text-gray-400 mb-6">Run the demo, then drop the GitHub Action into your workflow.</p>
        <Link
          href="/docs/demo"
          className="inline-block px-6 py-3 bg-cyan-500 hover:bg-cyan-400 text-gray-900 font-semibold rounded-lg transition-colors"
        >
          Run the 15-minute demo
        </Link>
      </div>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
  href,
}: {
  icon: string;
  title: string;
  description: string;
  href: string;
}) {
  return (
    <Link
      href={href}
      className="block p-6 bg-gray-800/30 hover:bg-gray-800/50 rounded-lg border border-gray-700 hover:border-gray-600 transition-colors"
    >
      <span className="text-2xl mb-3 block">{icon}</span>
      <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
      <p className="text-sm text-gray-400">{description}</p>
    </Link>
  );
}

