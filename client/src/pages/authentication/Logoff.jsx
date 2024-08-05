import { Link } from 'react-router-dom';

// material-ui
import { Grid, Stack, Typography } from '@mui/material';

// project import
import AuthRegister from './auth-forms/AuthRegister';
import AuthWrapper from './AuthWrapper';
import { useEffect } from 'react';

import * as API from '../../api';
import { redirectTo, redirectToLocal } from '../../utils/indexs';
import { useTranslation } from 'react-i18next';

// ================================|| REGISTER ||================================ //

const Logout = () => {
  const { t } = useTranslation();
    useEffect(() => {
      API.auth.logout()
       .then(() => {
          setTimeout(() => {
            redirectToLocal('/cosmos-ui/login');
          }, 2000);
        });
    },[]);

    return <AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
              <Typography variant="h3">
                  {t('auth.logoffText')}
              </Typography>
            </Grid>
        </Grid>
    </AuthWrapper>;
}

export default Logout;
