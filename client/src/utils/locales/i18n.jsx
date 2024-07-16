import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import resourcesToBackend from 'i18next-resources-to-backend'

i18n
  // detect user language
  .use(LanguageDetector)
  // pass the i18n instance to react-i18next.
  .use(initReactI18next)
  //load resources as Backend (supports lazy loading)
  .use(resourcesToBackend((language, namespace) => import(`./${language}/${namespace}.json`)))
  // init i18next
  .init({
    keySeparator: false,
    debug: false,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // not needed for react as it escapes by default
    },
    react: {
      transKeepBasicHtmlNodesFor: ['br', 'strong', 'i', 'p', 'pre', 'b']
    }
  });

  export const getLanguage = () => {
    return i18n.resolvedLanguage || i18n.language ||
      (typeof window !== 'undefined' && window.localStorage.i18nextLng) ||
      'en';
  };

export { i18n };