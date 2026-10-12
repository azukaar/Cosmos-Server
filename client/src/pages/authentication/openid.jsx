import { Link, useSearchParams } from 'react-router-dom';

// material-ui
import { Checkbox, Grid, Stack, Typography } from '@mui/material';

// project import
import AuthLogin from './auth-forms/AuthLogin';
import AuthWrapper from './AuthWrapper';
import { getOpenIDClientFaviconURL } from '../../utils/routes';
import { LoadingButton } from '@mui/lab';
import { Field, useFormik } from 'formik';
import { useEffect, useRef, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';

// ================================|| LOGIN ||================================ //

const OpenID = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const client_id = searchParams.get("client_id")
  const redirect_uri = searchParams.get("redirect_uri")
  const scope = searchParams.get("scope")
  const entireSearch = searchParams.toString()
  const [checkedScopes, setCheckedScopes] = useState(() => [...new Set([...(scope ? scope.split(' ') : []), "openid"])])
  const formRef = useRef(null)

  // if only 'account' as scope, auto redirect
  useEffect(() => {
    if (scope == "openid" && formRef && formRef.current) {
      formRef.current.submit();
    }
  }, [formRef]);
  

  const icon = getOpenIDClientFaviconURL(client_id);

  const selfport = new URL(window.location.href).port
  const selfprotocol = new URL(window.location.href).protocol + "//"
  const selfHostname = selfprotocol + (new URL(window.location.href).hostname) + (selfport ? ":" + selfport : "")

  const onchange = (e, scope) => {
    if (e.target.checked) {
      setCheckedScopes([...new Set([...checkedScopes, scope])])
    } else {
      setCheckedScopes(checkedScopes.filter((s) => s != scope))
    }
  }

  return (<AuthWrapper>
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Stack spacing={2}>
          <Typography variant="h3">{t('oidc.title', {client_id: client_id.replace("__route_", "")})}</Typography>
          <Stack direction="row" justifyContent="space-between" alignItems="baseline" spacing={2} style={{
            alignItems: 'center',
          }}>
            <img src={icon} alt={'icon'} width="64px" />
            <div>
              <Trans i18nKey='oidc.loginDescription' values={{client_id: client_id.replace("__route_", "")}} />
            </div>
          </Stack>
        </Stack>
      </Grid>
      <Grid item xs={12}>
			  <link rel="openid2.provider openid.server" href={selfHostname + "/oauth2/auth"} />
        <form action={"/oauth2/auth?" + entireSearch} method="post" ref={formRef}>
          <input type="hidden" name="client_id" value={client_id} />
          {[...new Set(scope ? scope.split(' ') : [])].map((scope) => {
            return scope == "openid" ? <div>
              <input type="checkbox" name="scopes" value={scope} checked hidden />
              <Checkbox checked disabled />
              {t('oidc.account')}
            </div>
              : <div>
                <input type="checkbox" name="scopes" hidden value={scope} checked={checkedScopes.includes(scope)} />
                <Checkbox checked={checkedScopes.includes(scope)} onChange={(e) => onchange(e, scope)} />
                {scope}
              </div>
          })}
          <div style={{
            fontSize: '0.8rem',
            marginTop: '15px',
            marginBottom: '20px',
            opacity: '0.8',
            fontStyle: 'italic',
          }}>
            <Trans i18nkey='oidc.redirectInfo' values={{redirect_uri: redirect_uri}} />
          </div>

          <LoadingButton
            disableElevation
            fullWidth
            size="large"
            type="submit"
            variant="contained"
            color="primary"
          >
            {t('oidc.loginTitle')}
          </LoadingButton>
        </form>
      </Grid>
    </Grid>
  </AuthWrapper>)
};

export default OpenID;
