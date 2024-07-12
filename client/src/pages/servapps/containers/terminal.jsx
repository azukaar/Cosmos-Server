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

const DockerTerminal = ({containerInfo, refresh}) => {
  const { t } = useTranslation();
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const isInteractive = Config.Tty;
  const theme = useTheme();
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'))
  const [history, setHistory] = useState([]);
  const [historyCursor, setHistoryCursor] = useState(0);

  const [terminal] = useState(new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    disableStdin: false,
    fontFamily: 'monospace',
    fontSize: 14,
    fontWeight: 'normal',
    letterSpacing: 0,
    lineHeight: 1,
    logLevel: 'info',
    screenReaderMode: false,
    scrollback: 1000,
    tabStopWidth: 8,
  }));

  const xtermRef = useRef(null);

  const [message, setMessage] = useState('');
  const [output, setOutput] = useState([
    {
      output: 'Not Connected.',
      type: 'stdout'
    }
  ]);
  const ws = useRef(null);
  const [isConnected, setIsConnected] = useState(false);

  const makeInteractive = () => {
    API.docker.updateContainer(Name.slice(1), {interactive: 2})
      .then(() => {
        refresh && refresh();
      }).catch((e) => {
        console.error(e);
        refresh && refresh();
      });
  };

  const connect = (newProc) => {
    if(ws.current) {
      ws.current.close();
    }

    ws.current = newProc ? 
        API.docker.createTerminal(Name.slice(1))
      : API.docker.attachTerminal(Name.slice(1));

    ws.current.onmessage = (event) => {
      try {
        terminal.write(event.data);
      } catch (e) {
        console.error("error", e);
      }
    };

    ws.current.onclose = () => {
      setIsConnected(false);
      let terminalBoldRed = '\x1b[1;31m';
      let terminalReset = '\x1b[0m';
      terminal.write(terminalBoldRed + t('mgmt.servapps.containers.terminal.disconnectedFromText') + (newProc ? 'shell' : t('mgmt.servapps.containers.terminal.mainprocessTty')) + '\r\n' + terminalReset);
    };
    
    ws.current.onopen = () => {
      setIsConnected(true);
      let terminalBoldGreen = '\x1b[1;32m';
      let terminalReset = '\x1b[0m';
      terminal.write(terminalBoldGreen + t('mgmt.servapps.containers.terminal.connectedToText') + (newProc ? 'shell' : t('mgmt.servapps.containers.terminal.mainprocessTty')) + '\r\n' + terminalReset);
      // focus terminal
      terminal.focus();
    };

    return () => {
      setIsConnected(false);
      ws.current.close();
    };
  };

  const [SelectedText, setSelectedText] = useState('');

  useEffect(() => {
    // xtermRef.current.innerHTML = '';
    terminal.open(xtermRef.current);

    // const fitAddon = new FitAddon();
    // terminal.loadAddon(fitAddon);
    // fitAddon.fit();

    // const onFocus = () => {
    //   terminal.focus();
    // }

    // xtermRef.current.removeEventListener('touchstart', onFocus);

    // xtermRef.current.addEventListener('touchstart', onFocus);

    terminal.onSelectionChange((e) => {
      let sel = terminal.getSelection();
      if (typeof sel === 'string') {
        console.log(sel);
        setSelectedText(sel);
      }
    });

    terminal.onData((data) => {
      if (data.startsWith("\x1b[200~") && data.endsWith("\x1b[201~")) {
        ws.current.send(data);
      }
    });    

    terminal.attachCustomKeyEventHandler((e) => {
      const codes = {
        'Enter': '\r',
        'Backspace': '\x7f',
        'ArrowUp': '\x1b[A',
        'ArrowDown': '\x1b[B',
        'ArrowRight': '\x1b[C',
        'ArrowLeft': '\x1b[D',
        'Escape': '\x1b',
        'Home': '\x1b[H',
        'End': '\x1b[F',
        'Tab': '\t',
        'PageUp': '\x1b[5~',
        'PageDown': '\x1b[6~',
      };

      const cancelKeys = [
        'Shift',
        'Meta',
        'Alt',
        'Control',
        'CapsLock',
        'NumLock',
        'ScrollLock',
        'Pause',
      ]

      const codesCtrl = {
        'c': '\x03',
        'd': '\x04',
        'l': '\x0c',
        'z': '\x1a',
        'Backspace': '\x08',
      };
            
      if (e.type === 'keydown') {
        if (codesCtrl[e.key] && e.ctrlKey) {
          ws.current.send(codesCtrl[e.key]);
        } else if (codes[e.key]) {
          ws.current.send(codes[e.key]);
        } else if (cancelKeys.includes(e.key)) {
          return false;
        } else {
          ws.current.send(e.key);
        }
      } else if (e.type === 'keyup') {
      }

      return true;
    });   
  }, []);

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
      <div style={{
        overflowX: 'auto',
        width: '100%', 
        maxWidth: '900px', 
      }}>
        <div ref={xtermRef}></div>
        <br/>
      </div>

      <Stack 
        direction="row"
        spacing={1}
        alignItems="center"
        sx={{
          background: '#272d36',
          color: '#fff',
          padding: '10px',
          width: '100%',
        }}
      >
      <div>{
        isConnected ? (
          <ApiOutlined style={{color: '#00ff00', margin: '0px 5px', fontSize: '20px'}} />
        ) : (
          <ApiOutlined style={{color: '#ff0000', margin: '0px 5px', fontSize: '20px'}} />
        )
      }</div>
      
      {isConnected ? (<>
        <Button  variant="contained" onClick={() => ws.current.close()}>{t('mgmt.servapps.containers.terminal.disconnectButton')}</Button>
        <Button  variant="outlined" onClick={() => ws.current.send('\t')}>TAB</Button>
        <Button  variant="outlined" onClick={() => ws.current.send('\x03')}>Ctrl+C</Button>
      </>
      ) :
        <>  
          <Button variant="contained"
          onClick={() => connect(false)}>{t('mgmt.servapps.containers.terminal.connectButton')}</Button>
          <Button variant="contained" onClick={() => connect(true)}>{t('mgmt.servapps.containers.terminal.newShellButton')}</Button>
        </>
      }
      </Stack>
    </div>
  );
};

export default DockerTerminal;