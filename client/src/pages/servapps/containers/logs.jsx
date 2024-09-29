import React, { useState, useEffect, useRef } from 'react';
import { Box, Button, Checkbox, CircularProgress, Input, Stack, TextField, Tooltip, Typography, useMediaQuery } from '@mui/material';
import * as API from '../../../api';
import LogLine from '../../../components/logLine';
import { useTheme } from '@emotion/react';
import { useTranslation } from 'react-i18next';
import ResponsiveButton from '../../../components/responseiveButton';
import { ArrowDownOutlined, SyncOutlined } from '@ant-design/icons';
import { DownloadFile } from '../../../api/downloadButton';

const Logs = ({ containerInfo }) => {
  const { t } = useTranslation();
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const containerName = Name;
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [errorOnly, setErrorOnly] = useState(false);
  const [limit, setLimit] = useState(200);
  const [searchTerm, setSearchTerm] = useState('');
  const [hasMore, setHasMore] = useState(true);
  const [hasScrolled, setHasScrolled] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [lastReceivedLogs, setLastReceivedLogs] = useState('');
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const screenMin = useMediaQuery((theme) => theme.breakpoints.up('sm'))
  const pingInterval = useRef(null);
  const [message, setMessage] = useState('');
  const [output, setOutput] = useState([
    {
      output: 'Not Connected.',
      type: 'stdout'
    }
  ]);
  const ws = useRef(null);
  const [isConnected, setIsConnected] = useState(false);

  const bottomRef = useRef(null);
  const terminalRef = useRef(null);
  const oldScrollHeightRef = useRef(0);
  const oldScrollTopRef = useRef(0);
  
  const connect = () => {
    let wasFirst = true;

    if(ws.current) {
      ws.current.close();
    }

    ws.current = API.docker.attachTerminal(Name.slice(1));

    ws.current.onmessage = (event) => {
      if(event.data === '_PONG_') {
        return;
      }

      try {
        let newLogs = event.data;
        if (newLogs) {
          let isFirstLine = true;
          newLogs = newLogs.split('\n')
            .map((log, index) => {
              if (log && log.length > 0) {
                if (isFirstLine && !wasFirst) {
                  console.log('YES 1');
                  isFirstLine = false;
                  return {
                    output: log,
                    needAppend: true,
                  };
                } else {
                  return {
                    output: (new Date()).toISOString() + ' ' + log,
                  };
                }
              }
              return {
                output: "",
              }; // Explicitly return null for empty logs
            })

          // if last websocket was a clean cut, it needs to not append
          if(newLogs.length > 2) {
            if(newLogs[0].output === '' && newLogs[1].needAppend) {
              newLogs = newLogs.slice(1);
              newLogs[0].needAppend = false;
              // it needs the date
              if(!newLogs[0].output.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z/)) {
                newLogs[0].output = (new Date()).toISOString() + ' ' + newLogs[0].output;
              }
            }
          }
            
          if (newLogs.length > 0) {
            console.log('newLogs', newLogs);
            (_newLogs => {
              setLogs((prevLogs) => {
                console.log(prevLogs.length > 0, _newLogs.length > 0, _newLogs[0].needAppend)
                if (prevLogs.length > 0 && _newLogs.length > 0 && _newLogs[0].needAppend) {
                  console.log('_newLogs', _newLogs);
                  // if missing date, add 
                  if(!_newLogs[0].output.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z/)) {
                    prevLogs[prevLogs.length - 1].output += (new Date()).toISOString() + ' ' + _newLogs[0].output;
                  } else {
                    prevLogs[prevLogs.length - 1].output += _newLogs[0].output;
                  }
                  _newLogs = _newLogs.slice(1);  // Create a new array without the first element
                }
                wasFirst = false;
                return [...prevLogs, ..._newLogs];
              });
            })(newLogs);
          }
        }
      } catch (error) {
        console.error('Error processing logs:', error);
      }
    };

    ws.current.onclose = () => {
      setIsConnected(false);
      let terminalBoldRed = '\x1b[1;31m';
      let terminalReset = '\x1b[0m';
      if(pingInterval.current)
        clearInterval(pingInterval.current);
    };
    
    ws.current.onopen = () => {
      setIsConnected(true);
      let terminalBoldGreen = '\x1b[1;32m';
      let terminalReset = '\x1b[0m';

      pingInterval.current = setInterval(() => {
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
          ws.current.send('_PING_');
          console.log('Sent PING');
        }
      }, 30000); // Send a ping every 30 seconds
    };

    return () => {
      setIsConnected(false);
      if(pingInterval.current)
        clearInterval(pingInterval.current);
      ws.current.close();
    };
  };

  useEffect(() => {
    return connect();
  }, [containerInfo]);

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const fetchLogs = async (reset, ignoreState) => {
    setLoading(true);
    try {
      const response = await API.docker.getContainerLogs(
        containerName,
        searchTerm,
        limit,
        ignoreState ? '' : lastReceivedLogs,
        errorOnly
      );
      const { data } = response;
      if (data.length > 0) {
        const date = data[0].output.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z/)[0];
        if (date) {
          date.replace(/(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.)(\d+)Z/, (match, p1, p2) => {
            const newNumber = parseInt(p2) - 1;
            const newDate = `${p1}${newNumber}Z`;
            setLastReceivedLogs(newDate);
          });
        } else {
          console.error('Could not parse date from log: ', data[0]);
          setLastReceivedLogs('');
        }
      }
      if(reset) {
        setLogs(data);
      } else {
        oldScrollHeightRef.current = terminalRef.current.scrollHeight;
        oldScrollTopRef.current = terminalRef.current.scrollTop;
        setLogs((prevLogs) => [...data, ...prevLogs]);
      }
      setHasMore(true);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
      setFetching(false);
    }
  };

  useEffect(() => {
    fetchLogs(true);
  }, [searchTerm, errorOnly, limit]);

  useEffect(() => {
    if (!fetching) return;
    fetchLogs();
  }, [fetching]);

  useEffect(() => {
    if (!hasScrolled) {
      scrollToBottom();
    } else if (terminalRef.current) {
      const newScrollHeight = terminalRef.current.scrollHeight;
      const heightDifference = newScrollHeight - oldScrollHeightRef.current;
      terminalRef.current.scrollTop = oldScrollTopRef.current + heightDifference;
    }
  }, [logs]);

  const handleScroll = (event) => {
    const { scrollTop } = event.target;
    setHasScrolled(true);
    if (scrollTop === 0) {
      if(!hasMore) return;
      setFetching(true);
      setHasMore(false);
    } else {
      setHasMore(true);
    }
  };

  return (
    <Stack spacing={2} sx={{  width: '100%'}}>
      <Stack
        spacing={2}
        direction="column"
      >
        <Stack direction={screenMin ? 'row' : 'column'} spacing={3}>
          <Stack direction="row" spacing={3}>
          <Input
            label={t('global.searchPlaceholder')}
            value={searchTerm}
            placeholder={t('global.searchPlaceholder')+"..."}
            onChange={(e) => {
              setHasScrolled(false);
              setSearchTerm(e.target.value);
              setLastReceivedLogs('');
            }}
          />
          <Box>
            <Checkbox
              checked={errorOnly}
              onChange={(e) => {
                setHasScrolled(false);
                setErrorOnly(e.target.checked);
                setLastReceivedLogs('');
              }}
            />
            {t('mgmt.servApps.container.protocols.errorOnlyCheckbox')}
          </Box>
        </Stack>
        <Stack direction="row" spacing={3} alignItems="center">
          <Box>
            <TextField
              label="Limit"
              type="number"
              value={limit}
              onChange={(e) => {
                setHasScrolled(false);
                setLimit(e.target.value);
                setLastReceivedLogs('');
              }}
            />
          </Box>
          <ResponsiveButton
            startIcon={<SyncOutlined />}
            variant="outlined"
            onClick={() => {
              setHasScrolled(false);
              setLastReceivedLogs('');
              fetchLogs(true, true);
            }}
          >
            {t('global.refresh')}
          </ResponsiveButton>
          <DownloadFile
            filename={`${(new Date()).toISOString()} - ${containerName} - logs.txt`}
            content={logs.map((log) => log.output).join('\n')}
            label={t('global.downloadLogs')}
          />
          {isConnected ? (
            <Tooltip title={t('mgmt.servApps.container.terminal.connected')}>
              <div style={{ background: '#00ff00', borderRadius:'15px', width: '15px', height: '15px' }}></div>
            </Tooltip>
          ) : (
            <>
              <Tooltip title={t('mgmt.servApps.container.terminal.connected')}>
                <div style={{ background: 'red', borderRadius:'15px', width: '15px', height: '15px' }}></div>
              </Tooltip>
              <ResponsiveButton
                variant="contained"
                onClick={connect}
              >
                {t('mgmt.servApps.container.terminal.reconnectButton')}
              </ResponsiveButton>
            </>
          )}
        </Stack>
        </Stack>
        {loading && <CircularProgress />}
      </Stack>
      <Box
        ref={terminalRef}
        sx={{
          maxHeight: 'calc(1vh * 80 - 200px)',
          overflow: 'auto',
          wordBreak: 'break-all',
          background: '#272d36',
          color: '#fff',
          borderTop: '3px solid ' + theme.palette.primary.main
        }}
        onScroll={handleScroll}
      >
        {logs.map((log, index) => (
          <Box key={index + log.output} sx={{
            paddingTop: (!screenMin) ? '6px' : '2px',
            paddingBottom: (!screenMin) ? '6px' : '2px',
            paddingLeft: '10px',
            background: index % 2 === 0 ? '#272d36' : '#1e222c',
            '&:hover': {
              background: '#222244',
            }
          }}>
            <LogLine message={log.output} docker isMobile={!screenMin} />
          </Box>
        ))}
        {fetching && <CircularProgress sx={{ mt: 1, mb: 2 }} />}
        <div ref={bottomRef} />
      </Box>
    </Stack>
  );
};

export default Logs;