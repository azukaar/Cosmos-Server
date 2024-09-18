import React, { useState } from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Grid, Chip, Stack } from '@mui/material';
import "./flag.css";
import { useTranslation } from 'react-i18next';

const languages = [
  { code: 'en', name: 'English (US)', flag: 'us' },
  { code: 'en-GB', name: 'English (UK)', flag: 'gb' },
  { code: 'ar', name: 'العربية', flag: 'sa' },
  { code: 'cn', name: '中文', flag: 'cn' },
  { code: 'de', name: 'Deutsch', flag: 'de' },
  { code: 'de-CH', name: 'Deutsch (CH)', flag: 'ch' },
  { code: 'es', name: 'Español', flag: 'es' },
  { code: 'fr', name: 'Français', flag: 'fr' },
  { code: 'hi', name: 'हिन्दी', flag: 'in' },
  { code: 'it', name: 'Italiano', flag: 'it' },
  { code: 'jp', name: '日本語', flag: 'jp' },
  { code: 'kr', name: '한국어', flag: 'kr' },
  { code: 'nl', name: 'Nederlands', flag: 'nl' },
  { code: 'pl', name: 'Polski', flag: 'pl' },
  { code: 'pt', name: 'Português', flag: 'pt' },
  { code: 'ru', name: 'Русский', flag: 'ru' },
  { code: 'tr', name: 'Türkçe', flag: 'tr' },
  { code: 'en-FUNNYSHAKESPEARE', name: 'Shakespeare', flag: 'gb-eng' },
];

export const LanguagesSelectModal = ({ open, setOpen }) => {
  const { t, i18n } = useTranslation();
  const [selectedLanguage, setSelectedLanguage] = useState(i18n.language);

  const handleLanguageChange = (langCode) => {
    setSelectedLanguage(langCode);
    i18n.changeLanguage(langCode);
  };

  const handleClose = () => {
    setOpen(false);
  };

  const handleSave = () => {
    i18n.changeLanguage(selectedLanguage);
    setOpen(false);
  };

  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>{t('language.selectLanguage')}</DialogTitle>
      <DialogContent>
        <Stack style={{ marginTop: '10px', maxHeight: '260px' }} direction='column' spacing={2}>
          {languages.map((lang) => (
              <Chip
                key={lang.code}
                label={<div style={{
                  minWidth: "150px",
                }}>
                  <span className={`fi fi-${lang.flag}`} style={{ marginRight: '8px' }}></span>
                  <span >{lang.name}</span>
                </div>}
                onClick={() => handleLanguageChange(lang.code)}
                color={selectedLanguage === lang.code ? "primary" : "default"}
                variant={selectedLanguage === lang.code ? "filled" : "outlined"}
                sx={{ padding: '10px 0px' }}
              />
          ))}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{t('global.close')}</Button>
      </DialogActions>
    </Dialog>
  );
};

export const LanguagesSelect = () => {
  const [open, setOpen] = useState(false);
  const { i18n } = useTranslation();

  return (
    <>
      <Chip
        label={<span className={`fi fi-${languages.find(lang => lang.code === i18n.language)?.flag || 'gb'}`}></span>}
        sx={{ height: 24, '& .MuiChip-label': { fontSize: '0.85rem', py: 0.25 } }}
        component="a"
        clickable
        onClick={() => setOpen(true)}
      />
      <LanguagesSelectModal open={open} setOpen={setOpen} />
    </>
  );
};