import React, { createContext, useContext, useState, useEffect } from 'react';
import en from './locales/en.json';
import fa from './locales/fa.json';

const dictionaries = { en, fa };

const I18nContext = createContext({
  language: 'en',
  direction: 'ltr',
  t: (key) => key,
  setLanguage: () => {},
  toggleLanguage: () => {},
});

export const I18nProvider = ({ children }) => {
  const [language, setLanguageState] = useState(() => {
    return localStorage.getItem('oryn_lang') || 'en';
  });

  const direction = language === 'fa' ? 'rtl' : 'ltr';

  useEffect(() => {
    localStorage.setItem('oryn_lang', language);
    document.documentElement.setAttribute('lang', language);
    document.documentElement.setAttribute('dir', direction);
  }, [language, direction]);

  const setLanguage = (lang) => {
    if (dictionaries[lang]) {
      setLanguageState(lang);
    }
  };

  const toggleLanguage = () => {
    setLanguageState((prev) => (prev === 'en' ? 'fa' : 'en'));
  };

  // Helper for nested dot-notation keys: e.g. "hero.title"
  const t = (path) => {
    const keys = path.split('.');
    let current = dictionaries[language] || dictionaries.en;
    for (const k of keys) {
      if (current && typeof current === 'object' && k in current) {
        current = current[k];
      } else {
        // Fallback to English dictionary if key missing in current language
        let fallback = dictionaries.en;
        for (const fbKey of keys) {
          if (fallback && typeof fallback === 'object' && fbKey in fallback) {
            fallback = fallback[fbKey];
          } else {
            return path;
          }
        }
        return fallback;
      }
    }
    return current || path;
  };

  return (
    <I18nContext.Provider value={{ language, direction, t, setLanguage, toggleLanguage }}>
      {children}
    </I18nContext.Provider>
  );
};

export const useTranslation = () => useContext(I18nContext);
