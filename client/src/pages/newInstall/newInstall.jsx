import { Link } from 'react-router-dom';

import * as Yup from 'yup';

// material-ui
import { Alert, Button, Checkbox, CircularProgress, FormControl, FormHelperText, Grid, Stack, Typography } from '@mui/material';

// ant-ui icons
import { CheckCircleOutlined, LeftOutlined, QuestionCircleFilled, QuestionCircleOutlined, RightOutlined } from '@ant-design/icons';

// project import
import AuthWrapper from '../authentication/AuthWrapper';
import { useEffect, useState } from 'react';

import * as API from '../../api';
import { Formik } from 'formik';
import LogsInModal from '../../components/logsInModal';
import { CosmosCheckbox, CosmosInputPassword, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import AnimateButton from '../../components/@extended/AnimateButton';
import { Box } from '@mui/system';
import { pull } from 'lodash';
import { isDomain } from '../../utils/indexs';
import { DnsChallengeComp } from '../../utils/dns-challenge-comp';
// ================================|| LOGIN ||================================ //

const debounce = (func, wait) => {
    let timeout;
    return function (...args) {
      const context = this;
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(context, args), wait);
    };
  };
  
  const checkHost = debounce((host, setHostError, setHostIp) => {
    if(isDomain(host)) {
      API.getDNS(host).then((data) => {
        setHostError(null)
        setHostIp(data.data)
      }).catch((err) => {
        setHostError(err.message)
        setHostIp(null)
      });
    } else {
        setHostError(null);
        setHostIp(null);
    }
  }, 500)

const hostnameIsDomainReg = /^((?!localhost|\d+\.\d+\.\d+\.\d+)[a-zA-Z0-9\-]{1,63}\.)+[a-zA-Z]{2,63}$/

const NewInstall = () => {
    const [activeStep, setActiveStep] = useState(0);
    const [status, setStatus] = useState(null);
    const [counter, setCounter] = useState(0);
    let [hostname, setHostname] = useState('');
    const [databaseEnable, setDatabaseEnable] = useState(true);
    const [pullRequest, setPullRequest] = useState(null);
    const [pullRequestOnSuccess, setPullRequestOnSuccess] = useState(null);
    const [hostError, setHostError] = useState(null);
    const [hostIp, setHostIp] = useState(null);
    const [cleanInstall, setCleanInstall] = useState(true);

    const refreshStatus = async () => {
        try {
            const res = await API.getStatus()
            setStatus(res.data);
        } catch(error) {
            if(error.status == 401)
                window.location.href = "/cosmos-ui/login";
        }
        if (typeof status !== 'undefined') {
            setTimeout(() => {
                setCounter(counter + 1);
            }, 2500);
        }
    }

    useEffect(() => {
        refreshStatus();
    }, [counter]);
    
    useEffect(() => {
        if(activeStep == 4 && status && !databaseEnable) {
            setActiveStep(5);
        }
    }, [activeStep, status]);

    const getHTTPSOptions = (hostname) => {
        if(!hostname) {
            return [["", "Set your hostname first"]];
        }

        if(hostname.match(hostnameIsDomainReg)) {
            return [
                ["", "Select an option"],
                ["LETSENCRYPT", "Use Let's Encrypt automatic HTTPS (recommended)"],
                ["SELFSIGNED", "Generate self-signed certificate"],
                ["PROVIDED", "Supply my own HTTPS certificate"],
                ["DISABLED", "Use HTTP only (not recommended)"],
            ]
        } else {
            return [
                ["", "Select an option"],
                ["SELFSIGNED", "Generate self-signed certificate (recommended)"],
                ["PROVIDED", "Supply my own HTTPS certificate"],
                ["DISABLED", "Use HTTP only (not recommended)"],
            ] 
        }
    }

    const steps = [
        {
            label: 'Welcome! 💖',
            component: <div>
                First of all, thanks a lot for trying out Cosmos! And Welcome to the setup wizard.
                This wizard will guide you through the setup of Cosmos. It will take about 2-3 minutes and you will be ready to go.
                <br /><br />
                <Checkbox checked={cleanInstall} onChange={(e) => setCleanInstall(e.target.checked)} />Clean install (remove any existing config files)
                <br /><br />
                <a style={{color: 'white', textDecoration: 'none'}} target='_blank' rel="noopener noreferrer" href="https://cosmos-cloud.io/doc/2%20setup">
                    <Button variant="outlined" color="inherit" startIcon={<QuestionCircleOutlined />}>
                     Link to documentation
                    </Button>
                </a>
            </div>,
            nextButtonLabel: () => {
                return 'Start';
            }
        },
        {
            label: 'Docker 🐋 (step 1/4)',
            component: <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> Cosmos is using docker to run applications. It is optional, but Cosmos will run in reverse-proxy-only mode if it cannot connect to Docker.
                </div>
                {status && (status.docker ? 
                <Alert severity="success">
                    Docker is installed and running.
                </Alert> :
                <Alert severity="error">
                    Docker is not connected! Please check your docker connection.<br/>
                    Did you forget to add <pre>-v /var/run/docker.sock:/var/run/docker.sock</pre> to your docker run command?<br />
                    if your docker daemon is running somewhere else, please add <pre>-e DOCKER_HOST=...</pre> to your docker run command.
                </Alert>)}
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
                <QuestionCircleOutlined /> Cosmos is using a MongoDB database to store all the data. It is optional, but Authentication as well as the UI will not work without a database.
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
                            setSubmitting(true);
                            
                            setPullRequest(() => ((cb) => {
                                API.newInstall({
                                    step: "2",
                                    MongoDBMode: values.DBMode,
                                    MongoDB: values.MongoDB,
                                }, cb)
                            }));

                            return new Promise(() => {});
                        }}>
                        {(formik) => (
                            <form noValidate onSubmit={formik.handleSubmit}>
                                <LogsInModal
                                    request={pullRequest}
                                    title="Installing Database..."
                                    OnSuccess={() => {
                                        if(formik.values.DBMode === "DisableUserManagement") {
                                            setDatabaseEnable(false);
                                        }
                                        pullRequestOnSuccess();
                                        API.getStatus().then((res) => {
                                            formik.setSubmitting(false);
                                            formik.setStatus({ success: true });
                                        });
                                    }}
                                    OnError={(error) => {
                                        formik.setStatus({ success: false });
                                        formik.setErrors({ submit: error.message });
                                        formik.setSubmitting(false);
                                        pullRequestOnSuccess();
                                    }}
                                />
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
                ) : (<>
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
                This requires a valid domain name pointing to this server. If you don't have one, <strong>you can select "Generate self-signed certificate" in the dropdown. </strong> 
                If you enable HTTPS, it will be effective after the next restart.
            </div>
            <div>
            <Formik
                initialValues={{
                    HTTPSCertificateMode: "",
                    UseWildcardCertificate: false,
                    DNSChallengeProvider: '',
                    DNSChallengeConfig: {},
                    __success: false,
                }}
                validationSchema={Yup.object().shape({
                        SSLEmail: Yup.string().when('HTTPSCertificateMode', {
                            is: "LETSENCRYPT",
                            then: Yup.string().email().required(),
                            otherwise: Yup.string().email(),
                        }),
                        TLSKey: Yup.string().when('HTTPSCertificateMode', {
                            is: "PROVIDED",
                            then: Yup.string().required(),
                        }),
                        TLSCert: Yup.string().when('HTTPSCertificateMode', {
                            is: "PROVIDED",
                            then: Yup.string().required(),
                        }),
                        Hostname: Yup.string().when('HTTPSCertificateMode', {
                            is: "LETSENCRYPT",
                            then: Yup.string().required().matches(hostnameIsDomainReg, 'Let\'s Encrypt only accepts domain names'),
                            otherwise: Yup.string().required()
                        }),
                })}
                onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                    try {
                        setSubmitting(true);
                        const res = await API.newInstall({
                            step: "3",
                            HTTPSCertificateMode: values.HTTPSCertificateMode,
                            SSLEmail: values.SSLEmail,
                            UseWildcardCertificate: values.UseWildcardCertificate,
                            TLSKey: values.HTTPSCertificateMode === "PROVIDED" ? values.TLSKey : '',
                            TLSCert: values.HTTPSCertificateMode === "PROVIDED" ? values.TLSCert : '',
                            Hostname: values.Hostname,
                            DNSChallengeProvider: values.DNSChallengeProvider,
                            DNSChallengeConfig: values.DNSChallengeConfig,
                        });
                        if(res.status == "OK") {
                            setStatus({ success: true });
                            setHostname((values.HTTPSCertificateMode == "DISABLED" ? "http://" : "https://") + values.Hostname);
                            setActiveStep(4);
                        }
                        return res;
                    } catch (error) {
                        setStatus({ success: false });
                        setErrors({ submit: "Please check you have filled all the inputs properly" });
                        setSubmitting(false);
                    }
                }}>
                {(formik) => (
                    <form noValidate onSubmit={formik.handleSubmit}>
                        <Stack item xs={12} spacing={2}>
                        <CosmosInputText
                            name="Hostname"
                            label="Hostname (Domain required for Let's Encrypt)"
                            placeholder="yourdomain.com, your ip, or localhost"
                            formik={formik}
                            onChange={(e) => {
                              checkHost(e.target.value, setHostError, setHostIp);
                            }}
                        />
                        {formik.values.Hostname && (formik.values.Hostname.match(hostnameIsDomainReg) ? 
                            <Alert severity="info">
                                You seem to be using a domain name. <br />
                                Let's Encrypt can automatically generate a certificate for you.
                            </Alert>
                            :
                            <Alert severity="info">
                                You seem to be using an IP address or local domain. <br />
                                You can use automatic Self-Signed certificates.
                            </Alert>)
                        }
                        <CosmosSelect
                            name="HTTPSCertificateMode"
                            label="Select your choice"
                            formik={formik}
                            options={getHTTPSOptions(formik.values.Hostname && formik.values.Hostname)}
                        />
                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && (
                            <>
                            <Alert severity="warning">
                                If you are using Cloudflare, make sure the DNS record is <strong>NOT</strong> set to <b>Proxied</b> (you should not see the orange cloud but a grey one).
                                Otherwise Cloudflare will not allow Let's Encrypt to verify your domain. <br />
                                Alternatively, you can also use the DNS challenge.
                            </Alert>
                            <CosmosInputText
                                name="SSLEmail"
                                label="Let's Encrypt Email"
                                placeholder={"email@domain.com"}
                                formik={formik}
                            />
                            {formik.values.DNSChallengeProvider && formik.values.DNSChallengeProvider != '' && (
                                <Alert severity="info">
                                    You have enabled the DNS challenge. Make sure you have set the environment variables for your DNS provider.
                                    You can enable it now, but make sure you have set up your API tokens accordingly before attempting to access 
                                    Cosmos after this installer. See doc here: <a target="_blank" rel="noopener noreferrer" href="https://go-acme.github.io/lego/dns/">https://go-acme.github.io/lego/dns/</a>
                                </Alert>
                            )}
                            <DnsChallengeComp 
                                label="Pick a DNS provider (if you are using a DNS Challenge, otherwise leave empty)"
                                name="DNSChallengeProvider"
                                configName="DNSChallengeConfig"
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
                                    placeholder="-----BEGIN RSA PRIVATE KEY-----\nQCdYIUkYi...."
                                    formik={formik}
                                    />
                                <CosmosInputText
                                    multiline
                                    name="TLSCert"
                                    label="Public Certificate"
                                    placeholder="-----BEGIN CERTIFICATE-----\nMIIEowIBwIBAA...."
                                    formik={formik}
                                />
                            </>
                        )}
                        
                        {hostError && <Grid item xs={12}>
                          <Alert color='error'>{hostError}</Alert>
                        </Grid>}
                        {hostIp && <Grid item xs={12}>
                            <Alert color='info'>This hostname is pointing to <strong>{hostIp}</strong>, check that it is your server IP!</Alert>
                        </Grid>}

                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && formik.values.UseWildcardCertificate && (!formik.values.DNSChallengeProvider || formik.values.DNSChallengeProvider == '') && (
                            <Alert severity="error">
                                You have enabled wildcard certificates with Let's Encrypt. This only works if you use the DNS challenge!
                                Please edit the DNS Provider text input.
                            </Alert>
                        )}
                        
                        {(formik.values.HTTPSCertificateMode === "LETSENCRYPT" || formik.values.HTTPSCertificateMode === "SELFSIGNED") && formik.values.Hostname && formik.values.Hostname.match(hostnameIsDomainReg) && (
                        <CosmosCheckbox
                            label={"Use Wildcard Certificate for *." + (formik.values.Hostname ||  "")}
                            name="UseWildcardCertificate"
                            formik={formik}
                        />)}

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
                                disabled={formik.isSubmitting || !formik.isValid}
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
                return (status && status.hostname != '0.0.0.0') ? 'Next' : '';
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
                        // nickname cant be admin or root
                        nickname: Yup.string().required('Nickname is required').min(3).max(32)
                        .matches(/^(?!admin|root).*$/, 'Nickname cannot be admin or root'),
                        password: Yup.string().required('Password is required').min(8).max(128).matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[~!@#$%\^&\*\(\)_\+=\-\{\[\}\]:;"'<,>\.\?\/])(?=.{9,})/, 'Password must contain 9 characters: at least 1 lowercase, 1 uppercase, 1 number, and 1 special character'),
                        email: Yup.string().email('Must be a valid email').max(255),
                        confirmPassword: Yup.string().oneOf([Yup.ref('password'), null], 'Passwords must match'),
                    })}
                    onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
                        setSubmitting(true);
                        return await API.newInstall({
                            step: "4",
                            nickname: values.nickname,
                            password: values.password,
                            email: values.email,
                        }).then((res) => {
                            setStatus({ success: true });
                            setActiveStep(5);
                        }).catch((error) => {
                            setStatus({ success: false });
                            setErrors({ submit: error.message });
                            setSubmitting(false);
                        });
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
                                    disabled={formik.isSubmitting || !formik.isValid}
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
                If you still don't manage, please join our <a target="_blank" rel="noopener noreferrer" href="https://discord.gg/PwMWwsrwHA">Discord server</a> and we'll be happy to help!
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
                        onClick={() => {
                            if(activeStep == 5 && !databaseEnable) {
                                setActiveStep(activeStep - 2)
                            }
                            setActiveStep(activeStep - 1)
                        }} 
                        disabled={activeStep <= 0}
                    >Back</Button>

                    <Button 
                        variant="contained"
                        endIcon={<RightOutlined />}
                        disabled={steps[activeStep].nextButtonLabel() == ''}
                        onClick={() => {
                            if(activeStep == 0 && cleanInstall) {
                                API.newInstall({
                                    step: "-1",
                                }).then((res) => {
                                    refreshStatus()
                                });
                            }

                            if(activeStep >= steps.length - 1) {
                                API.newInstall({
                                    step: "5",
                                })
                                setTimeout(() => {
                                    window.location.href = hostname + "/cosmos-ui/login";
                                }, 500);
                            } else {
                                setActiveStep(activeStep + 1)
                            }
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
