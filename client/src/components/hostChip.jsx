import { SettingOutlined } from "@ant-design/icons";
import { Chip, Tooltip } from "@mui/material";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { getOrigin, getFullOrigin, IsRouteSocketProxy } from "../utils/routes";
import { isContainerRunning } from "../utils/container-status";
import StatusDot from "./statusDot";

// Green (service is up): 2xx/3xx (incl. opaqueredirect for cross-origin login
// redirects) and the 4xx codes that mean the reverse proxy answered while
// refusing this caller/method (401, 403, 405, 407, 429, 511) - the app is
// running and reachable, it just won't serve this HEAD/probe.
// A 503 from the lazy probe means "container is dormant" -> still reachable
// (sleeping), handled by the caller.
// Red (service is down): 404, 408, genuine 5xx, and network errors.
function classifyProbeStatus(res) {
  if (res.type === 'opaqueredirect') return true;
  const s = res.status;
  if (s >= 200 && s < 400) return true;
  if (s === 401 || s === 403 || s === 405 || s === 407 || s === 429 || s === 511) return true;
  return false;
}

// Probes the route through the Cosmos reverse proxy without waking a lazy
// (sleeping) container. The proxy answers the __cosmos_probe HEAD request
// itself with X-Cosmos-Container: sleeping (HTTP 503) when the container is
// dormant - that is "reachable but asleep", not "offline".
const probeRoute = async (route) => {
  const origin = getFullOrigin(route);
  const probeUrl = origin + (origin.includes('?') ? '&' : '?') + '__cosmos_probe=1';

  try {
    const res = await fetch(probeUrl, {
      method: 'HEAD',
      mode: 'cors',
      credentials: 'include',
      redirect: 'manual',
      cache: 'no-store',
    });
    const sleeping = res.headers.get('X-Cosmos-Container') === 'sleeping';
    // A sleeping lazy container is 503 + sleeping header: reachable, not offline.
    const online = sleeping || classifyProbeStatus(res);
    return { sleeping, online };
  } catch (e) {
    // CORS error: the proxy answered but hid the status (cross-origin route,
    // header hardening). Fall back to an opaque no-cors probe of the same URL.
    try {
      await fetch(probeUrl, { method: 'HEAD', mode: 'no-cors', cache: 'no-store' });
      return { sleeping: false, online: true };
    } catch (e2) {
      return { sleeping: false, online: false };
    }
  }
};

const HostChip = ({route, settings, container, style, ellipsis}) => {
  const { t } = useTranslation();
  const [status, setStatus] = useState(null); // null | { sleeping: bool, online: bool }
  const url = getOrigin(route);
  const isSocketProxy = route && IsRouteSocketProxy(route);

  // Container run state gating: when a container is passed and it is not
  // actually running, show a grey dot and do not probe. Socket proxies show
  // green purely from container run state (no HTTP layer to probe).
  const containerDown = container && !isContainerRunning(container);

  useEffect(() => {
    if (containerDown) {
      setStatus({ sleeping: false, online: null });
      return;
    }
    if (isSocketProxy) {
      setStatus({ sleeping: false, online: true });
      return;
    }
    let cancelled = false;
    probeRoute(route).then((s) => { if (!cancelled) setStatus(s); });
    return () => { cancelled = true; };
  }, [url, containerDown, isSocketProxy, route]);

  const dot = status && status.sleeping
    ? <Tooltip title={t('global.containerSleeping')}><StatusDot status="unknown" hollow size={8} style={{ marginRight: 6 }} /></Tooltip>
    : <StatusDot
        status={!status || status.online === null ? "unknown" : status.online ? "success" : "error"}
        size={8}
        style={{ marginRight: 6 }}
      />;

  return <Chip
    label={<>{dot}{url}</>}
    color="primary"
    variant="outlined"
    style={{
      paddingRight: '4px',
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
