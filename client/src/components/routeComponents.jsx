import { CheckOutlined, ClockCircleOutlined, CopyOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined } from "@ant-design/icons";
import { Card, Chip, Stack, Tooltip } from "@mui/material";
import { useState } from "react";
import { useTheme } from '@mui/material/styles';
import { Trans, useTranslation } from 'react-i18next';

let routeImages = {
  "TUNNEL": {
    label: "Tunnel",
    icon: "💫",
    backgroundColor: "#082452",
    color: "white",
    colorDark: "white",
  },
  "SERVAPP": {
    label: "ServApp",
    icon: "🐳",
    backgroundColor: "#0db7ed",
    color: "white",
    colorDark: "black",
  },
  "STATIC": {
    label: "Static",
    icon: "📁",
    backgroundColor: "#f9d71c",
    color: "black",
    colorDark: "black",
  },
  "REDIRECT": {
    label: "Redir",
    icon: "🔀",
    backgroundColor: "#2c3e50",
    color: "white",
    colorDark: "white",
  },
  "PROXY": {
    label: "Proxy",
    icon: "🔗",
    backgroundColor: "#2ecc71",
    color: "white",
    colorDark: "black",
  },
  "SPA": {
    label: "SPA",
    icon: "🌐",
    backgroundColor: "#e74c3c",
    color: "white",
    colorDark: "black",
  },
}

export const RouteMode = ({route}) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  let c = routeImages[(route._IsTunnel ? "TUNNEL" : route.Mode.toUpperCase())];

  let cicon = c.icon;
  
  if (!route._IsTunnel && route.TunnelVia) {
    cicon = c.icon + " 💫";
  }

  return c ? <>
    <Chip
      icon={<span>{cicon}</span>}
      label={c.label}
      sx={{
        backgroundColor: c.backgroundColor,
        color: isDark ? c.colorDark : c.color,
        paddingLeft: "5px",
        alignItems: "right",
      }}
    ></Chip>
  </> : <></>;
}

export const RouteSecurity = ({route}) => {
  const { t } = useTranslation();

  return <div style={{fontWeight: 'bold', fontSize: '110%'}}>
    <Tooltip title={route.SmartShield && route.SmartShield.Enabled ? t('tooltip.route.SmartShield.enabled') : t('tooltip.route.SmartShield.disabled')}>
      <div style={{display: 'inline-block'}}>
        {route.SmartShield && route.SmartShield.Enabled ? 
          <SafetyOutlined style={{color: 'green'}} /> :
          <SafetyOutlined style={{color: 'red'}} />
        }
      </div>
    </Tooltip>
    &nbsp;
    <Tooltip title={route.AuthEnabled ? t('tooltip.route.authentication.enabled') : t('tooltip.route.authentication.disabled')}>
      <div style={{display: 'inline-block'}}>
        {route.AuthEnabled ? 
          <LockOutlined style={{color: 'green'}} /> :
          <LockOutlined style={{color: 'red'}} />
        }
      </div>
    </Tooltip>
    &nbsp;
    <Tooltip title={route.ThrottlePerMinute ? t('tooltip.route.throttling.enabled') : t('tooltip.route.throttling.disabled')}>
      <div style={{display: 'inline-block'}}>
        {route.ThrottlePerMinute ?
          <DashboardOutlined style={{color: 'green'}} /> :
          <DashboardOutlined style={{color: 'red'}} />
        }
      </div>
    </Tooltip>
    &nbsp;
    <Tooltip title={route.Timeout ? t('tooltip.route.timeout.enabled') : t('tooltip.route.timeout.disabled')}>
      <div style={{display: 'inline-block'}}>
        {route.Timeout ?
          <ClockCircleOutlined style={{color: 'green'}} /> :
          <ClockCircleOutlined style={{color: 'red'}} />
        }
      </div>
    </Tooltip>
  </div>
}


export const RouteActions = ({route, routeKey, up, down, deleteRoute, duplicateRoute}) => {
  const [confirmDelete, setConfirmDelete] = useState(false);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { t } = useTranslation();

  const miniChip = {
    width: '30px',
    height: '20px',
    display: 'inline-block',
    textAlign: 'center',
    cursor: 'pointer',
    color: theme.palette.text.secondary,
    fontSize: '12px',
    lineHeight: '20px',
    padding: '0px',
    borderRadius: '0px',
    background: isDark ? 'rgba(255, 255, 255, 0.03)' : '',
    fontWeight: 'bold',
  
    '&:hover': {
      backgroundColor: 'rgba(0, 0, 0, 0.04)',
    }
  }

  return <>
    <Stack direction={'row'} spacing={2} alignItems={'center'} justifyContent={'right'}>
      {!confirmDelete && (<Chip label={<DeleteOutlined />} onClick={() => setConfirmDelete(true)}/>)}
      {confirmDelete && (<Chip label={<CheckOutlined />} color="error" onClick={(event) => deleteRoute(event)}/>)}
      
      <Chip label={<CopyOutlined />} onClick={(event) => duplicateRoute(event)}/>

      <Tooltip title={t('tooltip.route.move')}>
        <Stack direction={'column'} spacing={0}>
          <Card sx={{...miniChip, borderBottom: 'none'}} onClick={(event) => up(event)}><UpOutlined /></Card>
          <Card sx={{...miniChip, cursor: 'auto'}}>{routeKey}</Card>
          <Card sx={{...miniChip, borderTop: 'none'}} onClick={(event) => down(event)}><DownOutlined /></Card>
        </Stack>
      </Tooltip>
    </Stack>
  </>;
}