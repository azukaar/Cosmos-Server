import React, { useEffect, useRef, useState } from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import * as API from '../../../api';
import { Alert, Input, Stack, useMediaQuery, useTheme } from '@mui/material';
import { ApiOutlined, SendOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';
import { useTranslation } from 'react-i18next';
import PermissionGuard from '../../../components/permissionGuard';
import { PERM_RESOURCES, PERM_ADMIN } from '../../../utils/permissions';
import { useClientInfos } from '../../../utils/hooks';

import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'
import { FitAddon } from '@xterm/addon-fit';
import TerminalComponent from '../../../components/terminal';

const DockerTerminal = ({containerInfo, refresh}) => {
  const { t } = useTranslation();
  const { hasPermission, hasRolePermission } = useClientInfos();
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const isInteractive = Config.Tty;

  const makeInteractive = () => {
    API.docker.updateContainer(Name.slice(1), {interactive: 2})
      .then(() => {
        refresh && refresh();
      }).catch((e) => {
        console.error(e);
        refresh && refresh();
      });
  };

  if (!hasPermission(PERM_ADMIN)) {
    return (
      <div style={{ maxWidth: '1000px', width: '100%', margin: 'auto', padding: '20px 0' }}>
        <Alert severity="warning">
          {hasRolePermission(PERM_ADMIN)
            ? t('sudo.required')
            : t('mgmt.servapps.containers.terminal.terminalAccessDenied')}
        </Alert>
      </div>
    );
  }

  return (
    <div className="terminal-container" style={{
      background: '#000',
      width: '100%', 
      maxWidth: '900px', 
      margin: 'auto',
    }}>
      {(!isInteractive) && (
        <Alert severity="warning">
          {t('mgmt.servapps.containers.terminal.terminalNotInteractiveWarning')}
          <PermissionGuard permission={PERM_RESOURCES}><Button onClick={() => makeInteractive()}>{t('mgmt.servapps.containers.terminal.ttyEnableButton')}</Button></PermissionGuard>
        </Alert>
      )}
      
      <TerminalComponent refresh={refresh} connectButtons={
        [
          {
            label: t('mgmt.servapps.containers.terminal.connectButton'),
            onClick: (connect) => connect(() => API.docker.attachTerminal(Name.slice(1)), 'Main Process TTY')
          },
          {
            label: t('mgmt.servapps.containers.terminal.newShellButton'),
            onClick: (connect) => connect(() => API.docker.createTerminal(Name.slice(1)), 'Shell')
          }
        ]} />
    </div>
  );
};

export default DockerTerminal;