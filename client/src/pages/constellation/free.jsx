import React, { useState } from 'react';
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

const VPNSalesPage = () => {
  const [isYearly, setIsYearly] = useState(true);

  const monthlyPrice = 10.30;
  const yearlyPrice = 8.25;

  const discountedMonthlyPrice = (monthlyPrice * 0.85).toFixed(2);
  const discountedYearlyPrice = (yearlyPrice * 0.85).toFixed(2);

  const monthlyLink = "https://buy.stripe.com/5kAbMN4qrbkVcBGcMR?prefilled_promo_code=EARLY15";
  const yearlyLink = "https://buy.stripe.com/cN2bMN9KLbkV59e9AE?prefilled_promo_code=EARLY15";

  return (
    <Container maxWidth="md">
      <Box my={4}>
        <img src={banner} alt="Constellation VPN Banner" style={{ width: '100%', borderRadius: '8px', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }} />
      </Box>

      <Typography variant="h3" gutterBottom>Unlock Constellation: Your Secure Gateway Home</Typography>

      <Typography variant="body1" paragraph>
        Constellation is a powerful VPN based technology that allows you to securely access your home server from anywhere, without the need to open ports on your router. Keep your data safe and your connections secure with our state-of-the-art encryption technology.
      </Typography>

      <Card variant="outlined" sx={{ mb: 4 }}>
        <CardHeader 
          title={<Typography variant="h5">Why Constellation VPN</Typography>}
        />
        <CardContent>
          <List>
            {[
              'Securely access your home server from anywhere in the world',
              'No need to open ports, reducing potential security vulnerabilities',
              'Encrypted connections keep your data safe from prying eyes',
              'Easy setup and management through the Cosmos interface',
              'Automatically switches from internet to local network when you get home',
              'Automatic DNS rewrite',
              'Block ads and trackers on all devices',
              'Support the ongoing development of new features and improvements of Cosmos'
            ].map((item, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  {index === 7 && <span>❤️</span>}
                  {index !== 7 &&
                  <CheckIcon color="primary" />}
                </ListItemIcon>
                <ListItemText primary={item} />
              </ListItem>
            ))}
          </List>
        </CardContent>
      </Card>

      <Box display="flex" justifyContent="center" alignItems="center" mb={4}>
        <Typography>Monthly</Typography>
        <Switch 
          checked={isYearly}
          onChange={() => setIsYearly(!isYearly)}
          color="primary"
        />
        <Typography>Yearly</Typography>
      </Box>

      <Card variant="outlined" sx={{ maxWidth: 400, margin: 'auto' }}>
        <CardHeader 
          title={<Typography variant="h4" align="center">{isYearly ? 'Yearly Plan' : 'Monthly Plan'}</Typography>}
        />
        <CardContent>
          <Box textAlign="center" mb={2}>
            <Typography variant="h3" style={{ textDecoration: 'line-through', color: 'text.secondary' }}>
              ${isYearly ? yearlyPrice.toFixed(2) : monthlyPrice.toFixed(2)}
            </Typography>
            <Typography variant="h3" color="primary">
              ${isYearly ? discountedYearlyPrice : discountedMonthlyPrice}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              per month
            </Typography>
            <Chip 
              label="EARLY15: 15% OFF FOR LIFE" 
              color="secondary" 
              sx={{ mt: 1 }}
            />
            <Typography variant="body2" color="text.secondary" mt={1}>
              Limited time offer until Febrary 2025 for <strong>early adopters!</strong>
            </Typography>
          </Box>
          <List>
            {[
              'Unlimited devices',
              'All VPN features',
              isYearly ? 'Save 17% compared to monthly' : 'Flexible monthly billing',
              '15% lifetime discount applied'
            ].map((item, index) => (
              <ListItem key={index}>
                <ListItemIcon>
                  <CheckIcon color="primary" />
                </ListItemIcon>
                <ListItemText primary={item} />
              </ListItem>
            ))}
          </List>
          <Button variant="contained" color="primary" fullWidth href={isYearly ? yearlyLink : monthlyLink}>
            Upgrade Now
          </Button>
        </CardContent>
      </Card>
    </Container>
  );
};

export default VPNSalesPage;