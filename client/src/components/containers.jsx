import { WarningFilled } from "@ant-design/icons";
import { Tooltip } from "@mui/material";

export const ContainerNetworkWarning = ({container}) => (
  container.HostConfig.NetworkMode != "bridge" && container.HostConfig.NetworkMode != "default" && 
  <Tooltip title={`This container is using an incompatible network mode (${container.HostConfig.NetworkMode.slice(0, 16)}). If you want Cosmos to proxy to this container, enabling this option will change the network mode to bridge for you. Otherwise, you dont need to do anything, as the container is already isolated. Note that changing to bridge might break connectivity to other containers. To fix it, please use a private network and static ips instead.`}>
    <WarningFilled style={{color: 'red', fontSize: '18px', paddingLeft: '10px', paddingRight: '10px'}} />
  </Tooltip>
);
