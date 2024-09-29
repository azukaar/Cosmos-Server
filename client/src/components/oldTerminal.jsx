import React, { useState, useEffect, useRef } from 'react';
import { Box, Button, Checkbox, CircularProgress, Input, Stack, TextField, Typography, useMediaQuery } from '@mui/material';
import * as API from '../api';
import LogLine from '../components/logLine';
import { useTheme } from '@emotion/react';

const Terminal = ({ logs, setLogs, fetchLogs, docker }) => {
  const [hasMore, setHasMore] = useState(true);
  const [hasScrolled, setHasScrolled] = useState(false);
  const [fetching, setFetching] = useState(false);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const screenMin = useMediaQuery((theme) => theme.breakpoints.up('sm'))

  const bottomRef = useRef(null);
  const terminalRef = useRef(null);

  const scrollToBottom = () => {
    bottomRef.current.scrollIntoView({});
  };

  useEffect(() => {
    if (!hasScrolled) {
      scrollToBottom();
    }
  }, [logs]);

  const handleScroll = (event) => {
    if (event.target.scrollHeight - event.target.scrollTop === event.target.clientHeight) {
      setHasScrolled(false);
    }else {
      setHasScrolled(true);
    }
  };

  return (
      <Box
        ref={terminalRef}
        sx={{
          minHeight: '50px',
          maxHeight: 'calc(1vh * 80 - 200px)',
          overflow: 'auto',
          padding: '10px',
          wordBreak: 'break-all',
          background: '#272d36',
          color: '#fff',
          borderTop: '3px solid ' + theme.palette.primary.main
        }}
        onScroll={handleScroll}
      >
        {logs && logs.map((log, index) => (
          <div key={index} style={{paddingTop: (!screenMin) ? '10px' : '2px'}}>
            <LogLine message={log.output} docker isMobile={!screenMin} />
          </div>
        ))}
        {fetching && <CircularProgress sx={{ mt: 1, mb: 2 }} />}
        <div ref={bottomRef} />
      </Box>
  );
};

export default Terminal;