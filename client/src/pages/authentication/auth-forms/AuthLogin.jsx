import React, { useEffect } from 'react';
import { Link, Link as RouterLink } from 'react-router-dom';

// material-ui
import {
    Button,
    Checkbox,
    Divider,
    FormControlLabel,
    FormHelperText,
    Grid,
    IconButton,
    InputAdornment,
    InputLabel,
    OutlinedInput,
    Stack,
    Typography,
    Alert
} from '@mui/material';
import { useTranslation } from 'react-i18next';

import * as API from "../../../api";

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project import
import FirebaseSocial from './FirebaseSocial';
import AnimateButton from '../../../components/@extended/AnimateButton';

// assets
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import { LoadingButton } from '@mui/lab';
import { redirectToLocal } from '../../../utils/indexs';

// ============================|| FIREBASE - LOGIN ||============================ //

const AuthLogin = () => {
    const { t } = useTranslation();
    const [checked, setChecked] = React.useState(false);
    const [showResetPassword, setShowResetPassword] = React.useState(false);

    const [showPassword, setShowPassword] = React.useState(false);
    const handleClickShowPassword = () => {
        setShowPassword(!showPassword);
    };

    const handleMouseDownPassword = (event) => {
        event.preventDefault();
    };

    // TODO: Add ?notlogged=1 and ?invalid=1 to check for errors
    // TODO: Extract ?redirect=<URL> to redirect to a specific page after login
    const urlSearchParams = new URLSearchParams(window.location.search);
    const notLogged = urlSearchParams.get('notlogged') == 1;
    const notLoggedAdmin = urlSearchParams.get('notlogged') == 2;
    const invalid = urlSearchParams.get('invalid') == 1;
    const redirectToURL = urlSearchParams.get('redirect') ? urlSearchParams.get('redirect') : '/cosmos-ui';

    useEffect(() => {
        API.auth.me().then((data) => {
            if(data.status == 'OK') {
                redirectToLocal(redirectToURL);
            } else if(data.status == 'NEW_INSTALL') {
                redirectToLocal('/cosmos-ui/newInstall');
            }
        });


        API.config.canSendEmail().then((resp) => {
            if(resp.status == 'OK' && resp.data.canSendEmail) {
                setShowResetPassword(true);
            }
        });
    }, []);       

    return (
        <>
            { notLogged &&<Grid container spacing={2} justifyContent="center">
                <Alert severity="error">{t('auth.notLoggedInError')}</Alert>
                <br />
            </Grid>}

            { notLoggedAdmin &&<Grid container spacing={2} justifyContent="center">
                <Alert severity="error">{t('auth.notAdminError')}</Alert>
                <br />
            </Grid>}

            { invalid &&<Grid container spacing={2} justifyContent="center">
                <Alert severity="error">{t('auth.loggedOutError')}</Alert>
                <br />
            </Grid>}
            <Formik
                initialValues={{
                    nickname: '',
                    password: '',
                    submit: null
                }}
                validationSchema={Yup.object().shape({
                    nickname: Yup.string().max(255).required(t('global.nicknameRequiredValidation')),
                    password: Yup.string().max(255).required(t('auth.pwdRequired'))
                })}
                onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                    setSubmitting(true);
                    return API.auth.login(values).then((data) => {
                        setStatus({ success: true });
                        setSubmitting(false);
                        redirectToLocal(redirectToURL);
                    }).catch((err) => {
                        setStatus({ success: false });
                        if(err.code == 'UL001') {
                            setErrors({ submit: t('auth.wrongCredError') });
                        } else if (err.code == 'UL002') {
                            setErrors({ submit: t('auth.accountUnconfirmedError') });
                        } else {
                            setErrors({ submit: t('auth.unexpectedErrorValidation') });
                        }
                        setSubmitting(false);
                    });
                }}
            >
                {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
                    <form noValidate onSubmit={handleSubmit}>
                        <Grid container spacing={3}>
                            <Grid item xs={12}>
                                <Stack spacing={1}>
                                    <InputLabel htmlFor="nickname-login">{t('global.nicknameLabel')}</InputLabel>
                                    <OutlinedInput
                                        id="nickname-login"
                                        type="nickname"
                                        value={values.nickname}
                                        name="nickname"
                                        onBlur={handleBlur}
                                        onChange={handleChange}
                                        placeholder={t('auth.usernameInput')}
                                        fullWidth
                                        error={Boolean(touched.nickname && errors.nickname)}
                                    />
                                    {touched.nickname && errors.nickname && (
                                        <FormHelperText error id="standard-weight-helper-text-nickname-login">
                                            {errors.nickname}
                                        </FormHelperText>
                                    )}
                                </Stack>
                            </Grid>
                            <Grid item xs={12}>
                                <Stack spacing={1}>
                                    <InputLabel htmlFor="password-login">{t('auth.pwd')}</InputLabel>
                                    <OutlinedInput
                                        fullWidth
                                        error={Boolean(touched.password && errors.password)}
                                        id="-password-login"
                                        type={showPassword ? 'text' : 'password'}
                                        value={values.password}
                                        name="password"
                                        onBlur={handleBlur}
                                        onChange={handleChange}
                                        endAdornment={
                                            <InputAdornment position="end">
                                                <IconButton
                                                    aria-label="toggle password visibility"
                                                    onClick={handleClickShowPassword}
                                                    onMouseDown={handleMouseDownPassword}
                                                    edge="end"
                                                    size="large"
                                                >
                                                    {showPassword ? <EyeOutlined /> : <EyeInvisibleOutlined />}
                                                </IconButton>
                                            </InputAdornment>
                                        }
                                        placeholder={t('auth.enterPwd')}
                                    />
                                    {touched.password && errors.password && (
                                        <FormHelperText error id="standard-weight-helper-text-password-login">
                                            {errors.password}
                                        </FormHelperText>
                                    )}
                                </Stack>
                            </Grid>

                            <Grid item xs={12} sx={{ mt: -1 }}>
                                <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                                    {/* <FormControlLabel
                                        control={
                                            <Checkbox
                                                checked={checked}
                                                onChange={(event) => setChecked(event.target.checked)}
                                                name="checked"
                                                color="primary"
                                                size="small"
                                            />
                                        }
                                        label={<Typography variant="h6">Keep me sign in</Typography>}
                                    />*/}
                                    {showResetPassword && <Link variant="h6" component={RouterLink} to="/cosmos-ui/forgot-password" color="primary">
                                        {t('auth.forgotPwd')}
                                    </Link>}
                                    {!showResetPassword &&  <Typography variant="h6">
                                        {t('auth.pwdResetNotAllowed')}
                                    </Typography>}
                                </Stack>
                            </Grid>
                            {errors.submit && (
                                <Grid item xs={12}>
                                    <FormHelperText error>{errors.submit}</FormHelperText>
                                </Grid>
                            )}
                            <Grid item xs={12}>
                                    <LoadingButton
                                        disableElevation
                                        loading={isSubmitting}
                                        fullWidth
                                        size="large"
                                        type="submit"
                                        variant="contained"
                                        color="primary"
                                    >
                                        {t('auth.login')}
                                    </LoadingButton>
                            </Grid>
                            {/* <Grid item xs={12}>
                                <Divider>
                                    <Typography variant="caption"> Login with</Typography>
                                </Divider>
                            </Grid>
                            <Grid item xs={12}>
                                <FirebaseSocial />
                            </Grid> */}
                        </Grid>
                    </form>
                )}
            </Formik>
        </>
    );
};

export default AuthLogin;
