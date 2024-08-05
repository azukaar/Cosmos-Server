import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// material-ui
import { Button, FormHelperText, Grid, InputLabel, OutlinedInput, Stack, Typography } from '@mui/material';

// project import
import AuthWrapper from './AuthWrapper';
import { Formik } from 'formik';

// third-party
import * as Yup from 'yup';
import * as API from '../../api';
import { CosmosInputText } from '../config/users/formShortcuts';
import { useState } from 'react';

// ================================|| LOGIN ||================================ //

const ForgotPassword = () => {
    const { t } = useTranslation();
    const [isSuccess, setIsSuccess] = useState(false);

    return (<AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                    <Typography variant="h3">Password Reset</Typography>
                    {/* <Typography component={Link} to="/register" variant="body1" sx={{ textDecoration: 'none' }} color="primary">
                        Don&apos;t have an account?
                    </Typography> */}
                </Stack>
            </Grid>
            <Grid item xs={12}>
                {!isSuccess && <Formik
                    initialValues={{
                        nickname: '',
                        email: '',
                    }}
                    validationSchema={Yup.object().shape({
                        nickname: Yup.string().max(255).required(t('global.nicknameRequiredValidation')),
                        email: Yup.string().email(t('global.emailInvalidValidation')).max(255).required(t('global.emailRequiredValidation')),
                    })}
                    onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                        try {
                            API.users.resetPassword(values).then((data) => {
                                if (data.status == 'error') {
                                    setStatus({ success: false });
                                    setErrors({ submit: t('auth.unexpectedErrorValidation') });
                                    setSubmitting(false);
                                    return;
                                } else {
                                    setStatus({ success: true });
                                    setSubmitting(false);
                                    setIsSuccess(true);
                                }
                            })
                        } catch (err) {
                            setStatus({ success: false });
                            setErrors({ submit: err.message });
                            setSubmitting(false);
                        }
                    }}
                >
                    {(formik) => (
                        <form noValidate onSubmit={formik.handleSubmit}>
                            <Grid container spacing={3}>

                                <CosmosInputText
                                    name="nickname"
                                    label={t('global.nicknameLabel')}
                                    formik={formik}
                                />

                                <CosmosInputText
                                    name="email"
                                    label="Email"
                                    type="email"
                                    formik={formik}
                                />

                                {formik.errors.submit && (
                                    <Grid item xs={12}>
                                        <FormHelperText error>{formik.errors.submit}</FormHelperText>
                                    </Grid>
                                )}
                                <Grid item xs={12}>
                                    <Button
                                        disableElevation
                                        disabled={formik.isSubmitting}
                                        fullWidth
                                        size="large"
                                        type="submit"
                                        variant="contained"
                                        color="primary"
                                    >
                                        {t('auth.forgotPassword.resetPassword')}
                                    </Button>
                                </Grid>
                            </Grid>
                        </form>
                    )}
                </Formik>}
                {isSuccess && <div>
                    <Typography variant="h6">{t('auth.forgotPassword.checkEmail')}</Typography>
                    <br/><br/>
                    <Button
                        disableElevation
                        fullWidth
                        size="large"
                        type="submit"
                        variant="contained"
                        color="primary"
                        component={Link}
                        to="/cosmos-ui/login"
                    >
                        {t('auth.forgotPassword.backToLogin')}
                    </Button>
                </div>}
            </Grid>
        </Grid>
    </AuthWrapper>
)};

export default ForgotPassword;
