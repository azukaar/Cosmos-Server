import { Link } from 'react-router-dom';

// material-ui
import { Grid, Stack, Typography } from '@mui/material';

// project import
import AuthRegister from './auth-forms/AuthRegister';
import AuthWrapper from './AuthWrapper';
import { useEffect } from 'react';

import * as API from '../../api';

// ================================|| REGISTER ||================================ //

const Logout = () => {
    useEffect(() => {
      API.auth.logout()
       .then(() => {
          setTimeout(() => {
            window.location.href = '/cosmos-ui/login';
          }, 2000);
        });
    },[]);

    return <AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
              <Typography variant="h3">
                  You have been logged off. Redirecting you...
              </Typography>
            </Grid>
        </Grid>
    </AuthWrapper>;
}

export default Logout;
