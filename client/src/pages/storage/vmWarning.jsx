import { Alert } from "@mui/material";

export default function VMWarning() {
  return (<Alert severity="warning">
    You are running Cosmos inside a Docker container or a VM. As such, it only has limited access to your disks and their informations.
    For your safety, potentially destructive operations such as formatting, mounting, RAIDing, are disabled as your VM/Docker setup could vary and potentially mislead you, causing irreversible damages.
  </Alert>);
}