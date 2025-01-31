import React, { useEffect, useRef } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button } from '@mui/material';
import * as API from '../../api';
import { useTranslation } from 'react-i18next';

const TerminalComponent = ({ job }) => {
  const terminalRef = useRef(null);
  const xtermRef = useRef(null);
  const fitAddonRef = useRef(null);

  useEffect(() => {
    if (job && terminalRef.current) {
      // Initialize XTerm
      const terminal = new Terminal({
        theme: {
          background: '#000000',
          foreground: '#ffffff',
          cursor: '#ffffff',
          selection: '#333333',
        },
        fontSize: 12,
        fontFamily: 'monospace',
        cursorBlink: true,
        convertEol: true,
        scrollback: 1000,
      });

      // Create and load addons
      const fitAddon = new FitAddon();
      terminal.loadAddon(fitAddon);

      // Save references
      xtermRef.current = terminal;
      fitAddonRef.current = fitAddon;

      // Open terminal in the container
      terminal.open(terminalRef.current);
      fitAddon.fit();

      // Fetch and display logs
      API.cron.get(job.Scheduler, job.Name).then((res) => {
        if (res.data && res.data.Logs) {
          res.data.Logs.forEach((log) => {
            terminal.writeln(log);
          });
        }
      });

      // Handle window resize
      const handleResize = () => {
        fitAddon.fit();
      };
      window.addEventListener('resize', handleResize);

      // Cleanup
      return () => {
        window.removeEventListener('resize', handleResize);
        terminal.dispose();
      };
    }
  }, [job]);

  return (
    <div
      ref={terminalRef}
      style={{
        height: '520px',
        backgroundColor: '#000',
        padding: '10px',
        borderRadius: '5px',
      }}
    />
  );
};

const JobLogsDialog = ({ job, OnClose }) => {
  const { t } = useTranslation();

  return (
    <Dialog
      open={Boolean(job)}
      onClose={OnClose}
      maxWidth="md"
      fullWidth
    >
      <DialogTitle>{t('mgmt.scheduler.lastLogs')} {job?.Name}</DialogTitle>
      <DialogContent>
        <TerminalComponent job={job} />
      </DialogContent>
      <DialogActions>
        <Button onClick={OnClose}>{t('global.close')}</Button>
      </DialogActions>
    </Dialog>
  );
};

export default JobLogsDialog;