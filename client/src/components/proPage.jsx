import * as React from 'react';
import { Alert, Box, CircularProgress } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router';

import { useIsPro } from '../utils/pro';
import { useClientInfos } from '../utils/hooks';
import { PERM_RESOURCES_READ } from '../utils/permissions';
import PremiumSalesPage from '../utils/free';

// Gate for Pro-only pages: admin check, licence spinner, sales page when unlicensed
// or when `component` is null (community build stub in pro/index.js).
const ProPage = ({ children, component }) => {
  const { t } = useTranslation();
  const { hasPermission } = useClientInfos();
  const isPro = useIsPro();

  if (!hasPermission(PERM_RESOURCES_READ)) {
    return <Alert severity="error">{t('global.adminOnly')}</Alert>;
  }

  if (isPro === null) {
    return <Box style={{ display: 'flex', justifyContent: 'center', marginTop: '150px' }}>
      <CircularProgress size={100} />
    </Box>;
  }

  if (!isPro || !component) {
    return <PremiumSalesPage salesKey="constellation" />;
  }

  return children;
};

// Route element for a Pro page: renders `component` behind the ProPage gate, with URL params as props.
export const ProRoute = ({ component: Component }) => {
  const params = useParams();
  return <ProPage component={Component}>
    {Component ? <Component {...params} /> : null}
  </ProPage>;
};

export default ProPage;
