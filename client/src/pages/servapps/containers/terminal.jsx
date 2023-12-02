import React, { useEffect, useRef, useState } from 'react';
import * as API from '../../../api';
import { Alert, Input, Stack, useMediaQuery, useTheme, Button, TextField } from '@mui/material';
import Terminal from '../../../components/terminal';
import { ApiOutlined, SendOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';

const DockerTerminal = ({containerInfo, refresh}) => {
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const isInteractive = Config.Tty;
  const theme = useTheme();
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'))
  const [history, setHistory] = useState([]);
  const [historyCursor, setHistoryCursor] = useState(0);

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
    if (e.key === 'Enter') {
      // shift + enter for new line
      if (e.shiftKey) {
        return;
      }
      e.preventDefault();
      sendMessage();
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (historyCursor > 0) {
        setHistoryCursor(historyCursor - 1);
        setMessage(history[historyCursor - 1]);
      }
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyCursor < history.length - 1) {
        setHistoryCursor(historyCursor + 1);
        setMessage(history[historyCursor + 1]);
      } else {
        setHistoryCursor(history.length);
        setMessage('');
      }
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
        let data = JSON.parse(event.data);
        setOutput((prevOutput) => [...prevOutput, ...data]);
      } catch (e) {
        console.error("error", e);
      }
    };

    ws.current.onclose = () => {
      setIsConnected(false);
      let terminalBoldRed = '\x1b[1;31m';
      setOutput((prevOutput) => [...prevOutput,
        {output: terminalBoldRed + 'Disconnected from ' + (newProc ? 'bash' : 'main process TTY'), type: 'stdout'}]);
    };
    
    ws.current.onopen = () => {
      setIsConnected(true);
      let terminalBoldGreen = '\x1b[1;32m';
      setOutput((prevOutput) => [...prevOutput, 
        {output: terminalBoldGreen + 'Connected to ' + (newProc ? 'bash' : 'main process TTY'), type: 'stdout'}]);
    };

    return () => {
      setIsConnected(false);
      ws.current.close();
    };
  };

  useEffect(() => {
  }, []);
  

  const sendMessage = () => {
    if (ws.current) {
      ws.current.send(message);
      setHistoryCursor(history.length + 1);
      setMessage('');
      setHistory((prevHistory) => [...prevHistory, message]);
    }
  };

  return (
    <div className="terminal-container" onKeyDown={handleKeyDown}>
      {(!isInteractive) && (
        <Alert severity="warning">
          This container is not interactive. 
          If you want to connect to the main process, 
          <Button onClick={() => makeInteractive()}>Enable TTY</Button>
        </Alert>
      )}
      
      <Terminal
        logs={output}
        setLogs={setOutput}
        docker
      />

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
      >
      <div style={{
        fontSize: '125%',
        padding: '10px 0',
      }}>
        {
          isConnected ? (
            <ApiOutlined style={{color: '#00ff00'}} />
          ) : (
            <ApiOutlined style={{color: '#ff0000'}} />
          )
        }
      </div>
      <Input
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        multiline
        fullWidth
        placeholder={isMobile ? "Enter command" : "Enter command (CTRL+Enter to send)"}
        style={{
          fontSize: '125%',
          padding: '10px 10px',
          background: 'rgba(0,0,0,0.1)',
        }}
        disableUnderline
        disabled={!isConnected}
      />
      
      <ResponsiveButton variant="outlined" disabled={!isConnected} onClick={sendMessage} endIcon={
        <SendOutlined />
      }>Send</ResponsiveButton>
      </Stack>


      <Stack 
        direction="row"
        spacing={1}
      >
      <Button variant="outlined"
       onClick={() => connect(false)}>Connect</Button>
      <Button variant="outlined" onClick={() => connect(true)}>New Shell</Button>
      {isConnected && (
        <Button  variant="outlined" onClick={() => ws.current.close()}>Disconnect</Button>
      )}
      </Stack>

      </Stack>
    </div>
  );
};

export default DockerTerminal;