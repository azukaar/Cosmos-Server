import { Link } from 'react-router-dom';
import * as Yup from 'yup';
import { Trans, useTranslation } from 'react-i18next';

// material-ui
import { Alert, Button, Checkbox, CircularProgress, FormControl, FormHelperText, Grid, Stack, Tooltip, Typography } from '@mui/material';

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
import { isDomain, redirectTo, redirectToLocal } from '../../utils/indexs';
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
    const { t } = useTranslation();
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
            redirectToLocal("/cosmos-ui/login");
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
            return [["", t('SetHostname')]];
        }

        if(hostname.match(hostnameIsDomainReg)) {
            return [
                ["", "Select an option"],
                ["LETSENCRYPT", t('SSLLetsEncrypt')],
                ["SELFSIGNED", t('SSLSelfSigned')],
                ["PROVIDED", t('SSLProvided')],
                ["DISABLED", t('SSLDisabled')],
            ]
        } else {
            return [
                ["", t('SelectOption')],
                ["SELFSIGNED", t('SSLLetsEncrypt')],
                ["PROVIDED", t('SSLProvided')],
                ["DISABLED", t('SSLDisabled')],
            ]
        }
    }

    const steps = [
        {
            label: t('Welcome'),
            component: <div>
                {t('WelcomeText')}
                <br /><br />
                <Checkbox checked={cleanInstall} onChange={(e) => setCleanInstall(e.target.checked)} />{t('CleanInstall')}
                <br /><br />
                <a style={{color: 'white', textDecoration: 'none'}} target='_blank' rel="noopener noreferrer" href="https://cosmos-cloud.io/doc/2%20setup">
                    <Button variant="outlined" color="inherit" startIcon={<QuestionCircleOutlined />}>
                     {t('LinkDocs')}
                    </Button>
                </a>
            </div>,
            nextButtonLabel: () => {
                return 'Start';
            }
        },
        {
            label: t('Docker'),
            component: <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> {t('WhatIsCosmos')}
                </div>
                {status && (status.docker ? 
                <Alert severity="success">
                    {t('DockerAvailable')}
                </Alert> :
                <Alert severity="error"><Trans i18nKey="DockerNotConnected">
                    Docker is not connected! Please check your docker connection.<br/>
                    Did you forget to add <pre>-v /var/run/docker.sock:/var/run/docker.sock</pre> to your docker run command?<br />
                    if your docker daemon is running somewhere else, please add <pre>-e DOCKER_HOST=...</pre> to your docker run command.
                </Trans></Alert>)}
                {(status && status.docker) ? (
                    <div>
                        <center>
                            <CheckCircleOutlined 
                                style={{ fontSize: '30px', color: '#52c41a' }}
                            />
                        </center>
                    </div>
                ) : (<><div>
                    {t('CheckingDocker')}
                </div>
                <div>
                    <center><CircularProgress color="inherit" /></center>
                </div></>)}
            </Stack>,
            nextButtonLabel: () => {
                return status && status.docker ? t('Next') : t('Skip');
            }
        },
        {
            label: t('Database'),
            component:  <Stack item xs={12} spacing={2}>
                <div>
                <QuestionCircleOutlined /> {t('CosmosDatabaseText')}
                </div>
                {(status && status.database) ? 
                    <Alert severity="success">
                        {t('DatabaseConnected')}
                    </Alert> :
                    <><Alert severity="error">
                        {t('DatabaseNotConnected')}
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
                                return API.newInstall({
                                    step: "2",
                                    MongoDBMode: values.DBMode,
                                    MongoDB: values.MongoDB,
                                }, cb)
                            }));

                            return new Promise(() => {});
                        }}>
                        {(formik) => (
                            <form noValidate onSubmit={formik.handleSubmit}>
                                {pullRequest && <LogsInModal
                                    request={pullRequest}
                                    title={t('InstallDB')}
                                    OnSuccess={() => {
                                        if(formik.values.DBMode === "DisableUserManagement") {
                                            setDatabaseEnable(false);
                                        }
                                        pullRequestOnSuccess();
                                        return API.getStatus().then((res) => {
                                            formik.setSubmitting(false);
                                            formik.setStatus({ success: true });
                                        });
                                    }}
                                    OnError={(error) => {
                                        formik.setStatus({ success: false });
                                        formik.setErrors({ submit: error.message });
                                        formik.setSubmitting(false);
                                        console.error(error)
                                        pullRequestOnSuccess();
                                    }}
                                    OnClose={() => {
                                        setPullRequest(null);
                                    }}
                                />}
                                <Stack item xs={12} spacing={2}>
                                <CosmosSelect
                                    name="DBMode"
                                    label={t('MakeSelection')}
                                    formik={formik}
                                    options={[
                                        ["Create", t('CreateDB')],
                                        ["Provided", t('ProvidedDB')],
                                        ["DisableUserManagement", t('NoDB')],
                                    ]}
                                />
                                {formik.values.DBMode === "Provided" && (
                                    <>
                                    <CosmosInputText
                                        name="MongoDB"
                                        label={t('DatabaseURL')}
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
                                        {formik.isSubmitting ? t('Loading') : (
                                            formik.values.DBMode === "DisableUserManagement" ? t('Disable') : t('Connect')
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
                return (status && status.database) ? t('Next') : '';
            }
        },
        {
            label: t('HTTPS'),
            component: (<Stack item xs={12} spacing={2}>
            <div>
                <QuestionCircleOutlined /> {t('HTTPSDescription')}
            </div>
            <div>
            <Formik
                initialValues={{
                    HTTPSCertificateMode: "",
                    UseWildcardCertificate: false,
                    DNSChallengeProvider: '',
                    DNSChallengeConfig: {},
                    allowHTTPLocalIPAccess: false,
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
                            then: Yup.string().required().matches(hostnameIsDomainReg, t('LetsEncryptOnlyAcceptsDomains')),
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
                            allowHTTPLocalIPAccess: values.allowHTTPLocalIPAccess,
                        });
                        if(res.status == "OK") {
                            setStatus({ success: true });
                            setHostname((values.HTTPSCertificateMode == "DISABLED" ? "http://" : "https://") + values.Hostname);
                            setActiveStep(4);
                        }
                        return res;
                    } catch (error) {
                        setStatus({ success: false });
                        setErrors({ submit: t('CheckInput') });
                        setSubmitting(false);
                    }
                }}>
                {(formik) => (
                    <form noValidate onSubmit={formik.handleSubmit}>
                        <Stack item xs={12} spacing={2}>
                        <CosmosInputText
                            name="Hostname"
                            label={t('HostnameText')}
                            placeholder={t('HostnamePlaceholder')}
                            formik={formik}
                            onChange={(e) => {
                              checkHost(e.target.value, setHostError, setHostIp);
                            }}
                        />
                        {formik.values.Hostname && (formik.values.Hostname.match(hostnameIsDomainReg) ? 
                            <Alert severity="info">
                                {t('SeemsDomain')} <br />
                                {t('LetsEncryptAuto')}
                            </Alert>
                            :
                            <Alert severity="info">
                                {t('SeemsLocal')} <br />
                                {t('SelfSignedAuto')}
                            </Alert>)
                        }
                        <CosmosSelect
                            name="HTTPSCertificateMode"
                            label={t('SelectOption')}
                            formik={formik}
                            options={getHTTPSOptions(formik.values.Hostname && formik.values.Hostname)}
                        />
                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && (
                            <>
                            <Alert severity="warning"><Trans i18nKey="IfCloudFlareText">
                                If you are using Cloudflare, make sure the DNS record is <strong>NOT</strong> set to <b>Proxied</b> (you should not see the orange cloud but a grey one).
                                Otherwise Cloudflare will not allow Let's Encrypt to verify your domain. <br />
                                Alternatively, you can also use the DNS challenge.
                            </Trans></Alert>
                            <CosmosInputText
                                name="SSLEmail"
                                label={t('SSLEmail')}
                                placeholder={"email@domain.com"}
                                formik={formik}
                            />
                            {formik.values.DNSChallengeProvider && formik.values.DNSChallengeProvider != '' && (
                                <Alert severity="info"><Trans i18nKey="IfDNSChallengeText">
                                    You have enabled the DNS challenge. Make sure you have set the environment variables for your DNS provider.
                                    You can enable it now, but make sure you have set up your API tokens accordingly before attempting to access 
                                    Cosmos after this installer. See doc here: <a target="_blank" rel="noopener noreferrer" href="https://go-acme.github.io/lego/dns/">https://go-acme.github.io/lego/dns/</a>
                                </Trans></Alert>
                            )}
                            <DnsChallengeComp 
                                label={t('DNSChallengeProvider')}
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
                                    label={t('PrivateCertificate')}
                                    placeholder="-----BEGIN RSA PRIVATE KEY-----\nQCdYIUkYi...."
                                    formik={formik}
                                    />
                                <CosmosInputText
                                    multiline
                                    name="TLSCert"
                                    label={t('PublicCertificate')}
                                    placeholder="-----BEGIN CERTIFICATE-----\nMIIEowIBwIBAA...."
                                    formik={formik}
                                />
                            </>
                        )}
                        
                        {hostError && <Grid item xs={12}>
                          <Alert color='error'>{hostError}</Alert>
                        </Grid>}
                        {hostIp && <Grid item xs={12}>
                            <Alert color='info'><Trans i18nKey="HostnamePointsTo" hostIp={hostIp}>This hostname is pointing to <strong>{{hostIp}}</strong>, check that it is your server IP!</Trans></Alert>
                        </Grid>}

                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && formik.values.UseWildcardCertificate && (!formik.values.DNSChallengeProvider || formik.values.DNSChallengeProvider == '') && (
                            <Alert severity="error">
                                {t('WildcardLetsEncrypt')}
                            </Alert>
                        )}
                        
                        {(formik.values.HTTPSCertificateMode === "LETSENCRYPT" || formik.values.HTTPSCertificateMode === "SELFSIGNED") && formik.values.Hostname && formik.values.Hostname.match(hostnameIsDomainReg) && (
                        <CosmosCheckbox
                            label={t('UseWildcartCertificate') + (formik.values.Hostname ||  "")}
                            name="UseWildcardCertificate"
                            formik={formik}
                        />)}
                        
                        
                        {formik.values.HTTPSCertificateMode != "" && (formik.values.HTTPSCertificateMode != "DISABLED" || isDomain(formik.values.Hostname)) ? (
                        <Grid item xs={12}>
                        <CosmosCheckbox 
                            label={<span>{t('AllowHTTPLocalIPAccess')} &nbsp;
                            <Tooltip title={<span style={{fontSize:'110%'}}><Trans i18nKey="AllowHTTPLocalIPAccessTitle">
                                When HTTPS is used along side a domain, depending on your networking configuration, it is possible that your server is not receiving direct local connections. <br />
                                This option allows you to also access your Cosmos admin using your local IP address, like ip:port. <br />
                                You can already create ip:port URLs for your apps, <strong>but this will make them HTTP-only</strong>.</Trans></span>}>
                                <QuestionCircleOutlined size={'large'} />
                            </Tooltip></span>}
                            name="allowHTTPLocalIPAccess"
                            formik={formik}
                        />
                        {formik.values.allowHTTPLocalIPAccess && <Alert severity="warning"><Trans i18nKey="AllowHTTPLocalIPAccessAlert">
                        This option is not recommended as it exposes your server to security risks on your local network. <br />
                        Your local network is safer than the internet, but not safe, as devices like IoTs, smart-TVs, smartphones or even your router can be compromised. <br />
                        <strong>If you want to have a secure offline / local-only access to a server that uses a domain name and HTTPS, use Constellation instead.</strong>
                        </Trans></Alert>}
                        </Grid>) : ""}

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
                                {formik.isSubmitting ? t('Loading') : (
                                    formik.values.HTTPSCertificateMode === "DISABLE" ? t('Disable') : t('Update')
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
                return (status && status.hostname != '0.0.0.0') ? t('Next') : '';
            }
        },
        {
            label: t('AdminAccount'),
            component: <div>
                <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> {t('AdminAccountText')}
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
                        nickname: Yup.string().required(t('NicknameRequired')).min(3).max(32)
                        .matches(/^(?!admin|root).*$/, t('NicknameRootAdmin')),
                        password: Yup.string().required(t('PasswordRequired')).min(8).max(128).matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[~!@#$%\^&\*\(\)_\+=\-\{\[\}\]:;"'<,>\/])(?=.{9,})/, 'Password must contain 9 characters: at least 1 lowercase, 1 uppercase, 1 number, and 1 special character'),
                        email: Yup.string().email(t('InvalidEmail')).max(255),
                        confirmPassword: Yup.string().oneOf([Yup.ref('password'), null], t('PasswordsMatch')),
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
                                label={t('Nickname')}
                                placeholder={t('Nickname')}
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
                                label={t('Password')}
                                placeholder={t('Password')}
                                formik={formik}
                                type="password"
                            />
                            <CosmosInputText
                                name="confirmPassword"
                                label={t('ConfirmPassword')}
                                placeholder={t('Password')}
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
                                    {formik.isSubmitting ? t('Loading') : t('Create')}
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
            label: t('Finish'),
            component: <div><Trans i18nKey="newInstallFinishText">
                Well done! You have successfully installed Cosmos. You can now login to your server using the admin account you created.
                If you have changed the hostname, don't forget to use that URL to access your server after the restart.
                If you have are running into issues, check the logs for any error messages and edit the file in the /config folder. 
                If you still don't manage, please join our <a target="_blank" rel="noopener noreferrer" href="https://discord.gg/PwMWwsrwHA">Discord server</a> and we'll be happy to help!
            </Trans></div>,
            nextButtonLabel: () => {
                return t('ApplyRestart');
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
                    >{t('Back')}</Button>

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
                                    redirectTo(hostname + "/cosmos-ui/login");
                                }, 500);
                            } else {
                                setActiveStep(activeStep + 1)
                            }
                        }}
                    >{
                        steps[activeStep].nextButtonLabel() ? 
                            steps[activeStep].nextButtonLabel() :
                            t('Next')
                    }</Button>
                </Stack>
            </Grid>
        </Grid>
    </AuthWrapper>
};

export default NewInstall;
