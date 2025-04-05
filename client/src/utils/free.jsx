import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { 
  Button, 
  Card, 
  Stack,
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
import bannerConst from '../assets/images/const_banner.png';
import bannerStor from '../assets/images/stor_banner.png';
import { getCurrencyFromLanguage } from './indexs';

const PremiumSalesPage = ({salesKey, extra}) => {
  const banners = {
    "constellation": bannerConst,
    "remote": bannerStor,
    "backup": bannerStor,
  }

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

  const lifePrices = {
    "USD": 249.00,
    "EUR": 229.00,
    "GBP": 199.00
  }

  const monthlyPrice = monthlyPrices[currency];
  const yearlyPrice = yearlyPrices[currency];
  const lifePrice = lifePrices[currency];

  const discountedMonthlyPrice = (monthlyPrice * 0.85).toFixed(2);
  const discountedYearlyPrice = (yearlyPrice * 0.85).toFixed(2);
  const discountedLifePrice = (lifePrice * 0.85).toFixed(2);


  const monthlyLink = "https://buy.stripe.com/5kAbMN4qrbkVcBGcMR?prefilled_promo_code=EARLY15";
  const yearlyLink = "https://buy.stripe.com/cN2bMN9KLbkV59e9AE?prefilled_promo_code=EARLY15";
  const lifeLink = "https://buy.stripe.com/8wM1896yz74F59e6ov?prefilled_promo_code=LIFE15";

  const featureKeys = Array.from({ length: 6 }, (_, i) => `mgmt.sales.${salesKey}.features.${i}`);
  const planFeatureKeys = Array.from({ length: 5 }, (_, i) => `mgmt.sales.plan_features.${i}`);

  return (
    <Container maxWidth="md">
      <Box my={4}>
        <img src={banners[salesKey]} alt={t(`mgmt.sales.${salesKey}.banner_alt`)} style={{ width: '100%', borderRadius: '8px', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }} />
      </Box>

      <Typography variant="h3" gutterBottom>{t(`mgmt.sales.${salesKey}.title`)}</Typography>

      <Typography variant="body1" paragraph>
        {t(`mgmt.sales.${salesKey}.description`)}
      </Typography>

      <Card variant="outlined" sx={{ mb: 4 }}>
        <CardHeader 
          title={<Typography variant="h5">{t(`mgmt.sales.${salesKey}.why_title`)}</Typography>}
        />
        <CardContent>
          <List>
            {featureKeys.map((key, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  <CheckIcon color="primary" />
                </ListItemIcon>
                <ListItemText primary={t(key)} />
              </ListItem>
            ))}
            <ListItem key={"last"}>
              <ListItemIcon>
              <span>❤️</span>
              </ListItemIcon>
              <ListItemText primary={t(`mgmt.sales.support`)} />
            </ListItem>
          </List>
          {extra}
        </CardContent>
      </Card>

      <Box display="flex" justifyContent="center" alignItems="center" mb={4}>
        <Typography>{t('mgmt.sales.monthly')}</Typography>
        <Switch 
          checked={isYearly}
          onChange={() => setIsYearly(!isYearly)}
          color="primary"
        />
        <Typography>{t('mgmt.sales.yearly')}</Typography>
      </Box>

      <Stack spacing={2} direction="row" justifyContent="center" alignItems="center">
      <Card variant="outlined" sx={{ width: 400, margin: 'auto' }}>
        <CardHeader 
          title={<Typography variant="h4" align="center">{t(isYearly ? 'mgmt.sales.yearly_plan' : 'mgmt.sales.monthly_plan')}</Typography>}
        />
        <CardContent>
          <Box textAlign="center" mb={2}>
            <Typography variant="h3" color="primary">
            {/* <Typography variant="h3" style={{ textDecoration: 'line-through', color: 'text.secondary' }}> */}
              {currencyCode} {isYearly ? (yearlyPrice / 12).toFixed(2) : monthlyPrice.toFixed(2)}
            </Typography>
            {/* <Typography variant="h3" color="primary">
              {currencyCode} {isYearly ? ((discountedYearlyPrice / 12).toFixed(2)) : discountedMonthlyPrice}
            </Typography> */}
            <Typography variant="body2" color="text.secondary">
              {t('mgmt.sales.per_month')}
            </Typography>
            {/* <Chip 
              label={t('mgmt.sales.discount_chip')}
              color="secondary" 
              sx={{ mt: 1 }}
            /> */}
            {/* <Typography variant="body2" color="text.secondary" mt={1}>
              {t('mgmt.sales.early_adopter_offer')}
            </Typography> */}
          </Box>
          <List>
            {planFeatureKeys.map((key, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  <CheckIcon color="primary" />
                </ListItemIcon>
                <ListItemText 
                  primary={t(key)} 
                />
              </ListItem>
            ))}
            <ListItem key={"last"}>
              <ListItemIcon>
                <CheckIcon color="primary" />
              </ListItemIcon>
              <ListItemText 
                primary={t('mgmt.sales.subs_pro')} 
              />
            </ListItem>
          </List>
          <Button variant="contained" color="primary" target='_blank' fullWidth href={isYearly ? yearlyLink : monthlyLink}>
            {t('mgmt.sales.upgrade_button')}
          </Button>
        </CardContent>
      </Card>
      <Card variant="outlined" sx={{ width: 400, margin: 'auto' }}>
        <CardHeader 
          title={<Typography variant="h4" align="center">{t('mgmt.sales.life_plan')}</Typography>}
        />
        <CardContent>
          <Box textAlign="center" mb={2}>
            
            <Typography variant="h3" color="primary">
            {/* <Typography variant="h3" style={{ textDecoration: 'line-through', color: 'text.secondary' }}> */}
              {currencyCode} {lifePrice.toFixed(2)}
            </Typography>
            {/* <Typography variant="h3" color="primary">
              {currencyCode} {discountedLifePrice}
            </Typography> */}
            <Typography variant="body2" color="text.secondary">
              -
            </Typography>
            {/* <Chip 
              label={t('mgmt.sales.discount_chip')}
              color="secondary" 
              sx={{ mt: 1 }}
            />
            <Typography variant="body2" color="text.secondary" mt={1}>
              {t('mgmt.sales.early_adopter_offer2')}
            </Typography> */}
          </Box>
          <List>
            {planFeatureKeys.map((key, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  <CheckIcon color="primary" />
                </ListItemIcon>
                <ListItemText 
                  primary={t(key)} 
                />
              </ListItem>
            ))}
            <ListItem key={"last"}>
              <ListItemIcon>
                <CheckIcon color="primary" />
              </ListItemIcon>
              <ListItemText 
                primary={t('mgmt.sales.life_pro')} 
              />
            </ListItem>
          </List>
          <Button variant="contained" color="primary" target='_blank' fullWidth href={lifeLink}>
            {t('mgmt.sales.upgrade_button')}
          </Button>
        </CardContent>
      </Card>
      </Stack>
    </Container>
  );
};

export default PremiumSalesPage;