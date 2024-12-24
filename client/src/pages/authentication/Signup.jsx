import { Link } from 'react-router-dom';

// material-ui
import { Grid, Stack, Typography } from '@mui/material';

// project import
import FirebaseRegister from './auth-forms/AuthRegister';
import AuthWrapper from './AuthWrapper';

// ================================|| REGISTER ||================================ //

const Signup = () => (
    <AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                    <Typography variant="h3">{t('auth.signUp.signUpTitle')}</Typography>
                    <Typography component={Link} to="/cosmos-ui/login" variant="body1" sx={{ textDecoration: 'none' }} color="primary">
                        {t('auth.signUp.signUpCancelToLogin')}
                    </Typography>
                </Stack>
            </Grid>
            <Grid item xs={12}>
                <FirebaseRegister />
            </Grid>
        </Grid>
    </AuthWrapper>
);

export default Signup;
