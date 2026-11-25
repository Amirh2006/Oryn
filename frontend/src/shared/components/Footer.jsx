import React from 'react';
import { useTranslation } from '../i18n/I18nContext';
import { Sparkles, Activity } from 'lucide-react';

export const Footer = () => {
  const { t } = useTranslation();

  return (
    <footer className="mt-auto border-t border-light-border dark:border-cinema-border bg-white dark:bg-cinema-900 transition-colors duration-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-gray-500 dark:text-gray-400">
          
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 rounded-md bg-oryn-500 flex items-center justify-center text-white">
              <Sparkles className="w-3.5 h-3.5" />
            </div>
            <span className="font-semibold text-gray-900 dark:text-white">{t('brand.name')}</span>
            <span>—</span>
            <span>{t('brand.tagline')}</span>
          </div>

          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[11px] font-medium border border-emerald-500/20">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              <Activity className="w-3 h-3" />
              {t('footer.status')}
            </span>
          </div>

          <div className="text-center sm:text-end">
            <p>{t('footer.copyright')}</p>
            <p className="text-[10px] text-gray-400 dark:text-gray-500 mt-0.5">{t('footer.disclaimer')}</p>
          </div>

        </div>
      </div>
    </footer>
  );
};
