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
            return [["", t('auth.hostnameInput')]];
        }

        if(hostname.match(hostnameIsDomainReg) && !hostname.endsWith('.local')) {
            return [
                ["", t('mgmt.config.security.encryption.httpsCertSelection.httpsCertSelection')],
                ["LETSENCRYPT", t('mgmt.config.security.encryption.httpsCertSelection.sslLetsEncryptChoice')],
                ["SELFSIGNED", t('mgmt.config.security.encryption.httpsCertSelection.sslSelfSignedChoice')],
                ["PROVIDED", t('mgmt.config.security.encryption.httpsCertSelection.sslProvidedChoice')],
                ["DISABLED", t('mgmt.config.security.encryption.httpsCertSelection.sslDisabledChoice')],
            ]
        } else {
            return [
                ["", t('mgmt.config.security.encryption.httpsCertSelection.httpsCertSelection')],
                ["SELFSIGNED", t('mgmt.config.security.encryption.httpsCertSelection.sslSelfSignedChoice')],
                ["PROVIDED", t('mgmt.config.security.encryption.httpsCertSelection.sslProvidedChoice')],
                ["DISABLED", t('mgmt.config.security.encryption.httpsCertSelection.sslDisabledChoice')],
            ]
        }
    }

    const steps = [
        {
            label: t('newInstall.welcomeTitle'),
            component: <div>
                {t('newInstall.welcomeText')}
                <br /><br />
                <Checkbox checked={cleanInstall} onChange={(e) => setCleanInstall(e.target.checked)} />{t('newInstall.cleanInstallCheckbox')}
                <br /><br />
                <a style={{color: 'white', textDecoration: 'none'}} target='_blank' rel="noopener noreferrer" href="https://cosmos-cloud.io/doc/2%20setup">
                    <Button variant="outlined" color="inherit" startIcon={<QuestionCircleOutlined />}>
                     {t('newInstall.linkToDocs')}
                    </Button>
                </a>
            </div>,
            nextButtonLabel: () => {
                return 'Start';
            }
        },
        {
            label: t('newInstall.dockerTitle'),
            component: <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> {t('newInstall.whatIsCosmos')}
                </div>
                {status && (status.docker ? 
                <Alert severity="success">
                    {t('newInstall.dockerAvail')}
                </Alert> :
                <Alert severity="error"><Trans i18nKey="newInstall.dockerNotConnected" />
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
                    {t('newInstall.dockerChecking')}
                </div>
                <div>
                    <center><CircularProgress color="inherit" /></center>
                </div></>)}
            </Stack>,
            nextButtonLabel: () => {
                return status && status.docker ? t('global.next') : t('newInstall.skipAction');
            }
        },
        {
            label: t('newInstall.dbTitle'),
            component:  <Stack item xs={12} spacing={2}>
                <div>
                <QuestionCircleOutlined /> {t('newInstall.dbText')}
                </div>
                {(status && status.database) ? 
                    <Alert severity="success">
                        {t('newInstall.dbConnected')}
                    </Alert> :
                    <><Alert severity="error">
                        {t('newInstall.dbNotConnected')}
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
                                    title={t('newInstall.dbInstalling')}
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
                                    label={t('newInstall.dbSelection.dbLabel')}
                                    formik={formik}
                                    options={[
                                        ["Create", t('newInstall.dbSelection.createChoice')],
                                        ["Provided", t('newInstall.dbSelection.providedChoice')],
                                        ["DisableUserManagement", t('newInstall.dbSelection.disabledChoice')],
                                    ]}
                                />
                                {formik.values.DBMode === "Provided" && (
                                    <>
                                    <CosmosInputText
                                        name="MongoDB"
                                        label={t('newInstall.dbUrlInput.dbUrlLabel')}
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
                                        {formik.isSubmitting ? t('newInstall.loading') : (
                                            formik.values.DBMode === "DisableUserManagement" ? t('newInstall.usermgmt.disableButton') : t('mgmt.servapps.containers.terminal.connectButton')
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
                return (status && status.database) ? t('global.next') : '';
            }
        },
        {
            label: t('newInstall.httpsTitle'),
            component: (<Stack item xs={12} spacing={2}>
            <div>
                <QuestionCircleOutlined /> <Trans i18nKey="newInstall.httpsText" />
                <Alert severity="info">
                  {t('newInstall.httpsText.info')}
                </Alert>
                <Alert severity="warning">
                    <Trans i18nKey="newInstall.httpsText.warning" />
                </Alert>
            </div>
            <div>
            <Formik
                initialValues={{
                    SSLEmail: "",
                    TLSKey: "",
                    TLSCert: "",
                    Hostname: "cosmos.local",
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
                            then: Yup.string().required().matches(hostnameIsDomainReg, t('newInstall.letsEncryptChoiceOnlyfqdnValidation')),
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
                        setErrors({ submit: t('newInstall.checkInputValidation') });
                        setSubmitting(false);
                    }
                }}>
                {(formik) => (
                    <form noValidate onSubmit={formik.handleSubmit}>
                        <Stack item xs={12} spacing={2}>
                        <CosmosInputText
                            name="Hostname"
                            label={t('newInstall.hostnameInput.hostnameLabel')}
                            placeholder={t('newInstall.hostnameInput.hostnamePlaceholder')}
                            formik={formik}
                            onChange={(e) => {
                              checkHost(e.target.value, setHostError, setHostIp);
                            }}
                        />
                        {formik.values.Hostname && ((formik.values.Hostname.match(hostnameIsDomainReg) && !formik.values.Hostname.endsWith('.local')) ? 
                            <Alert severity="info">
                                <Trans i18nKey="newInstall.fqdnAutoLetsEncryptInfo" />
                            </Alert>
                            :
                            <Alert severity="info">
                                <Trans i18nKey="newInstall.localAutoSelfSignedInfo" />
                            </Alert>)
                        }
                        <CosmosSelect
                            name="HTTPSCertificateMode"
                            label={t('auth.selectOption')}
                            formik={formik}
                            options={getHTTPSOptions(formik.values.Hostname && formik.values.Hostname)}
                        />
                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && (
                            <>
                            <Alert severity="warning"><Trans i18nKey="newInstall.LetsEncrypt.cloudflareWarning" /> </Alert>
                            <CosmosInputText
                                name="SSLEmail"
                                label={t('newInstall.sslEmailInput.sslEmailLabel')}
                                placeholder={"email@domain.com"}
                                formik={formik}
                            />
                            {formik.values.DNSChallengeProvider && formik.values.DNSChallengeProvider != '' && (
                                <Alert severity="info"><Trans i18nKey="newInstall.LetsEncrypt.dnsChallengeInfo"/></Alert>
                            )}
                            <DnsChallengeComp 
                                label={t('mgmt.config.security.encryption.sslLetsEncryptDnsSelection.sslLetsEncryptDnsLabel')}
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
                                    label={t('newInstall.privCertInput.privCertLabel')}
                                    placeholder="-----BEGIN RSA PRIVATE KEY-----\nQCdYIUkYi...."
                                    formik={formik}
                                    />
                                <CosmosInputText
                                    multiline
                                    name="TLSCert"
                                    label={t('newInstall.pubCertInput.pubCertLabel')}
                                    placeholder="-----BEGIN CERTIFICATE-----\nMIIEowIBwIBAA...."
                                    formik={formik}
                                />
                            </>
                        )}
                        
                        {hostError && <Grid item xs={12}>
                          <Alert color='error'>{hostError}</Alert>
                        </Grid>}
                        {hostIp && <Grid item xs={12}>
                            <Alert color='info'><Trans i18nKey="newInstall.hostnamePointsToInfo" values={{hostIp: hostIp}} /></Alert>
                        </Grid>}

                        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && formik.values.UseWildcardCertificate && (!formik.values.DNSChallengeProvider || formik.values.DNSChallengeProvider == '') && (
                            <Alert severity="error">
                                {t('newInstall.wildcardLetsEncryptError')}
                            </Alert>
                        )}
                        
                        {(formik.values.HTTPSCertificateMode === "LETSENCRYPT" || formik.values.HTTPSCertificateMode === "SELFSIGNED") && formik.values.Hostname && formik.values.Hostname.match(hostnameIsDomainReg) && (
                        <CosmosCheckbox
                            label={t('newInstall.wildcardLetsEncryptCheckbox.wildcardLetsEncryptLabel') + (formik.values.Hostname ||  "")}
                            name="UseWildcardCertificate"
                            formik={formik}
                        />)}
                        
                        
                        {formik.values.HTTPSCertificateMode != "" && (formik.values.HTTPSCertificateMode != "DISABLED" || isDomain(formik.values.Hostname)) ? (
                        <Grid item xs={12}>
                        <CosmosCheckbox 
                            label={<span>{t('mgmt.config.http.allowInsecureLocalAccessCheckbox.allowInsecureLocalAccessLabel')} &nbsp;
                            <Tooltip title={<span style={{fontSize:'110%'}}><Trans i18nKey="mgmt.config.http.allowInsecureLocalAccessCheckbox.allowInsecureLocalAccessTooltip" /></span>}>
                                <QuestionCircleOutlined size={'large'} />
                            </Tooltip></span>}
                            name="allowHTTPLocalIPAccess"
                            formik={formik}
                        />
                        {formik.values.allowHTTPLocalIPAccess && <Alert severity="warning"><Trans i18nKey="mgmt.config.http.allowInsecureLocalAccessCheckbox.allowInsecureLocalAccessWarning" /></Alert>}
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
                                {formik.isSubmitting ? t('newInstall.loading') : (
                                    formik.values.HTTPSCertificateMode === "DISABLE" ? t('newInstall.usermgmt.disableButton') : t('global.update')
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
                return (status && status.hostname != '0.0.0.0') ? t('global.next') : '';
            }
        },
        {
            label: t('newInstall.adminAccountTitle'),
            component: <div>
                <Stack item xs={12} spacing={2}>
                <div>
                    <QuestionCircleOutlined /> {t('newInstall.adminAccountText')}
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
                        nickname: Yup.string().required(t('global.nicknameRequiredValidation')).min(3).max(32)
                        .matches(/^(?!admin|root).*$/, t('newInstall.setupUser.nicknameRootAdminNotAllowedValidation')),
                        password: Yup.string().required(t('auth.pwdRequired')).min(8).max(128).matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[~!@#$%\^&\*\(\)_\+=\-\{\[\}\]:;"'<,>\/])(?=.{9,})/, 'Password must contain 9 characters: at least 1 lowercase, 1 uppercase, 1 number, and 1 special character'),
                        email: Yup.string().email(t('global.emailInvalidValidation')).max(255),
                        confirmPassword: Yup.string().oneOf([Yup.ref('password'), null], t('newInstall.setupUser.passwordMustMatchValidation')),
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
                                label={t('global.nicknameLabel')}
                                placeholder={t('global.nicknameLabel')}
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
                                label={t('auth.pwd')}
                                placeholder={t('auth.pwd')}
                                formik={formik}
                                type="password"
                            />
                            <CosmosInputText
                                name="confirmPassword"
                                label={t('auth.confirmPassword')}
                                placeholder={t('auth.pwd')}
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
                                    {formik.isSubmitting ? t('newInstall.loading') : t('global.createAction')}
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
            label: t('newInstall.finishTitle'),
            component: <div><Trans i18nKey="newInstall.finishText" components={[<a target="_blank" rel="noopener noreferrer" href="https://discord.gg/PwMWwsrwHA"></a>]} /></div>,
            nextButtonLabel: () => {
                return t('newInstall.applyRestartAction');
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
                    >{t('global.backAction')}</Button>

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
                            t('global.next')
                    }</Button>
                </Stack>
            </Grid>
        </Grid>
    </AuthWrapper>
};

export default NewInstall;
