import * as React from 'react';
import MainCard from '../../../components/MainCard';
import RestartModal from '../../config/users/restart';
import { Alert, Button, Chip, Divider, Stack, TextField, useMediaQuery } from '@mui/material';
import HostChip from '../../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../../components/routeComponents';
import { getFaviconURL } from '../../../utils/routes';
import * as API from '../../../api';
import { AppstoreOutlined, ArrowUpOutlined, BulbTwoTone, CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, PlusCircleOutlined, SyncOutlined, UpOutlined } from "@ant-design/icons";
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
import { LoadingButton } from '@mui/lab';
import LogLine from '../../../components/logLine';
import Highlighter from '../../../components/third-party/Highlighter';
import { useTranslation } from 'react-i18next';

import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-yaml';
import 'prismjs/themes/prism-okaidia.css';

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '520px',
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

const NewDockerService = ({service, refresh, edit}) => {
  const { t, i18n } = useTranslation();
  const { containerName } = useParams();
  const [container, setContainer] = React.useState(null);
  const [config, setConfig] = React.useState(null);
  const [log, setLog] = React.useState([]);
  const [isDone, setIsDone] = React.useState(false);
  const [openModal, setOpenModal] = React.useState(false);
  const preRef = React.useRef(null);
  const screenMin = useMediaQuery((theme) => theme.breakpoints.up('sm'))
  const [dockerCompose, setDockerCompose] = React.useState(JSON.stringify(service, null, 2));

  const installer = {...service['cosmos-installer']};
  service = {...service};
  delete service['cosmos-installer'];

  const refreshConfig = () => {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  };

  React.useEffect(() => {
    refreshConfig();
  }, []);

  const needsRestart = service && service.services && Object.values(service.services).some((c) => {
    return c.routes && c.routes.length > 0;
  }); 

  const create = () => {
    setLog([
      'Creating Service...                              ',
    ])
    API.docker.createService(JSON.parse(dockerCompose), (newlog) => {
      setLog((old) => smartDockerLogConcat(old, newlog));
      preRef.current.scrollTop = preRef.current.scrollHeight;
      if (newlog.includes('[OPERATION SUCCEEDED]')) {
        setIsDone(true);
        
        needsRestart && setOpenModal(true);
        refresh && refresh();
      }
    });
  }

  let isJSON = false;
  if(dockerCompose.trim().startsWith('{') && dockerCompose.trim().endsWith('}')) {
    isJSON = true;
  }

  return   <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
    <MainCard title={edit ? t('mgmt.servapps.container.compose.editServiceTitle') : t('mgmt.servapps.container.compose.createServiceButton')}>
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} config={config} newRoute />
    <Stack spacing={1}>
      {!isDone && <LoadingButton 
        onClick={create}
        variant="contained"
        color="primary"
        fullWidth
        className={edit ? '' : 'shinyButton'}
        loading={log.length && !isDone}
        startIcon={edit ? <SyncOutlined /> : <PlusCircleOutlined />}
      >{edit ? t('global.edit'): t('global.createAction')}</LoadingButton>}
      {isDone && <Stack spacing={1}>
        <Alert severity="success">{t('mgmt.servapps.container.compose.createServiceSuccess')}</Alert>
        {installer && installer['post-install'] && installer['post-install'].map(m =>{
          return <Alert severity={m.type}>{ /*couldn't test this yet, but shoud work for max 1 msg*/ installer?.translation?.[i18n?.resolvedLanguage]?.['post-install.label'] || installer?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['post-install.label'] || m.label }</Alert>
        })}
      </Stack>}
      

      {edit && !isDone && log.length ? <Button onClick={() => setLog([])}>{t('global.backAction')}</Button> : null}

      {log.length ? <pre style={preStyle} ref={preRef}>
        {log.map((l) => {
          return <LogLine message={tryParseProgressLog(l)} docker isMobile={!screenMin} />
        })}
      </pre> :
      <div>
        <Editor
          value={dockerCompose}
          placeholder={t('mgmt.servapps.pasteComposeButton.pasteComposePlaceholder')}
          onValueChange={code => setDockerCompose(code)}
          highlight={code => highlight(code, isJSON ? languages.json : languages.yaml)}
          padding={10}
          tabSize={2}
          insertSpaces={true}
          style={{
            fontFamily: '"Fira code", "Fira Mono", monospace',
            fontSize: 12,
            backgroundColor: '#000',
          }}
        />
      </div>
      }
    </Stack>
  </MainCard>
  </div>;
}

export default NewDockerService;