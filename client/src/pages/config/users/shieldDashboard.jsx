import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import {
  Button,
  Chip,
  Grid,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material';
import { SyncOutlined, DeleteOutlined, InfoCircleOutlined } from '@ant-design/icons';

import * as API from '../../../api';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { ConfirmModalDirect } from '../../../components/confirmModal';
import PermissionGuard from '../../../components/permissionGuard';
import { PERM_CONFIGURATION } from '../../../utils/permissions';
import PlotComponent from '../../dashboard/components/plot';
import TableComponent from '../../dashboard/components/table';
import MetricHeaders from '../../dashboard/MetricHeaders';
import { formatDate, simplifyNumber } from '../../dashboard/components/utils';

const statusColor = {
  perm: 'error',
  temp: 'error',
  strike: 'warning',
  clear: 'default',
};

const describeReason = (t, reason) => {
  if (!reason) return '';
  if (reason.limit === 'escalation') return t('mgmt.shield.bans.reason.escalation');
  const fmt = (v) => reason.limit === 'bytes' ? simplifyNumber(v, 'B') : Math.round(v);
  return t('mgmt.shield.bans.reason.limit', {
    limit: reason.limit,
    used: fmt(reason.used),
    allowed: fmt(reason.allowed),
    route: reason.route,
  });
};

const History = ({ t, history }) => (
  <Stack spacing={0.5}>
    {history.map((ban) => (
      <Typography key={ban.id} variant="body2" style={{ whiteSpace: 'nowrap' }}>
        {dayjs(ban.time).format('L LT')}: <strong>{t('mgmt.shield.bans.type.' + ban.banType)}</strong>
        {' '}{describeReason(t, ban.reason)}
        {ban.node && <span style={{ opacity: 0.6 }}> ({t('mgmt.shield.bans.node', { node: ban.node })})</span>}
      </Typography>
    ))}
  </Stack>
);

const ShieldDashboard = () => {
  const { t } = useTranslation();
  const [slot, setSlot] = useState('latest');
  const [zoom, setZoom] = useState({ xaxis: {} });
  const [metrics, setMetrics] = useState(null);
  const [clients, setClients] = useState(null);
  const [confirmUnban, setConfirmUnban] = useState(null);
  const timer = useRef(null);

  const refreshMetrics = () => {
    API.metrics.get([
      'cosmos.proxy.all.blocked',
      'cosmos.proxy.blocked.*',
    ]).then((res) => {
      let finalMetrics = {};
      if (res.data) {
        res.data.forEach((metric) => {
          finalMetrics[metric.Key] = metric;
        });
      }
      setMetrics(finalMetrics);
    });
  };

  const refreshBans = () => {
    API.shield.bans().then((res) => {
      setClients(res.data || []);
    });
  };

  const refresh = () => {
    refreshMetrics();
    refreshBans();
  };

  useEffect(() => {
    refresh();
    timer.current = setInterval(refresh, 10000);
    return () => clearInterval(timer.current);
  }, []);

  let xAxis = [];
  if (slot === 'latest') {
    for (let i = 0; i < 100; i++) xAxis.unshift(i);
  } else if (slot === 'hourly') {
    for (let i = 0; i < 48; i++) {
      let now = new Date();
      now.setHours(now.getHours() - i);
      now.setMinutes(0);
      now.setSeconds(0);
      xAxis.unshift(formatDate(now, true));
    }
  } else if (slot === 'daily') {
    for (let i = 0; i < 30; i++) {
      let now = new Date();
      now.setDate(now.getDate() - i);
      xAxis.unshift(formatDate(now));
    }
  }

  return <div style={{ maxWidth: '1200px', margin: 'auto' }}>
    {confirmUnban && <ConfirmModalDirect
      callback={() => {
        API.shield.unban(confirmUnban.clientID).then(() => refreshBans());
      }}
      content={t('mgmt.shield.bans.unbanConfirm', { client: confirmUnban.clientID })}
      onClose={() => setConfirmUnban(null)}
    />}

    <Grid container rowSpacing={4.5} columnSpacing={2.75}>
      <Grid item xs={12}>
        <MetricHeaders loaded={metrics} slot={slot} setSlot={setSlot} zoom={zoom} setZoom={setZoom} />
      </Grid>

      {metrics && <>
        <Grid item xs={12} md={7} lg={8}>
          <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('mgmt.shield.metrics.blockedTitle')} data={[
            metrics['cosmos.proxy.all.blocked'],
          ]} />
        </Grid>
        <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={
          <span>
            {t('mgmt.shield.metrics.reasonsTitle')} <Tooltip title={<div>
              <div><strong>bots</strong>: {t('navigation.monitoring.resourceDashboard.reasonByBots')}</div>
              <div><strong>geo</strong>: {t('navigation.monitoring.resourceDashboard.reasonByGeo')}</div>
              <div><strong>referer</strong>: {t('navigation.monitoring.resourceDashboard.reasonByRef')}</div>
              <div><strong>hostname</strong>: {t('navigation.monitoring.resourceDashboard.reasonByHostname')}</div>
              <div><strong>ip-whitelists</strong>: {t('navigation.monitoring.resourceDashboard.reasonByWhitelist')}</div>
              <div><strong>smart-shield</strong>: {t('navigation.monitoring.resourceDashboard.reasonBySmartShield')}</div>
            </div>}><InfoCircleOutlined /></Tooltip>
          </span>} data={
          Object.keys(metrics).filter((key) => key.startsWith('cosmos.proxy.blocked.')).map((key) => metrics[key])
        } />
      </>}

      <Grid item xs={12}>
        <Typography variant="h5" style={{ marginBottom: '10px' }}>{t('mgmt.shield.bans.title')}</Typography>
        <PrettyTableView
          isLoading={clients === null}
          data={clients || []}
          getKey={(c) => c.clientID}
          buttons={[
            <Button key="refresh" variant="outlined" startIcon={<SyncOutlined />} onClick={refresh}>{t('global.refresh')}</Button>,
          ]}
          columns={[
            {
              title: t('mgmt.shield.bans.client'),
              search: (c) => c.clientID,
              field: (c) => <strong>{c.clientID}</strong>,
            },
            {
              title: t('mgmt.shield.bans.status'),
              field: (c) => <Chip size="small" color={statusColor[c.status] || 'default'} label={t('mgmt.shield.bans.status.' + c.status)} />,
            },
            {
              title: t('mgmt.shield.bans.until'),
              screenMin: 'md',
              field: (c) => c.status === 'perm' ? '∞' : (c.blockedUntil ? dayjs(c.blockedUntil).format('L LT') : ''),
            },
            {
              title: t('mgmt.shield.bans.lastReason'),
              screenMin: 'md',
              search: (c) => c.history.map((b) => b.reason && b.reason.route).join(' '),
              field: (c) => describeReason(t, c.history[c.history.length - 1].reason),
            },
            {
              title: t('mgmt.shield.bans.history'),
              screenMin: 'lg',
              field: (c) => <Tooltip title={<History t={t} history={c.history} />}>
                <Chip size="small" variant="outlined" label={c.history.length} />
              </Tooltip>,
            },
            {
              title: '',
              field: (c) => <PermissionGuard permission={PERM_CONFIGURATION}>
                <Button size="small" variant="outlined" color="error" startIcon={<DeleteOutlined />} onClick={() => setConfirmUnban(c)}>
                  {c.status === 'strike' || c.status === 'clear' ? t('mgmt.shield.bans.unstrike') : t('mgmt.shield.bans.unban')}
                </Button>
              </PermissionGuard>,
              style: { textAlign: 'right' },
            },
          ]}
        />
      </Grid>
    </Grid>
  </div>;
};

export default ShieldDashboard;
