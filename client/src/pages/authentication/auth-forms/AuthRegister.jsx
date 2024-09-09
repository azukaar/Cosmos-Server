import { useEffect, useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// material-ui
import {
    Box,
    Button,
    Divider,
    FormControl,
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

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

import * as API from '../../../api';

// project import
import FirebaseSocial from './FirebaseSocial';
import AnimateButton from '../../../components/@extended/AnimateButton';
import { strengthColor, strengthIndicator } from '../../../utils/password-strength';

// assets
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import { LoadingButton } from '@mui/lab';
import { redirectToLocal } from '../../../utils/indexs';

// ============================|| FIREBASE - REGISTER ||============================ //

const AuthRegister = ({nickname, isRegister, isInviteLink, regkey}) => {
    const { t } = useTranslation();
    const [level, setLevel] = useState();
    const [showPassword, setShowPassword] = useState(false);
    const handleClickShowPassword = () => {
        setShowPassword(!showPassword);
    };

    const handleMouseDownPassword = (event) => {
        event.preventDefault();
    };

    const changePassword = (value) => {
        const temp = strengthIndicator(value);
        setLevel(strengthColor(temp, t));
    };

    useEffect(() => {
        changePassword('');
    }, []);

    return (
        <>
            <Formik
                initialValues={{
                    nickname: nickname,
                    lastname: '',
                    email: '',
                    company: '',
                    password: '',
                    submit: null
                }}
                validationSchema={Yup.object().shape({
                    nickname: Yup.string().max(255).required('Nickname is required'),
                    password: Yup.string()
                        .max(255)
                        .required('Password is required')
                        .matches(
                            /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[~!@#$%\^&\*\(\)_\+=\-\{\[\}\]:;"'<,>\/])(?=.{9,})/,
                            'Must Contain 9 Characters, One Uppercase, One Lowercase, One Number and one special case Character (~!@#$%^&*()_+=-{[}]:;"\'<>.?/)'
                        ),
                })}
                onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                    return API.users.register({
                        nickname: nickname,
                        registerKey: regkey,
                        password: values.password,
                    }).then((res) => {
                        setStatus({ success: true });
                        setSubmitting(false);
                        redirectToLocal('/cosmos-ui/login');
                    }).catch((err) => {
                        setStatus({ success: false });
                        setErrors({ submit: err.message });
                        setSubmitting(false);
                    });
                }}
            >
                {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
                    <form noValidate onSubmit={handleSubmit}>
                        <Grid container spacing={3}>
                            {isInviteLink ? <Grid item xs={12}>
                                <Alert severity="info">
                                    <strong>Invite Link</strong> - You have been invited to join this Cosmos instance. This Nickname has been provided to us by your administrator. Keep note of it, you will need it to login. 
                                </Alert>
                            </Grid> : ''}
                            {isInviteLink ? <Grid item xs={12}>
                                <Stack spacing={1}>
                                    <InputLabel htmlFor="nickname-signup">Nickname</InputLabel>
                                    <OutlinedInput
                                        id="nickname-login"
                                        type="nickname"
                                        value={nickname}
                                        name="nickname"
                                        onBlur={handleBlur}
                                        onChange={handleChange}
                                        placeholder=""
                                        disabled={true}
                                        fullWidth
                                        error={Boolean(touched.nickname && errors.nickname)}
                                    />
                                    {touched.nickname && errors.nickname && (
                                        <FormHelperText error id="helper-text-nickname-signup">
                                            {errors.nickname}
                                        </FormHelperText>
                                    )}
                                </Stack>
                            </Grid> : ''}
                            <Grid item xs={12}>
                                <Stack spacing={1}>
                                    <InputLabel htmlFor="password-signup">Password</InputLabel>
                                    <OutlinedInput
                                        fullWidth
                                        error={Boolean(touched.password && errors.password)}
                                        id="password-signup"
                                        type={showPassword ? 'text' : 'password'}
                                        value={values.password}
                                        name="password"
                                        onBlur={handleBlur}
                                        onChange={(e) => {
                                            handleChange(e);
                                            changePassword(e.target.value);
                                        }}
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
                                        placeholder="********"
                                        inputProps={{}}
                                    />
                                    {touched.password && errors.password && (
                                        <FormHelperText error id="helper-text-password-signup">
                                            {errors.password}
                                        </FormHelperText>
                                    )}
                                </Stack>
                                <FormControl fullWidth sx={{ mt: 2 }}>
                                    <Grid container spacing={2} alignItems="center">
                                        <Grid item>
                                            <Box sx={{ bgcolor: level?.color, width: 85, height: 8, borderRadius: '7px' }} />
                                        </Grid>
                                        <Grid item>
                                            <Typography variant="subtitle1" fontSize="0.75rem">
                                                {level?.label}
                                            </Typography>
                                        </Grid>
                                    </Grid>
                                </FormControl>
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
                                        {
                                            isRegister ? 'Register' : 'Reset Password'
                                        }
                                    </LoadingButton>
                            </Grid>
                        </Grid>
                    </form>
                )}
            </Formik>
        </>
    );
};

export default AuthRegister;
