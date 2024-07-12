import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import en from "./en/translation.json";
import de from "./de/translation.json";

const resources = {
  en: {
    translation: en
  },
  de: {
    translation: de
  }
};

i18n
  // detect user language
  .use(LanguageDetector)
  // pass the i18n instance to react-i18next.
  .use(initReactI18next)
  // init i18next
  .init({
    resources,
    keySeparator: false,
    debug: true,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // not needed for react as it escapes by default
    },
    react: {
      transKeepBasicHtmlNodesFor: ['br', 'strong', 'i', 'p', 'pre', 'b']
    }
  });

export { i18n };