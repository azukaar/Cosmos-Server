import { SettingOutlined } from "@ant-design/icons";
import { Chip, Tooltip } from "@mui/material";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { getOrigin, getFullOrigin } from "../utils/routes";
import { useTheme } from '@mui/material/styles';
import StatusDot from "./statusDot";

const probeRoute = async (route) => {
  const origin = getFullOrigin(route);
  try {
    const res = await fetch(origin + (origin.includes('?') ? '&' : '?') + '__cosmos_probe=1', {
      method: 'HEAD',
      mode: 'cors',
      credentials: 'include',
      redirect: 'manual',
      cache: 'no-store',
    });
    return res.headers.get('X-Cosmos-Container') === 'sleeping' ? 'sleeping' : 'online';
  } catch (e) {
    try {
      await fetch(origin, { method: 'HEAD', mode: 'no-cors' });
      return 'online';
    } catch (e2) {
      return 'offline';
    }
  }
};

const HostChip = ({route, settings, style, ellipsis}) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [status, setStatus] = useState(null);
  const url = getOrigin(route)

  useEffect(() => {
    let cancelled = false;
    probeRoute(route).then((s) => { if (!cancelled) setStatus(s); });
    return () => { cancelled = true; };
  }, [url]);

  const dot = status === 'sleeping'
    ? <Tooltip title={t('global.containerSleeping')}><StatusDot status="unknown" hollow size={8} style={{ marginRight: 6 }} /></Tooltip>
    : <StatusDot status={status == null ? "unknown" : status === 'online' ? "success" : "error"} size={8} style={{ marginRight: 6 }} />;

  return <Chip
    label={<>{dot}{url}</>}
    color="primary"
    variant="outlined"
    style={{
      paddingRight: '4px',
      // textDecoration: isOnline ? 'none' : 'underline wavy red',
      ...style,
      ...(ellipsis ? { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '250px' } : {})
    }}
    onClick={() => {
      if(route.UseHost)
        window.open(window.location.origin.split("://")[0] + "://" + route.Host + route.PathPrefix, '_blank');
      else
        window.open(window.location.origin + route.PathPrefix, '_blank');
    }}
    onDelete={settings ? () => {
      window.open('/cosmos-ui/config-url/'+route.Name, '_blank');
    } : null}
    deleteIcon={settings ? <SettingOutlined /> : null}
  />
}

export default HostChip;