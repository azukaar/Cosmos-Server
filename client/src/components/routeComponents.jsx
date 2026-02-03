import { CheckOutlined, ClockCircleOutlined, CopyOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined } from "@ant-design/icons";
import { Card, Chip, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
import { useState } from "react";
import { useTheme } from '@mui/material/styles';
import { Trans, useTranslation } from 'react-i18next';
import MenuButton from './MenuButton';

let routeImages = {
  "TUNNEL": {
    label: "Tunnel",
    icon: "TU",
    backgroundColor: "#082452",
    color: "white",
  },
  "SERVAPP": {
    label: "ServApp",
    icon: "SA",
    backgroundColor: "#0db7ed",
    color: "black",
  },
  "STATIC": {
    label: "Static",
    icon: "ST",
    backgroundColor: "#f9d71c",
    color: "black",
  },
  "REDIRECT": {
    label: "Redir",
    icon: "RE",
    backgroundColor: "#2c3e50",
    color: "white",
  },
  "PROXY": {
    label: "Proxy",
    icon: "PR",
    backgroundColor: "#2ecc71",
    color: "black",
  },
  "SPA": {
    label: "SPA",
    icon: "SP",
    backgroundColor: "#e74c3c",
    color: "black",
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

  return c ? <><Tooltip title={c.label}>
    <Chip
      icon={<span>{cicon}</span>}
      // label={c.label}
      sx={{
        backgroundColor: c.backgroundColor,
        paddingLeft: "5px",

        '& .MuiChip-label': {
          paddingLeft: 1,
          paddingRight: 1,
        },

        '& .MuiChip-icon': {
          color: c.color,
        },
      }}
    ></Chip>
    </Tooltip>
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
  const { t } = useTranslation();

  return (
    <MenuButton>
      <MenuItem onClick={(event) => up(event)}>
        <ListItemIcon>
          <UpOutlined />
        </ListItemIcon>
        <ListItemText>{t('global.moveUp')}</ListItemText>
      </MenuItem>
      <MenuItem onClick={(event) => down(event)}>
        <ListItemIcon>
          <DownOutlined />
        </ListItemIcon>
        <ListItemText>{t('global.moveDown')}</ListItemText>
      </MenuItem>
      <MenuItem onClick={(event) => duplicateRoute(event)}>
        <ListItemIcon>
          <CopyOutlined />
        </ListItemIcon>
        <ListItemText>{t('global.duplicate')}</ListItemText>
      </MenuItem>
      <MenuItem onClick={(event) => {
        if (confirmDelete) {
          deleteRoute(event);
          setConfirmDelete(false);
        } else {
          setConfirmDelete(true);
        }
      }}>
        <ListItemIcon>
          <DeleteOutlined style={{color: confirmDelete ? 'red' : 'inherit'}} />
        </ListItemIcon>
        <ListItemText style={{color: confirmDelete ? 'red' : 'inherit'}}>
          {confirmDelete ? t('global.confirmDelete') : t('global.delete')}
        </ListItemText>
      </MenuItem>
    </MenuButton>
  );
}