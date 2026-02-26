import { SettingOutlined } from "@ant-design/icons";
import { Chip } from "@mui/material";
import { useEffect, useState } from "react";
import { getOrigin, getFullOrigin } from "../utils/routes";
import { useTheme } from '@mui/material/styles';
import StatusDot from "./statusDot";

const HostChip = ({route, settings, style, ellipsis}) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [isOnline, setIsOnline] = useState(null);
  const url = getOrigin(route)

  useEffect(() => {
    fetch(getFullOrigin(route), {
      method: 'HEAD',
      mode: 'no-cors',
    }).then((res) => {
      setIsOnline(true);
    }).catch((err) => {
      setIsOnline(false);
    });
  }, [url]);

  return <Chip
    label={<><StatusDot status={isOnline == null ? "unknown" : isOnline ? "success" : "error"} size={8} style={{ marginRight: 6 }} />{url}</>}
    color="primary"
    variant="outlined"
    style={{
      paddingRight: '4px',
      // textDecoration: isOnline ? 'none' : 'underline wavy red',
      ...style,
      ...(ellipsis ? { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '300px' } : {})
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