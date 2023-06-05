import { Link } from 'react-router-dom';

// material-ui
import { Grid, Stack, Typography } from '@mui/material';

// project import
import AuthLogin from './auth-forms/AuthLogin';
import AuthWrapper from './AuthWrapper';

const selfport = new URL(window.location.href).port
const selfprotocol = new URL(window.location.href).protocol + "//"
const selfHostname = selfprotocol + (new URL(window.location.href).hostname) + (selfport ? ":" + selfport : "")

// ================================|| LOGIN ||================================ //

const Login = () => (
    <AuthWrapper>
        <link rel="openid2.provider openid.server" href={selfHostname + "/oauth2/auth"} />
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                    <Typography variant="h3">Login</Typography>
                    {/* <Typography component={Link} to="/register" variant="body1" sx={{ textDecoration: 'none' }} color="primary">
                        Don&apos;t have an account?
                    </Typography> */}
                </Stack>
            </Grid>
            <Grid item xs={12}>
                <AuthLogin />
            </Grid>
        </Grid>
    </AuthWrapper>
);

export default Login;
