import { Link, useSearchParams } from 'react-router-dom';

// material-ui
import { Checkbox, Grid, Stack, Typography } from '@mui/material';

// project import
import AuthLogin from './auth-forms/AuthLogin';
import AuthWrapper from './AuthWrapper';
import { getFaviconURL } from '../../utils/routes';
import { LoadingButton } from '@mui/lab';
import { Field, useFormik } from 'formik';
import { useState } from 'react';

// ================================|| LOGIN ||================================ //

const OpenID = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const client_id = searchParams.get("client_id")
  const redirect_uri = searchParams.get("redirect_uri")
  const scope = searchParams.get("scope")
  const entireSearch = searchParams.toString()
  const [checkedScopes, setCheckedScopes] = useState(["openid"])

  let icon;

  // get hostname from redirect_uri with port
  let port, protocol, appHostname;

  try {
    port = new URL(redirect_uri).port
    protocol = new URL(redirect_uri).protocol + "//"
    appHostname = protocol + (new URL(redirect_uri).hostname) + (port ? ":" + port : "")
    icon = getFaviconURL({
      Mode: 'PROXY',
      Target: appHostname
    });
  } catch (e) {
    icon = getFaviconURL();
  }

  const selfport = new URL(window.location.href).port
  const selfprotocol = new URL(window.location.href).protocol + "//"
  const selfHostname = selfprotocol + (new URL(window.location.href).hostname) + (selfport ? ":" + selfport : "")

  const onchange = (e, scope) => {
    if (e.target.checked) {
      setCheckedScopes([...checkedScopes, scope])
    } else {
      setCheckedScopes(checkedScopes.filter((scope) => scope != scope))
    }
  }      

  return (<AuthWrapper>
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Stack spacing={2}>
          <Typography variant="h3">Login with OpenID - {client_id}</Typography>
          <Stack direction="row" justifyContent="space-between" alignItems="baseline" spacing={2} style={{
            alignItems: 'center',
          }}>
            <img src={icon} alt={'icon'} width="64px" />
            <div>
              You are about to login into <b>{client_id}</b>. <br />
              Check which permissions you are giving to this application. <br />
            </div>
          </Stack>
        </Stack>
      </Grid>
      <Grid item xs={12}>
			  <link rel="openid2.provider openid.server" href={selfHostname + "/oauth2/auth"} />
        <form action={"/oauth2/auth?" + entireSearch} method="post">
          <input type="hidden" name="client_id" value={client_id} />
          {scope.split(' ').map((scope) => {
            return scope == "openid" ? <div>
              <input type="checkbox" name="scopes" value={scope} checked hidden />
              <Checkbox checked disabled />
              account
            </div>
              : <div>
                <input type="checkbox" name="scopes" hidden value={scope} checked={checkedScopes.includes(scope)} />
                <Checkbox onChange={(e) => onchange(e, scope)} />
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
            You will be redirected to <b>{redirect_uri}</b> after login. <br />
          </div>

          <LoadingButton
            disableElevation
            fullWidth
            size="large"
            type="submit"
            variant="contained"
            color="primary"
          >
            OpenID Login
          </LoadingButton>
        </form>
      </Grid>
    </Grid>
  </AuthWrapper>)
};

export default OpenID;
