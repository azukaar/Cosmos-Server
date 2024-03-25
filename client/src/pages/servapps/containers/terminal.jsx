import React, { useEffect, useRef, useState } from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import * as API from '../../../api';
import { Alert, Input, Stack, useMediaQuery, useTheme } from '@mui/material';
import { ApiOutlined, SendOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';

import { Terminal, ITerminalOptions, ITerminalAddon } from 'xterm'
import 'xterm/css/xterm.css'

const DockerTerminal = ({containerInfo, refresh}) => {
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

  const handleKeyDown = (e) => {
    // if (e.key === 'Enter') {
    //   // shift + enter for new line
    //   if (e.shiftKey) {
    //     return;
    //   }
    //   e.preventDefault();
    //   sendMessage();
    // }
    // if (e.key === 'ArrowUp') {
    //   e.preventDefault();
    //   if (historyCursor > 0) {
    //     setHistoryCursor(historyCursor - 1);
    //     setMessage(history[historyCursor - 1]);
    //   }
    // }
    // if (e.key === 'ArrowDown') {
    //   e.preventDefault();
    //   if (historyCursor < history.length - 1) {
    //     setHistoryCursor(historyCursor + 1);
    //     setMessage(history[historyCursor + 1]);
    //   } else {
    //     setHistoryCursor(history.length);
    //     setMessage('');
    //   }
    // }

    console.log('e.target')

    // if terminal is focues 
    if (e.target === xtermRef.current) {
      // send char code to terminal with ws.current.send(charCode);
      ws.current.send(e.charCode);
    }
  };

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
      terminal.write(terminalBoldRed + 'Disconnected from ' + (newProc ? 'bash' : 'main process TTY') + '\r\n' + terminalReset);
    };
    
    ws.current.onopen = () => {
      setIsConnected(true);
      let terminalBoldGreen = '\x1b[1;32m';
      let terminalReset = '\x1b[0m';
      terminal.write(terminalBoldGreen + 'Connected to ' + (newProc ? 'bash' : 'main process TTY') + '\r\n' + terminalReset);
      // focus terminal
      xtermRef.current.focus();
    };

    return () => {
      setIsConnected(false);
      ws.current.close();
    };
  };

  const [SelectedText, setSelectedText] = useState('');

  useEffect(() => {
    xtermRef.current.innerHTML = '';
    terminal.open(xtermRef.current);

    if (navigator.clipboard) {
      navigator.permissions.query({ name: "clipboard-read" })
      navigator.permissions.query({ name: "clipboard-write" })
    } else {
      console.error('Clipboard API not available');
      return;
    }

    // remove all listener from xtermRef
    const contextMenuListener = async (e) => {
      e.preventDefault();
      console.log('context menu', SelectedText)
      if(SelectedText && SelectedText.length > 0 && SelectedText != "<empty string>") {
        await navigator.clipboard.writeText(SelectedText);
      } else {
        const toWrite = await navigator.clipboard.readText();
        ws.current.send(toWrite);
      }
      return false;
    }

    xtermRef.current.removeEventListener('contextmenu', contextMenuListener);

    xtermRef.current.addEventListener('contextmenu', contextMenuListener);

    terminal.onSelectionChange((e) => {
      let sel = terminal.getSelection();
      if (typeof sel === 'string') {
        console.log(sel);
        setSelectedText(sel);
      }
    });

    terminal.attachCustomKeyEventHandler((e) => {
      // only keys!!
      // if touch, bring up keyboard
      // if (e.type === 'touchstart') {
      //   xtermRef.current.focus();
      // }

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
    <div className="terminal-container" onKeyDown={handleKeyDown}>
      {(!isInteractive) && (
        <Alert severity="warning">
          This container is not interactive. 
          If you want to connect to the main process, 
          <Button onClick={() => makeInteractive()}>Enable TTY</Button>
        </Alert>
      )}
      
      <div ref={xtermRef}></div>

      <Stack 
        direction="column"
        spacing={1}
      >

      <Stack 
        direction="row"
        spacing={1}
        sx={{
          background: '#272d36',
          color: '#fff',
          padding: '10px',
        }}
        alignItems="center"
      >
      <div>{
        isConnected ? (
          <ApiOutlined style={{color: '#00ff00', margin: '0px 5px', fontSize: '20px'}} />
        ) : (
          <ApiOutlined style={{color: '#ff0000', margin: '0px 5px', fontSize: '20px'}} />
        )
      }</div>
      
      {isConnected ? (<>
        <Button  variant="contained" onClick={() => ws.current.close()}>Disconnect</Button>
        <Button  variant="outlined" onClick={() => ws.current.send('\t')}>TAB</Button>
        <Button  variant="outlined" onClick={() => ws.current.send('\x03')}>Ctrl+C</Button>
      </>
      ) :
        <>  
          <Button variant="contained"
          onClick={() => connect(false)}>Connect</Button>
          <Button variant="contained" onClick={() => connect(true)}>New Shell</Button>
        </>
      }
      </Stack>

      </Stack>
    </div>
  );
};

export default DockerTerminal;