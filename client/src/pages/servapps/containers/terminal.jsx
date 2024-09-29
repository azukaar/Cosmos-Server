import React, { useEffect, useRef, useState } from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import * as API from '../../../api';
import { Alert, Input, Stack, useMediaQuery, useTheme } from '@mui/material';
import { ApiOutlined, SendOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';
import { useTranslation } from 'react-i18next';

import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'
import { FitAddon } from '@xterm/addon-fit';
import TerminalComponent from '../../../components/terminal';

const DockerTerminal = ({containerInfo, refresh}) => {
  const { t } = useTranslation();
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

  return (
    <div className="terminal-container" style={{
      background: '#000',
      width: '100%', 
      maxWidth: '900px', 
      position:'relative'
    }}>
      {(!isInteractive) && (
        <Alert severity="warning">
          {t('mgmt.servapps.containers.terminal.terminalNotInteractiveWarning')}
          <Button onClick={() => makeInteractive()}>{t('mgmt.servapps.containers.terminal.ttyEnableButton')}</Button>
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