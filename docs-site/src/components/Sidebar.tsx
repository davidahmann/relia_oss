'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { navigation, NavItem } from '@/lib/navigation';

function NavLink({ item, depth = 0 }: { item: NavItem; depth?: number }) {
  const pathname = usePathname();
  const isActive = pathname === item.href || pathname === item.href + '/';

  return (
    <Link
      href={item.href}
      className={`
        block py-1.5 px-3 text-sm rounded-md transition-colors
        ${depth > 0 ? 'ml-4' : ''}
        ${
          isActive
            ? 'bg-cyan-500/10 text-cyan-400 border-l-2 border-cyan-400'
            : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/50 border-l-2 border-transparent'
        }
      `}
    >
      {item.title}
    </Link>
  );
}

function NavSection({ item }: { item: NavItem }) {
  const pathname = usePathname();
  const isInSection = item.children?.some((child) => pathname === child.href || pathname === child.href + '/');

  return (
    <div className="mb-6">
      <h3
        className={`
        px-3 mb-2 text-xs font-semibold uppercase tracking-wider
        ${isInSection ? 'text-cyan-400' : 'text-gray-500'}
      `}
      >
        {item.title}
      </h3>
      <div className="space-y-1">
        {item.children?.map((child) => (
          <NavLink key={child.href} item={child} depth={1} />
        ))}
      </div>
    </div>
  );
}

export default function Sidebar() {
  return (
    <aside className="w-64 flex-shrink-0 hidden lg:block">
      <div className="sticky top-0 h-screen overflow-y-auto py-8 pr-4">
        <Link href="/" className="flex items-center gap-2 px-3 mb-6">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-cyan-500 to-purple-600 flex items-center justify-center">
            <span className="text-white font-bold text-sm">R</span>
          </div>
          <span className="text-xl font-bold text-white">Relia</span>
        </Link>

        <nav>
          {navigation.map((section) => (
            <NavSection key={section.title} item={section} />
          ))}
        </nav>

        <div className="mt-8 pt-8 border-t border-gray-800 px-3">
          <a
            href="https://github.com/davidahmann/relia"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-sm text-gray-400 hover:text-gray-200"
          >
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path
                fillRule="evenodd"
                d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                clipRule="evenodd"
              />
            </svg>
            GitHub
          </a>
        </div>
      </div>
    </aside>
  );
}

