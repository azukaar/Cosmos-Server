import { Link } from 'react-router-dom';

import * as Yup from 'yup';

// material-ui
import { Alert, Button, CircularProgress, FormControl, FormHelperText, Grid, Stack, Typography } from '@mui/material';

// ant-ui icons
import { CheckCircleOutlined, LeftOutlined, QuestionCircleFilled, QuestionCircleOutlined, RightOutlined } from '@ant-design/icons';

// project import
import AuthWrapper from '../authentication/AuthWrapper';
import { useEffect, useState } from 'react';

import * as API from '../../api';
import { Formik } from 'formik';
import { CosmosInputPassword, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import AnimateButton from '../../components/@extended/AnimateButton';
import { Box } from '@mui/system';
// ================================|| LOGIN ||================================ //

const NewInstall = () => {
    const [activeStep, setActiveStep] = useState(0);
    const [status, setStatus] = useState(null);
    const [counter, setCounter] = useState(0);
    const refreshStatus = async () => {
        try {
            const res = await API.getStatus()
            setStatus(res.data);
        } catch(error) {
            if(error.status == 401)
                window.location.href = "/ui/login";
        }
        if (typeof status !== 'undefined') {
            setTimeout(() => {
                setCounter(counter + 1);
            }, 2000);
        }
    }

    useEffect(() => {
        refreshStatus();
    }, [counter]);
    
    useEffect(() => {
        if(activeStep == 4 && status && !status.database) {
            setActiveStep(5);
        }
    }, [activeStep, status]);

    const steps = [
        {
            label: 'Welcome! 💖',
            component: <div>
                First of all, thanks a lot for trying out Cosmos! And Welcome to the setup wizard.
                This wizard will guide you through the setup of Cosmos. It will take about 2-3 minutes and you will be ready to go.
            </div>,
            nextButtonLabel: () => {
                return 'Start';
            }
        },
        {
            label: 'Docker 🐋 (step 1/4)',
            component: <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> Cosmos is using docker to run applications. It is optionnal, but Cosmos will run in reverse-proxy-only mode if it cannot connect to Docker.
                </div>
                {(status && status.docker) ? 
                    <Alert severity="success">
                        Docker is installed and running.
                    </Alert> :
                    <Alert severity="error">
                        Docker is not connected! Please check your docker connection.<br/>
                        Did you forget to add <pre>-v /var/run/docker.sock:/var/run/docker.sock</pre> to your docker run command?<br />
                        if your docker daemon is running somewhere else, please add <pre>-e DOCKER_HOST=...</pre> to your docker run command.
                    </Alert>
                }
                {(status && status.docker) ? (
                    <div>
                        <center>
                            <CheckCircleOutlined 
                                style={{ fontSize: '30px', color: '#52c41a' }}
                            />
                        </center>
                    </div>
                ) : (<><div>
                    Rechecking Docker Status...
                </div>
                <div>
                    <center><CircularProgress color="inherit" /></center>
                </div></>)}
            </Stack>,
            nextButtonLabel: () => {
                return status && status.docker ? 'Next' : 'Skip';
            }
        },
        {
            label: 'Database 🗄️ (step 2/4)',
            component:  <Stack item xs={12} spacing={2}>
                <div>
                <QuestionCircleOutlined /> Cosmos is using a MongoDB database to store all the data. It is optionnal, but Authentication as well as the UI will not work without a database.
                </div>
                {(status && status.database) ? 
                    <Alert severity="success">
                        Database is connected.
                    </Alert> :
                    <><Alert severity="error">
                        Database is not connected!
                    </Alert>
                    <div>
                    <Formik
                        initialValues={{
                            DBMode: "Create"
                        }}
                        validate={(values) => {
                        }}
                        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                            try {
                                setSubmitting(true);
                                const res = await API.newInstall({
                                    step: "2",
                                    MongoDBMode: values.DBMode,
                                    MongoDB: values.MongoDB,
                                });
                                if(res.status == "OK")
                                    setStatus({ success: true });
                            } catch (error) {
                                setStatus({ success: false });
                                setErrors({ submit: error.message });
                                setSubmitting(false);
                            }
                        }}>
                        {(formik) => (
                            <form noValidate onSubmit={formik.handleSubmit}>
                                <Stack item xs={12} spacing={2}>
                                <CosmosSelect
                                    name="DBMode"
                                    label="Select your choice"
                                    formik={formik}
                                    options={[
                                        ["Create", "Automatically create a secure database (recommended)"],
                                        ["Provided", "Supply my own database credentials"],
                                        ["DisableUserManagement", "Disable User Management and UI"],
                                    ]}
                                />
                                {formik.values.DBMode === "Provided" && (
                                    <>
                                    <CosmosInputText
                                        name="MongoDB"
                                        label="Database URL"
                                        placeholder={"mongodb://user:password@localhost:27017"}
                                        formik={formik}
                                    />
                                    </>
                                )}
                                {formik.errors.submit && (
                                  <Grid item xs={12}>
                                    <FormHelperText error>{formik.errors.submit}</FormHelperText>
                                  </Grid>
                                )}
                                <AnimateButton>
                                    <Button
                                        type="submit"
                                        variant="contained"
                                        color="primary"
                                        disabled={formik.isSubmitting}
                                        fullWidth>
                                        {formik.isSubmitting ? 'Loading' : (
                                            formik.values.DBMode === "DisableUserManagement" ? 'Disable' : 'Connect'
                                        )}
                                    </Button>
                                </AnimateButton>
                                </Stack>
                            </form>
                        )}
                    </Formik>
                    </div>
                    </>
                }
                {(status && status.database) ? (
                    <div>
                        <center>
                            <CheckCircleOutlined 
                                style={{ fontSize: '30px', color: '#52c41a' }}
                            />
                        </center>
                    </div>
                ) : (<><div>
                    Rechecking Database Status...
                </div>
                <div>
                    <center><CircularProgress color="inherit" /></center>
                </div></>)}
            </Stack>,
            nextButtonLabel: () => {
                return (status && status.database) ? 'Next' : '';
            }
        },
        {
            label: 'HTTPS 🌐 (step 3/4)',
            component: (<Stack item xs={12} spacing={2}>
            <div>
                <QuestionCircleOutlined /> It is recommended to use Let's Encrypt to automatically provide HTTPS Certificates.
                This requires a valid domain name pointing to this server. If you don't have one, you can use a self-signed certificate. 
                If you enable HTTPS, it will be effective after the next restart.
            </div>
            <div>
                {status && <div>
                    HTTPS Certificate Mode is currently: <b>{status.HTTPSCertificateMode}</b>
                </div>}
            </div>
            <div>
            <Formik
                initialValues={{
                    HTTPSCertificateMode: "LETSENCRYPT"
                }}
                onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                    try {
                        setSubmitting(true);
                        const res = await API.newInstall({
                            step: "3",
                            HTTPSCertificateMode: values.HTTPSCertificateMode,
                            SSLEmail: values.SSLEmail,
                            TLSKey: values.HTTPSCertificateMode === "PROVIDED" ? values.TLSKey : '',
                            TLSCert: values.HTTPSCertificateMode === "PROVIDED" ? values.TLSCert : '',
                            Hostname: values.Hostname,
                        });
                        if(res.status == "OK")
                            setStatus({ success: true });
                    } catch (error) {
                        setStatus({ success: false });
                        setErrors({ submit: error.message });
                        setSubmitting(false);
                    }
                }}>
                {(formik) => (
                    <form noValidate onSubmit={formik.handleSubmit}>
                        <Stack item xs={12} spacing={2}>
                        <CosmosSelect
                            name="HTTPSCertificateMode"
                            label="Select your choice"
                            formik={formik}
                            options={[
                                ["LETSENCRYPT", "Use Let's Encrypt automatic HTTPS (recommended)"],
                                ["PROVIDED", "Supply my own HTTPS certificate"],
                                ["SELFSIGNED", "Generate a self-signed certificate"],
                                ["DISABLE", "Use HTTP only (not recommended)"],
                            ]}
                        />
                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && (
                            <>
                            <CosmosInputText
                                name="SSLEmail"
                                label="Let's Encrypt Email"
                                placeholder={"email@domain.com"}
                                formik={formik}
                            />
                            </>
                        )}
                        {formik.values.HTTPSCertificateMode === "PROVIDED" && (
                            <>
                                <CosmosInputText
                                    multiline
                                    name="TLSKey"
                                    label="Private Certificate"
                                    placeholder="-----BEGIN CERTIFICATE-----\nMIIEowIBwIBAA...."
                                    formik={formik}
                                />
                                <CosmosInputText
                                    multiline
                                    name="TLSCert"
                                    label="Public Certificate"
                                    placeholder="-----BEGIN RSA PRIVATE KEY-----\nQCdYIUkYi...."
                                    formik={formik}
                                />
                            </>
                        )}
                        
                        <CosmosInputText
                            name="Hostname"
                            label="Hostname (Domain required for Let's Encrypt)"
                            placeholder="yourdomain.com, your ip, or localhost"
                            formik={formik}
                        />

                        {formik.errors.submit && (
                          <Grid item xs={12}>
                            <FormHelperText error>{formik.errors.submit}</FormHelperText>
                          </Grid>
                        )}
                        <AnimateButton>
                            <Button
                                type="submit"
                                variant="contained"
                                color="primary"
                                disabled={formik.isSubmitting}
                                fullWidth>
                                {formik.isSubmitting ? 'Loading' : (
                                    formik.values.HTTPSCertificateMode === "DISABLE" ? 'Disable' : 'Update'
                                )}
                            </Button>
                        </AnimateButton>
                        </Stack>
                    </form>
                )}
            </Formik>
            </div>
            </Stack>),
            nextButtonLabel: () => {
                return status ? 'Next' : 'Skip';
            }
        },
        {
            label: 'Admin Account 🔑 (step 4/4)',
            component: <div>
                <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> Create a local admin account to manage your server. Email is optional and used for notifications and password recovery.
                </div>
                <Formik
                    initialValues={{
                        nickname: '',
                        password: '',
                        confirmPassword: '',
                        email: '',
                    }}
                    validationSchema={Yup.object().shape({
                        nickname: Yup.string().required('Nickname is required').min(3).max(32),
                        password: Yup.string().required('Password is required').min(8).max(128).matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%^&*])(?=.{8,})/, 'Password must contain at least 1 lowercase, 1 uppercase, 1 number, and 1 special character'),
                        email: Yup.string().email('Must be a valid email').max(255),
                        confirmPassword: Yup.string().oneOf([Yup.ref('password'), null], 'Passwords must match'),
                    })}
                    onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                        try {
                            setSubmitting(true);
                            const res = await API.newInstall({
                                step: "4",
                                nickname: values.nickname,
                                password: values.password,
                                email: values.email,
                            });
                            if(res.status == "OK") {
                                setStatus({ success: true });
                                setActiveStep(5);
                            }
                        } catch (error) {
                            setStatus({ success: false });
                            setErrors({ submit: error.message });
                            setSubmitting(false);
                        }
                    }}>
                    {(formik) => (
                        <form noValidate onSubmit={formik.handleSubmit}>
                            <Stack item xs={12} spacing={2}>
                            <CosmosInputText
                                name="nickname"
                                label="Nickname"
                                placeholder="admin"
                                formik={formik}
                            />
                            <CosmosInputText
                                name="email"
                                label="Email"
                                placeholder="Email (optional)"
                                formik={formik}
                                type="email"
                            />
                            <CosmosInputPassword
                                name="password"
                                label="Password"
                                placeholder="password"
                                formik={formik}
                                type="password"
                            />
                            <CosmosInputText
                                name="confirmPassword"
                                label="Confirm Password"
                                placeholder="password"
                                formik={formik}
                                type="password"
                            />
                            {formik.errors.submit && (
                                <Grid item xs={12}>
                                    <FormHelperText error>{formik.errors.submit}</FormHelperText>
                                </Grid>
                            )}
                            <AnimateButton>
                                <Button
                                    type="submit"
                                    variant="contained"
                                    color="primary"
                                    disabled={formik.isSubmitting}
                                    fullWidth>
                                    {formik.isSubmitting ? 'Loading' : 'Create'}
                                </Button>
                            </AnimateButton>
                            </Stack>
                        </form>
                    )}
                </Formik>
                </Stack>
            </div>,
            nextButtonLabel: () => {
                return '';
            }
        },
        {
            label: 'Finish 🎉',
            component: <div>
                Well done! You have successfully installed Cosmos. You can now login to your server using the admin account you created.
                If you have changed the hostname, don't forget to use that URL to access your server after the restart.
                If you have are running into issues, check the logs for any error messages and edit the file in the /config folder. 
                If you still don't manage, please join our <a href="https://discord.gg/PwMWwsrwHA">Discord server</a> and we'll be happy to help!
            </div>,
            nextButtonLabel: () => {
                return 'Apply and Restart';
            }
        }
    ];

    return <AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                    <Typography variant="h3">{steps[activeStep].label}</Typography>
                </Stack>
            </Grid>
            <Grid item xs={12} spacing={2}>
                {steps[activeStep].component}

                {/*JSON.stringify(status)*/}
                
                <br />
                <Stack direction="row" spacing={2} sx={{ '& > *': { flexGrow: 1 } }}>
                    <Button 
                        variant="contained"
                        startIcon={<LeftOutlined />}
                        onClick={() => setActiveStep(activeStep - 1)} 
                        disabled={activeStep <= 0}
                    >Back</Button>

                    <Button 
                        variant="contained"
                        endIcon={<RightOutlined />}
                        disabled={steps[activeStep].nextButtonLabel() == ''}
                        onClick={() => {
                            if(activeStep >= steps.length - 1) {
                                API.newInstall({
                                    step: "5",
                                })
                                setTimeout(() => {
                                    window.location.href = "/ui/login";
                                }, 500);
                            } else 
                                setActiveStep(activeStep + 1)
                        }}
                    >{
                        steps[activeStep].nextButtonLabel() ? 
                            steps[activeStep].nextButtonLabel() :
                            'Next'
                    }</Button>
                </Stack>
            </Grid>
        </Grid>
    </AuthWrapper>
};

export default NewInstall;
