import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        gray: {
          50: '#f9fafb',
          100: '#f3f4f6',
          200: '#e5e7eb',
          300: '#d1d5db',
          400: '#9ca3af',
          500: '#6b7280',
          600: '#4b5563',
          700: '#374151',
          800: '#1f2937',
          900: '#111827',
          950: '#030712',
        },
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
        mono: ['Fira Code', 'Monaco', 'Consolas', 'monospace'],
      },
      typography: {
        DEFAULT: {
          css: {
            maxWidth: 'none',
            color: '#d1d5db',
            a: {
              color: '#22d3ee',
              '&:hover': {
                color: '#67e8f9',
              },
            },
            h1: {
              color: '#f9fafb',
            },
            h2: {
              color: '#f3f4f6',
            },
            h3: {
              color: '#e5e7eb',
            },
            h4: {
              color: '#e5e7eb',
            },
            code: {
              color: '#a5b4fc',
              backgroundColor: '#1f2937',
              padding: '0.25rem 0.375rem',
              borderRadius: '0.25rem',
              fontWeight: '400',
            },
            'code::before': {
              content: '""',
            },
            'code::after': {
              content: '""',
            },
            pre: {
              backgroundColor: '#1f2937',
              border: '1px solid #374151',
            },
            blockquote: {
              borderLeftColor: '#22d3ee',
              color: '#9ca3af',
            },
            hr: {
              borderColor: '#374151',
            },
            strong: {
              color: '#f9fafb',
            },
            thead: {
              color: '#f9fafb',
              borderBottomColor: '#374151',
            },
            'tbody tr': {
              borderBottomColor: '#374151',
            },
            th: {
              color: '#f9fafb',
            },
            td: {
              color: '#d1d5db',
            },
          },
        },
      },
    },
  },
  plugins: [require('@tailwindcss/typography')],
};

export default config;

