import React from 'react';
import { useTranslation } from './shared/i18n/I18nContext';
import { Header } from './shared/components/Header';
import { Footer } from './shared/components/Footer';
import { HeroBannerSkeleton, SectionSkeleton } from './shared/components/Skeleton';
import { Sparkles, Compass, Film, Flame, Award, SlidersHorizontal } from 'lucide-react';

export const App = () => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col min-h-screen bg-light-bg dark:bg-cinema-950 transition-colors duration-200">
      <Header />

      <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8 space-y-12">
        {/* Hero Section */}
        <section className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-oryn-500/10 via-transparent to-amber-500/5 dark:from-oryn-900/30 dark:via-cinema-900/40 dark:to-cinema-950 border border-light-border dark:border-cinema-border p-8 sm:p-14">
          <div className="relative z-10 max-w-2xl space-y-6">
            
            <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-oryn-500/10 dark:bg-oryn-500/20 border border-oryn-500/30 text-oryn-600 dark:text-oryn-gold text-xs font-semibold">
              <Sparkles className="w-3.5 h-3.5" />
              <span>{t('hero.badge')}</span>
            </div>

            <h1 className="text-3xl sm:text-5xl font-extrabold tracking-tight text-gray-900 dark:text-white leading-tight">
              {t('hero.title')}
            </h1>

            <p className="text-sm sm:text-base text-gray-600 dark:text-gray-300 leading-relaxed">
              {t('hero.description')}
            </p>

            <div className="flex flex-wrap items-center gap-4 pt-2">
              <a
                href="#explore"
                className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-gradient-to-r from-oryn-500 to-oryn-600 hover:from-oryn-600 hover:to-oryn-700 text-white font-medium text-sm shadow-goldGlow transition-transform active:scale-95"
              >
                <Compass className="w-4 h-4" />
                <span>{t('hero.ctaExplore')}</span>
              </a>

              <a
                href="#assistant"
                className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-gray-100 hover:bg-gray-200 dark:bg-cinema-850 dark:hover:bg-cinema-800 text-gray-900 dark:text-white font-medium text-sm border border-light-border dark:border-cinema-border transition-colors active:scale-95"
              >
                <Sparkles className="w-4 h-4 text-oryn-500" />
                <span>{t('hero.ctaAssistant')}</span>
              </a>
            </div>

          </div>

          {/* Decorative Background Orb */}
          <div className="absolute end-0 top-1/2 -translate-y-1/2 w-96 h-96 bg-gradient-to-br from-oryn-500/20 to-transparent rounded-full blur-3xl pointer-events-none" />
        </section>

        {/* Discovery Feed Preview & Skeletons */}
        <section className="space-y-10">
          <SectionSkeleton count={5} />
          <SectionSkeleton count={5} />
        </section>
      </main>

      <Footer />
    </div>
  );
};

export default App;
