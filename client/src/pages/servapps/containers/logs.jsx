import React, { useState, useEffect, useRef } from 'react';
import { Box, Button, Checkbox, CircularProgress, Input, Stack, TextField, Typography, useMediaQuery } from '@mui/material';
import * as API from '../../../api';
import LogLine from '../../../components/logLine';
import { useTheme } from '@emotion/react';
import { useTranslation } from 'react-i18next';

const Logs = ({ containerInfo }) => {
  const { t } = useTranslation();
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const containerName = Name;
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [errorOnly, setErrorOnly] = useState(false);
  const [limit, setLimit] = useState(100);
  const [searchTerm, setSearchTerm] = useState('');
  const [page, setPage] = useState('');
  const [hasMore, setHasMore] = useState(true);
  const [hasScrolled, setHasScrolled] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [forceUpdate, setForceUpdate] = useState(false);
  const [lastReceivedLogs, setLastReceivedLogs] = useState('');
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [scrollToMe, setScrollToMe] = useState(null);
  const screenMin = useMediaQuery((theme) => theme.breakpoints.up('sm'))

  const bottomRef = useRef(null);
  const topRef = useRef(null);
  const terminalRef = useRef(null);

  const scrollToBottom = () => {
    bottomRef.current.scrollIntoView({ behavior: 'smooth' });
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
        // const current = topRef.current;
        // setScrollToMe(() => current);
        setLogs((logs) => [...data, ...logs]);
        // calculate the height of the new logs and scroll to that position
        // OK I will fix this later
        // const newHeight = 999999999;
        // terminalRef.current.scrollTop = newHeight;
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
    } else {
      // scrollToMe && scrollToMe.scrollIntoView({ });
      // setScrollToMe(null);
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
          <Stack direction="row" spacing={3}>
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
          <Button
            variant="outlined"
            onClick={() => {
              setHasScrolled(false);
              setLastReceivedLogs('');
              fetchLogs(true, true);
            }}
          >
            {t('global.refresh')}
          </Button>
        </Stack>
        </Stack>
        {loading && <CircularProgress />}
      </Stack>
      <Box
        ref={terminalRef}
        sx={{
          maxHeight: 'calc(1vh * 80 - 200px)',
          overflow: 'auto',
          paddingRight: '10px',
          paddingLeft: '10px',
          paddingBottom: '10px',
          wordBreak: 'break-all',
          background: '#272d36',
          color: '#fff',
          borderTop: '3px solid ' + theme.palette.primary.main
        }}
        onScroll={handleScroll}
      >
        {logs.map((log, index) => (
          <div key={log.index} style={{paddingTop: (!screenMin) ? '10px' : '2px'}}>
            <LogLine message={log.output} docker isMobile={!screenMin} />
          </div>
        ))}
        {fetching && <CircularProgress sx={{ mt: 1, mb: 2 }} />}
        <div ref={bottomRef} />
      </Box>
    </Stack>
  );
};

export default Logs;