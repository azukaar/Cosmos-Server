import { Link } from 'react-router-dom';

// material-ui
import { Grid, Stack, Typography } from '@mui/material';

// project import
import AuthRegister from './auth-forms/AuthRegister';
import AuthWrapper from './AuthWrapper';

// ================================|| REGISTER ||================================ //

const Register = () => {
    const urlSearchParams = new URLSearchParams(window.location.search);
    const formType = urlSearchParams.get('t');
    const isInviteLink = formType === '2';
    const isRegister = formType === '1';
    const nickname = urlSearchParams.get('nickname');
    const regkey = urlSearchParams.get('key');

    return <AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                    <Typography variant="h3">{
                        isInviteLink ? 'Invitation' : 'Password Reset'
                    }</Typography>
                </Stack>
            </Grid>
            <Grid item xs={12}>
                <AuthRegister 
                    nickname={nickname}
                    isRegister={isRegister}
                    isInviteLink={isInviteLink}
                    regkey={regkey}
                />
            </Grid>
        </Grid>
    </AuthWrapper>;
}

export default Register;
