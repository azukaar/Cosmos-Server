import { CheckOutlined, ClockCircleOutlined, CopyOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined, VerticalAlignTopOutlined, VerticalAlignBottomOutlined } from "@ant-design/icons";
import { Card, Chip, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
import { useState } from "react";
import { useTheme } from '@mui/material/styles';
import { Trans, useTranslation } from 'react-i18next';
import MenuButton from './MenuButton';

let routeImages = {
  "TUNNEL": {
    label: "Tunnel",
    icon: "TU",
    background: "linear-gradient(135deg, #082452, #0a3a7e)",
    color: "white",
  },
  "SERVAPP": {
    label: "ServApp",
    icon: "SA",
    background: "linear-gradient(135deg, #0db7ed, #0a8abf)",
    color: "black",
  },
  "STATIC": {
    label: "Static",
    icon: "ST",
    background: "linear-gradient(135deg, #f9d71c, #d4b010)",
    color: "black",
  },
  "REDIRECT": {
    label: "Redir",
    icon: "RE",
    background: "linear-gradient(135deg, #2c3e50, #1a252f)",
    color: "white",
  },
  "PROXY": {
    label: "Proxy",
    icon: "PR",
    background: "linear-gradient(135deg, #2ecc71, #1fa85a)",
    color: "black",
  },
  "SPA": {
    label: "SPA",
    icon: "SP",
    background: "linear-gradient(135deg, #e74c3c, #c0392b)",
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
        background: c.background,
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


export const RouteActions = ({route, routeKey, up, down, moveToTop, moveToBottom, deleteRoute, duplicateRoute}) => {
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
      {moveToTop && <MenuItem onClick={(event) => moveToTop(event)}>
        <ListItemIcon>
          <VerticalAlignTopOutlined />
        </ListItemIcon>
        <ListItemText>{t('global.moveToTop')}</ListItemText>
      </MenuItem>}
      {moveToBottom && <MenuItem onClick={(event) => moveToBottom(event)}>
        <ListItemIcon>
          <VerticalAlignBottomOutlined />
        </ListItemIcon>
        <ListItemText>{t('global.moveToBottom')}</ListItemText>
      </MenuItem>}
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