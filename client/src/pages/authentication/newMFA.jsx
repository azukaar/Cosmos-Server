import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

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
  Alert,
  TextField,
  Tooltip
} from '@mui/material';

// project import
import AuthWrapper from './AuthWrapper';

import { useEffect, useState, useRef } from 'react';

import * as Yup from 'yup';
import * as API from '../../api';

import QRCode from 'qrcode';
import { useTheme } from '@mui/material/styles';
import { Formik } from 'formik';
import { LoadingButton } from '@mui/lab';
import { CosmosCollapse } from '../config/users/formShortcuts';
import { redirectToLocal } from '../../utils/indexs';

const MFALoginForm = () => {
  const { t } = useTranslation();
  const urlSearchParams = new URLSearchParams(window.location.search);
  const redirectToURL = urlSearchParams.get('redirect') ? urlSearchParams.get('redirect') : '/cosmos-ui';

  useEffect(() => {
    API.auth.me().then((data) => {
        if(data.status == 'OK') {
          redirectToLocal(redirectToURL);
        } else if(data.status == 'NEW_INSTALL') {
          redirectToLocal('/cosmos-ui/newInstall');
        }
    });
  });  
  
  return <Formik
    initialValues={{
      token: '',
    }}
    validationSchema={Yup.object().shape({
      token: Yup.string().required(t('mgmt.openid.newMfa.tokenRequiredValidation')).min(6, t('mgmt.openid.newMfa.tokenmin6charValidation')).max(6, t('mgmt.openid.newMfa.tokenmax6charValidation')),
    })}
    onSubmit={(values, { setSubmitting, setStatus, setErrors }) => {
      API.users.check2FA(values.token).then((data) => {
        redirectToLocal(redirectToURL);
      }).catch((error) => {
        console.log(error)
        setStatus({ success: false });
        setErrors({ submit: t('mgmt.openid.newMfa.wrongOtpValidation') });
        setSubmitting(false);
      });
    }}
  >
    {(formik) => (
      <form autoComplete="off" noValidate onSubmit={formik.handleSubmit}>
        <Stack spacing={3}>
          <TextField
            fullWidth
            autoComplete="off"
            type="text"
            label="Token"
            {...formik.getFieldProps('token')}
            error={formik.touched.token && formik.errors.token && true}
            helperText={formik.touched.token && formik.errors.token && formik.errors.token}
            autoFocus
          />
          {formik.errors.submit && (
              <Grid item xs={12}>
                  <FormHelperText error>{formik.errors.submit}</FormHelperText>
              </Grid>
          )}
          <LoadingButton
            fullWidth
            size="large"
            type="submit"
            variant="contained"
            loading={formik.isSubmitting}
          >
            {t('auth.login')}
          </LoadingButton>
        </Stack>
      </form>
    )}
  </Formik>;
}

const MFASetup = () => {
  const [mfaCode, setMfaCode] = useState('');
  const canvasRef = useRef(null);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { t, Trans } = useTranslation();

  const getCode = () => {
    return API.users.new2FA().then(({data}) => {
      if (data) {
        setMfaCode(data.key);
        QRCode.toCanvas(canvasRef.current, data.key, {
          width: 300,
          color: {
            dark: theme.palette.secondary.main,
            light: '#ffffff'
          }
        }, function (error) {
          if (error) console.error(error)
        })
      }
    });
  };

  useEffect(() => {
    getCode();
  }, []);

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant="h5">
            {t("mgmt.openid.newMfa.requires2faText")} (FreeOTP(+), Google Authenticator, Microsoft Authenticator, ...)
        </Typography>
      </Grid>
      <Grid  item xs={12} textAlign={'center'}>
        <canvas style={{borderRadius: '15px'}} ref={canvasRef} />
      </Grid>
      <Grid item xs={12}>
        <Typography variant="h5">{t('mgmt.openid.newMfa.otpManualCode')}</Typography>
      </Grid>
      <Grid item xs={12}>
        <CosmosCollapse title={t('mgmt.openid.newMfa.otpManualCode.showButton')} defaultExpanded={false}>
        <div style={{padding: '20px', fontSize: '90%', borderRadius: '15px', background: 'rgba(0,0,0,0.2)'}}>
          {mfaCode && <span>{mfaCode.split('?')[1].split('&').map(a => <div>{decodeURI(a).replace('=', ': ')}</div>)}</span>}
        </div>
        </CosmosCollapse>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="h5">{t('mgmt.openid.newMfa.otpEnterTokenText')}</Typography>
      </Grid>
      <Grid item xs={12}>
        <MFALoginForm />
      </Grid>
      <Grid item xs={12}>
        <Link to="/cosmos-ui/logout">
          <Typography variant="h5">{t('global.logout')}</Typography>
        </Link>
      </Grid>
    </Grid>
  );
}

const NewMFA = () => {
    const { t } = useTranslation();

    return <AuthWrapper>
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                    <Typography variant="h3">{t('mgmt.openid.newMfa')}</Typography>
                </Stack>
            </Grid>
            <Grid item xs={12}>
                <MFASetup />
            </Grid>
        </Grid>
    </AuthWrapper>
}

const MFALogin = () => {
  const { t } = useTranslation();

  return <AuthWrapper>
      <Grid container spacing={3}>
          <Grid item xs={12}>
              <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: { xs: -0.5, sm: 0.5 } }}>
                  <Typography variant="h3">{t('mgmt.openid.newMfa.enterOtp')}</Typography>
              </Stack>
          </Grid>
          <Grid item xs={12}>
              <MFALoginForm />
          </Grid>
      </Grid>
  </AuthWrapper>
}

export default NewMFA;

export {
  MFASetup,
  NewMFA,
  MFALogin,
  MFALoginForm
};
