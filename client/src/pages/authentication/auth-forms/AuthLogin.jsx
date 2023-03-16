import React, { useEffect } from 'react';
import { Link as RouterLink } from 'react-router-dom';

// material-ui
import {
    Button,
    Checkbox,
    Divider,
    FormControlLabel,
    FormHelperText,
    Grid,
    Link,
    IconButton,
    InputAdornment,
    InputLabel,
    OutlinedInput,
    Stack,
    Typography,
    Alert
} from '@mui/material';

import * as API from "../../../api";

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project import
import FirebaseSocial from './FirebaseSocial';
import AnimateButton from '../../../components/@extended/AnimateButton';

// assets
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';

// ============================|| FIREBASE - LOGIN ||============================ //

const AuthLogin = () => {
    const [checked, setChecked] = React.useState(false);

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
    const invalid = urlSearchParams.get('invalid') == 1;
    const redirectTo = urlSearchParams.get('redirect') ? urlSearchParams.get('redirect') : '/ui';

    useEffect(() => {
        API.auth.me().then((data) => {
            if(data.status == 'OK') {
                window.location.href = redirectTo;
            }
        });
    });       

    return (
        <>
            { notLogged &&<Grid container spacing={2} justifyContent="center">
                <Alert severity="error">You need to be logged in to access this</Alert>
                <br />
            </Grid>}

            { invalid &&<Grid container spacing={2} justifyContent="center">
                <Alert severity="error">You have been disconnected. Please login to continue</Alert>
                <br />
            </Grid>}
            <Formik
                initialValues={{
                    nickname: '',
                    password: '',
                    submit: null
                }}
                validationSchema={Yup.object().shape({
                    nickname: Yup.string().max(255).required('Nickname is required'),
                    password: Yup.string().max(255).required('Password is required')
                })}
                onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                    try {
                        API.auth.login(values).then((data) => {
                            if(data.status == 'error') {
                                setStatus({ success: false });
                                if(data.code == 'UL001') {
                                    setErrors({ submit: 'Wrong nickname or password. Try again or try resetting your password' });
                                } else if (data.code == 'UL002') {
                                    setErrors({ submit: 'You have not yet registered your account. You should have an invite link in your emails. If you need a new one, contact your administrator.' });
                                } else if(data.status == 'error') {
                                    setErrors({ submit: 'Unexpected error. Try again later.' });
                                }
                                setSubmitting(false);
                                return;
                            } else {
                                setStatus({ success: true });
                                setSubmitting(false);
                                window.location.href = redirectTo;
                            }
                        })
                    } catch (err) {
                        setStatus({ success: false });
                        setErrors({ submit: err.message });
                        setSubmitting(false);
                    }
                }}
            >
                {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
                    <form noValidate onSubmit={handleSubmit}>
                        <Grid container spacing={3}>
                            <Grid item xs={12}>
                                <Stack spacing={1}>
                                    <InputLabel htmlFor="nickname-login">Nickname</InputLabel>
                                    <OutlinedInput
                                        id="nickname-login"
                                        type="nickname"
                                        value={values.nickname}
                                        name="nickname"
                                        onBlur={handleBlur}
                                        onChange={handleChange}
                                        placeholder="Enter your nickname"
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
                                    <InputLabel htmlFor="password-login">Password</InputLabel>
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
                                        placeholder="Enter your password"
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
                                    />
                                    <Link variant="h6" component={RouterLink} to="" color="text.primary">
                                        Forgot Password?
                                    </Link> */}
                                </Stack>
                            </Grid>
                            {errors.submit && (
                                <Grid item xs={12}>
                                    <FormHelperText error>{errors.submit}</FormHelperText>
                                </Grid>
                            )}
                            <Grid item xs={12}>
                                <AnimateButton>
                                    <Button
                                        disableElevation
                                        disabled={isSubmitting}
                                        fullWidth
                                        size="large"
                                        type="submit"
                                        variant="contained"
                                        color="primary"
                                    >
                                        Login
                                    </Button>
                                </AnimateButton>
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
