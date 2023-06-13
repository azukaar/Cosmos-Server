import { Link } from 'react-router-dom';

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
                        nickname: Yup.string().max(255).required('Nickname is required'),
                        email: Yup.string().email('Must be a valid email').max(255).required('Email is required'),
                    })}
                    onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                        try {
                            API.users.resetPassword(values).then((data) => {
                                if (data.status == 'error') {
                                    setStatus({ success: false });
                                    setErrors({ submit: 'Unexpected error. Check your infos or try again later.' });
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
                                    label="Nickname"
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
                                        Reset Password
                                    </Button>
                                </Grid>
                            </Grid>
                        </form>
                    )}
                </Formik>}
                {isSuccess && <div>
                    <Typography variant="h6">Check your email for a link to reset your password. If it doesn’t appear within a few minutes, check your spam folder.</Typography>
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
                        Back to login
                    </Button>
                </div>}
            </Grid>
        </Grid>
    </AuthWrapper>
)};

export default ForgotPassword;
