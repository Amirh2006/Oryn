import React, { useState } from 'react';
import { useTranslation } from '../i18n/I18nContext';
import { useTheme } from '../theme/ThemeContext';
import { 
  Film, 
  Compass, 
  Sparkles, 
  Bookmark, 
  Sun, 
  Moon, 
  Languages, 
  Menu, 
  X, 
  Search 
} from 'lucide-react';

export const Header = () => {
  const { t, language, toggleLanguage } = useTranslation();
  const { isDark, toggleTheme } = useTheme();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navItems = [
    { key: 'nav.home', icon: Film, href: '#' },
    { key: 'nav.explore', icon: Compass, href: '#explore' },
    { key: 'nav.assistant', icon: Sparkles, href: '#assistant', highlight: true },
    { key: 'nav.collection', icon: Bookmark, href: '#collection' },
  ];

  return (
    <header className="sticky top-0 z-50 backdrop-blur-md bg-white/80 dark:bg-cinema-900/80 border-b border-light-border dark:border-cinema-border transition-colors duration-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          
          {/* Logo & Brand */}
          <div className="flex items-center gap-3">
            <a href="#" className="flex items-center gap-2 group">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-oryn-accent to-oryn-600 flex items-center justify-center shadow-goldGlow transition-transform group-hover:scale-105">
                <Sparkles className="w-5 h-5 text-white" />
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-bold tracking-tight text-gray-900 dark:text-white flex items-center gap-1">
                  {t('brand.name')}
                  <span className="text-oryn-500 text-xs px-1.5 py-0.5 rounded-full bg-oryn-500/10 dark:bg-oryn-500/20 font-medium">
                    AI
                  </span>
                </span>
                <span className="text-[10px] text-gray-500 dark:text-gray-400 font-medium hidden sm:inline">
                  {t('brand.tagline')}
                </span>
              </div>
            </a>
          </div>

          {/* Desktop Navigation Links */}
          <nav className="hidden md:flex items-center gap-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <a
                  key={item.key}
                  href={item.href}
                  className={`flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium transition-all duration-150 ${
                    item.highlight
                      ? 'text-oryn-600 dark:text-oryn-gold hover:bg-oryn-500/10'
                      : 'text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-cinema-850'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  <span>{t(item.key)}</span>
                </a>
              );
            })}
          </nav>

          {/* Search Mockup & Action Controls */}
          <div className="flex items-center gap-2">
            
            {/* Quick Search trigger */}
            <div className="hidden lg:flex items-center relative">
              <Search className="w-4 h-4 text-gray-400 absolute start-3 pointer-events-none" />
              <input
                type="text"
                placeholder={t('search.placeholder')}
                className="w-64 h-9 ps-9 pe-4 text-xs rounded-full bg-gray-100 dark:bg-cinema-850 border border-light-border dark:border-cinema-border text-gray-900 dark:text-gray-200 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-oryn-500/40 focus:border-oryn-500 transition-all"
              />
            </div>

            {/* Language Toggle (FA / EN) */}
            <button
              onClick={toggleLanguage}
              title={t('language.switch')}
              className="flex items-center gap-1.5 px-3 h-9 rounded-lg text-xs font-semibold bg-gray-100 hover:bg-gray-200 dark:bg-cinema-850 dark:hover:bg-cinema-800 text-gray-700 dark:text-gray-200 border border-light-border dark:border-cinema-border transition-colors"
            >
              <Languages className="w-3.5 h-3.5 text-oryn-500" />
              <span>{language === 'en' ? 'فارسی' : 'EN'}</span>
            </button>

            {/* Theme Toggle (Dark / Light) */}
            <button
              onClick={toggleTheme}
              title={t('theme.toggle')}
              className="w-9 h-9 flex items-center justify-center rounded-lg bg-gray-100 hover:bg-gray-200 dark:bg-cinema-850 dark:hover:bg-cinema-800 text-gray-700 dark:text-gray-200 border border-light-border dark:border-cinema-border transition-colors"
            >
              {isDark ? (
                <Sun className="w-4 h-4 text-amber-400" />
              ) : (
                <Moon className="w-4 h-4 text-indigo-600" />
              )}
            </button>

            {/* Mobile Menu Button */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="md:hidden w-9 h-9 flex items-center justify-center rounded-lg text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-cinema-850"
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>

        </div>

        {/* Mobile Navigation Drawer */}
        {mobileMenuOpen && (
          <div className="md:hidden py-4 border-t border-light-border dark:border-cinema-border animate-in slide-in-from-top-2">
            <div className="mb-3 px-2">
              <div className="flex items-center relative">
                <Search className="w-4 h-4 text-gray-400 absolute start-3 pointer-events-none" />
                <input
                  type="text"
                  placeholder={t('search.placeholder')}
                  className="w-full h-10 ps-9 pe-4 text-sm rounded-lg bg-gray-100 dark:bg-cinema-850 border border-light-border dark:border-cinema-border text-gray-900 dark:text-gray-200 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-oryn-500/40"
                />
              </div>
            </div>
            <div className="flex flex-col gap-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <a
                    key={item.key}
                    href={item.href}
                    onClick={() => setMobileMenuOpen(false)}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-cinema-850"
                  >
                    <Icon className="w-4 h-4 text-oryn-500" />
                    <span>{t(item.key)}</span>
                  </a>
                );
              })}
            </div>
          </div>
        )}

      </div>
    </header>
  );
};
