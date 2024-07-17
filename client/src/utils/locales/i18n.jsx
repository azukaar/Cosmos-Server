import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import resourcesToBackend from 'i18next-resources-to-backend';
import dayjs from 'dayjs';

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
    supportedLngs: [
      'en',
      'de',
      'de-CH'
    ],
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

  const locales = {
    en: () => import('dayjs/locale/en'),
    enGB: () => import('dayjs/locale/en-gb'),
    de: () => import('dayjs/locale/de'),
    deCH: () => import('dayjs/locale/de-ch'),
  }
  
  export function dayjsLocale (language) {
    locales[language.replace('-', '')]().then(() => dayjs.locale(language.toLowerCase()))
  }

export { i18n };