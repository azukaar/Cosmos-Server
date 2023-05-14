import * as React from 'react';
import MainCard from '../../../components/MainCard';
import RestartModal from '../../config/users/restart';
import { Alert, Button, Chip, Divider, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../../components/routeComponents';
import { getFaviconURL } from '../../../utils/routes';
import * as API from '../../../api';
import { AppstoreOutlined, ArrowUpOutlined, BulbTwoTone, CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, PlusCircleOutlined, SyncOutlined, UpOutlined } from "@ant-design/icons";
import IsLoggedIn from '../../../isLoggedIn';
import PrettyTabbedView from '../../../components/tabbedView/tabbedView';
import Back from '../../../components/back';
import { useParams } from 'react-router';
import ContainerOverview from './overview';
import Logs from './logs';
import DockerContainerSetup from './setup';
import NetworkContainerSetup from './network';
import VolumeContainerSetup from './volumes';
import DockerTerminal from './terminal';
import { Link } from 'react-router-dom';
import { smartDockerLogConcat, tryParseProgressLog } from '../../../utils/docker';

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '500px',
  maxWidth: '100%',
  width: '100%',
  margin: '0',
  position: 'relative',
  fontSize: '12px',
  fontFamily: 'monospace',
  whiteSpace: 'pre-wrap',
  wordWrap: 'break-word',
  wordBreak: 'break-all',
  lineHeight: '1.5',
  boxShadow: '0 0 10px rgba(0,0,0,0.5)',
  border: '1px solid rgba(255,255,255,0.1)',
  boxSizing: 'border-box',
  marginBottom: '10px',
  marginTop: '10px',
  marginLeft: '0',
  marginRight: '0',
  display: 'block',
  textAlign: 'left',
  verticalAlign: 'baseline',
  opacity: '1',
}

const NewDockerService = ({service}) => {
  const { containerName } = useParams();
  const [container, setContainer] = React.useState(null);
  const [config, setConfig] = React.useState(null);
  const [log, setLog] = React.useState([]);
  const [isDone, setIsDone] = React.useState(false);
  const [openModal, setOpenModal] = React.useState(false);
  const preRef = React.useRef(null);

  React.useEffect(() => {
    // refreshContainer();
  }, []);

  const create = () => {
    setLog([])
    API.docker.createService(service, (newlog) => {
      setLog((old) => smartDockerLogConcat(old, newlog));
      preRef.current.scrollTop = preRef.current.scrollHeight;
      if (newlog.includes('[OPERATION SUCCEEDED]')) {
        setIsDone(true);
      }
    });
  }

  const needsRestart = service && service.service && service.service.some((c) => {
    return c.routes && c.routes.length > 0;
  }); 

  return   <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
    <MainCard title="Create Service">
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} />
    <Stack spacing={1}>
      {!isDone && <Button 
        onClick={create}
        variant="contained"
        color="primary"
        fullWidth
        startIcon={<PlusCircleOutlined />}
      >Create</Button>}
      {isDone && <Stack spacing={1}>
        <Alert severity="success">Service Created!</Alert>
        {needsRestart && <Alert severity="warning">Cosmos needs to be restarted to apply changes to the URLs</Alert>}
        {needsRestart &&
          <Button
            variant="contained"
            color="primary"
            fullWidth
            startIcon={<SyncOutlined />}
            onClick={() => {
              setOpenModal(true);
            }}
            >Restart</Button>
        }
      </Stack>}
      <pre style={preStyle} ref={preRef}>
        {!log.length && JSON.stringify(service, false ,2)}
        {log.map((l) => {
          return <div>{tryParseProgressLog(l)}</div>
        })}
      </pre>
    </Stack>
  </MainCard>
  </div>;
}

export default NewDockerService;