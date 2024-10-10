import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { 
  Button, 
  Card, 
  CardContent, 
  CardHeader, 
  Switch, 
  Chip,
  Typography, 
  List, 
  ListItem, 
  ListItemIcon, 
  ListItemText,
  Container,
  Box
} from '@mui/material';
import { Warning as WarningIcon, Check as CheckIcon } from '@mui/icons-material';
import banner from '../../assets/images/const_banner.png';
import { getCurrencyFromLanguage } from '../../utils/indexs';
import PremiumSalesPage from '../../utils/free';

const VPNSalesPage = () => {
  const { t, i18n } = useTranslation();

  return <PremiumSalesPage salesKey="constellation" extra={
    <Typography variant="body2" color="text.secondary">
      {t(`mgmt.sales.constellation.lighthouse_note`)}
    </Typography>
  }/>
};


export default VPNSalesPage;