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

const VPNSalesPage = () => {
  const [isYearly, setIsYearly] = useState(true);
  const { t, i18n } = useTranslation();
  const currency = getCurrencyFromLanguage();
  const currencyCode = {
    "USD": "$",
    "EUR": "€",
    "GBP": "£"
  }[currency];

  const monthlyPrices = {
    "USD": 10.90,
    "EUR": 9.90,
    "GBP": 8.90
  }

  const yearlyPrices = {
    "USD": 99.00,
    "EUR": 93.00,
    "GBP": 88.00
  }

  const monthlyPrice = monthlyPrices[currency];
  const yearlyPrice = yearlyPrices[currency];

  const discountedMonthlyPrice = (monthlyPrice * 0.85).toFixed(2);
  const discountedYearlyPrice = (yearlyPrice * 0.85).toFixed(2);


  const monthlyLink = "https://buy.stripe.com/5kAbMN4qrbkVcBGcMR?prefilled_promo_code=EARLY15";
  const yearlyLink = "https://buy.stripe.com/cN2bMN9KLbkV59e9AE?prefilled_promo_code=EARLY15";

  const featureKeys = Array.from({ length: 8 }, (_, i) => `mgmt.constellation.features.${i}`);
  const planFeatureKeys = Array.from({ length: 4 }, (_, i) => `mgmt.constellation.plan_features.${i}`);

  return (
    <Container maxWidth="md">
      <Box my={4}>
        <img src={banner} alt={t('mgmt.constellation.banner_alt')} style={{ width: '100%', borderRadius: '8px', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }} />
      </Box>

      <Typography variant="h3" gutterBottom>{t('mgmt.constellation.title')}</Typography>

      <Typography variant="body1" paragraph>
        {t('mgmt.constellation.description')}
      </Typography>

      <Card variant="outlined" sx={{ mb: 4 }}>
        <CardHeader 
          title={<Typography variant="h5">{t('mgmt.constellation.why_title')}</Typography>}
        />
        <CardContent>
          <List>
            {featureKeys.map((key, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  {index === 7 ? <span>❤️</span> : <CheckIcon color="primary" />}
                </ListItemIcon>
                <ListItemText primary={t(key)} />
              </ListItem>
            ))}
          </List>
          <Typography variant="body2" color="text.secondary">
            {t('mgmt.constellation.lighthouse_note')}
          </Typography>
        </CardContent>
      </Card>

      <Box display="flex" justifyContent="center" alignItems="center" mb={4}>
        <Typography>{t('mgmt.constellation.monthly')}</Typography>
        <Switch 
          checked={isYearly}
          onChange={() => setIsYearly(!isYearly)}
          color="primary"
        />
        <Typography>{t('mgmt.constellation.yearly')}</Typography>
      </Box>

      <Card variant="outlined" sx={{ maxWidth: 400, margin: 'auto' }}>
        <CardHeader 
          title={<Typography variant="h4" align="center">{t(isYearly ? 'mgmt.constellation.yearly_plan' : 'mgmt.constellation.monthly_plan')}</Typography>}
        />
        <CardContent>
          <Box textAlign="center" mb={2}>
            <Typography variant="h3" style={{ textDecoration: 'line-through', color: 'text.secondary' }}>
              {currencyCode} {isYearly ? (yearlyPrice / 12).toFixed(2) : monthlyPrice.toFixed(2)}
            </Typography>
            <Typography variant="h3" color="primary">
              {currencyCode} {isYearly ? ((discountedYearlyPrice / 12).toFixed(2)) : discountedMonthlyPrice}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t('mgmt.constellation.per_month')}
            </Typography>
            <Chip 
              label={t('mgmt.constellation.discount_chip')}
              color="secondary" 
              sx={{ mt: 1 }}
            />
            <Typography variant="body2" color="text.secondary" mt={1}>
              {t('mgmt.constellation.early_adopter_offer')}
            </Typography>
          </Box>
          <List>
            {planFeatureKeys.map((key, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  <CheckIcon color="primary" />
                </ListItemIcon>
                <ListItemText 
                  primary={
                    isYearly && index === 2 
                      ? t('mgmt.constellation.yearly_savings') 
                      : t(key)
                  } 
                />
              </ListItem>
            ))}
          </List>
          <Button variant="contained" color="primary" target='_blank' fullWidth href={isYearly ? yearlyLink : monthlyLink}>
            {t('mgmt.constellation.upgrade_button')}
          </Button>
        </CardContent>
      </Card>
    </Container>
  );
};

export default VPNSalesPage;