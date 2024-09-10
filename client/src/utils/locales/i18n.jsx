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
      'en-GB',
      'cn',
      'de',
      'de-CH',
      'es',
      'fr',
      'hi',
      'it',
      'jp',
      'kr',
      'nl',
      'pl',
      'pt',
      'ru',
      'tr',
      'ar',
      'en-FUNNYSHAKESPEARE',
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
    fr: () => import('dayjs/locale/fr'),
    es: () => import('dayjs/locale/es'),
    cn: () => import('dayjs/locale/zh-cn'),
    hi: () => import('dayjs/locale/hi'),
    it: () => import('dayjs/locale/it'),
    jp: () => import('dayjs/locale/ja'),
    kr: () => import('dayjs/locale/ko'),
    nl: () => import('dayjs/locale/nl'),
    pl: () => import('dayjs/locale/pl'),
    pt: () => import('dayjs/locale/pt'),
    ru: () => import('dayjs/locale/ru'),
    tr: () => import('dayjs/locale/tr'),
    ar: () => import('dayjs/locale/ar-sa'),
    enFUNNYSHAKESPEARE: () => import('dayjs/locale/en-gb'),
  }
  
  export function dayjsLocale (language) {
    locales[language.replace('-', '')]().then(() => dayjs.locale(language.toLowerCase()))
  }

export { i18n };