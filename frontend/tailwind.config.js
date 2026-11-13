const withAlpha = (variable) => `rgb(var(${variable}) / <alpha-value>)`

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调（默认 Teal 青色系，可通过皮肤 CSS 变量覆盖）
        primary: {
          50: withAlpha('--color-primary-50'),
          100: withAlpha('--color-primary-100'),
          200: withAlpha('--color-primary-200'),
          300: withAlpha('--color-primary-300'),
          400: withAlpha('--color-primary-400'),
          500: withAlpha('--color-primary-500'),
          600: withAlpha('--color-primary-600'),
          700: withAlpha('--color-primary-700'),
          800: withAlpha('--color-primary-800'),
          900: withAlpha('--color-primary-900'),
          950: withAlpha('--color-primary-950')
        },
        // 辅助色
        accent: {
          50: withAlpha('--color-accent-50'),
          100: withAlpha('--color-accent-100'),
          200: withAlpha('--color-accent-200'),
          300: withAlpha('--color-accent-300'),
          400: withAlpha('--color-accent-400'),
          500: withAlpha('--color-accent-500'),
          600: withAlpha('--color-accent-600'),
          700: withAlpha('--color-accent-700'),
          800: withAlpha('--color-accent-800'),
          900: withAlpha('--color-accent-900'),
          950: withAlpha('--color-accent-950')
        },
        // 深色模式背景色阶
        dark: {
          50: withAlpha('--color-dark-50'),
          100: withAlpha('--color-dark-100'),
          200: withAlpha('--color-dark-200'),
          300: withAlpha('--color-dark-300'),
          400: withAlpha('--color-dark-400'),
          500: withAlpha('--color-dark-500'),
          600: withAlpha('--color-dark-600'),
          700: withAlpha('--color-dark-700'),
          800: withAlpha('--color-dark-800'),
          900: withAlpha('--color-dark-900'),
          950: withAlpha('--color-dark-950')
        },
        // 中性灰阶（默认对齐 Tailwind Cool Gray，皮肤可重映射为商务蓝灰）
        gray: {
          50: withAlpha('--color-gray-50'),
          100: withAlpha('--color-gray-100'),
          200: withAlpha('--color-gray-200'),
          300: withAlpha('--color-gray-300'),
          400: withAlpha('--color-gray-400'),
          500: withAlpha('--color-gray-500'),
          600: withAlpha('--color-gray-600'),
          700: withAlpha('--color-gray-700'),
          800: withAlpha('--color-gray-800'),
          900: withAlpha('--color-gray-900'),
          950: withAlpha('--color-gray-950')
        }
      },
      fontFamily: {
        sans: [
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: 'var(--shadow-glass, 0 8px 32px rgba(0, 0, 0, 0.08))',
        'glass-sm': 'var(--shadow-glass-sm, 0 4px 16px rgba(0, 0, 0, 0.06))',
        glow: 'var(--shadow-glow, 0 0 20px rgba(20, 184, 166, 0.25))',
        'glow-lg': 'var(--shadow-glow-lg, 0 0 40px rgba(20, 184, 166, 0.35))',
        card: 'var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06))',
        'card-hover': 'var(--shadow-card-hover, 0 10px 40px rgba(0, 0, 0, 0.08))',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary':
          'var(--bg-gradient-primary, linear-gradient(135deg, #14b8a6 0%, #0d9488 100%))',
        'gradient-dark': 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'var(--bg-mesh-gradient, radial-gradient(at 40% 20%, rgba(20, 184, 166, 0.12) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(6, 182, 212, 0.08) 0px, transparent 50%), radial-gradient(at 0% 50%, rgba(20, 184, 166, 0.08) 0px, transparent 50%))'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(20, 184, 166, 0.25)' },
          '100%': { boxShadow: '0 0 30px rgba(20, 184, 166, 0.4)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
