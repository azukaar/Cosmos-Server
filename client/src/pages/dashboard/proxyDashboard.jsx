import {
  Grid,
  LinearProgress,
  Tooltip,
} from '@mui/material';
import { useTranslation } from 'react-i18next';

import PlotComponent from './components/plot';
import TableComponent from './components/table';
import { InfoCircleOutlined } from '@ant-design/icons';

const ProxyDashboard = ({ xAxis, zoom, setZoom, slot, metrics }) => {
  const { t } = useTranslation();
  return (<>

    <Grid container rowSpacing={4.5} columnSpacing={2.75} >
      <Grid item xs={12} md={6} lg={6}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('RequestsResources')} data={[
          metrics["cosmos.proxy.all.time"],
          metrics["cosmos.proxy.all.bytes"],
        ]} />
      </Grid>

      <Grid item xs={12} md={6} lg={6}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('RequestsResponses')} data={[
          metrics["cosmos.proxy.all.success"],
          metrics["cosmos.proxy.all.error"],
        ]} />
      </Grid>

      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('RequestsPerURLs')} data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.proxy.route.")).map((key) => metrics[key])
      } />

      <Grid item xs={12} md={4} lg={4}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('BlockedRequests')} data={[
          metrics["cosmos.proxy.all.blocked"],
        ]} />
      </Grid>


      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={
        <span>
          {t('ReasonBlocked')} <Tooltip title={<div>
            <div><strong>bots</strong>: {t('ByBots')}</div>
            <div><strong>geo</strong>: {t('ByGeo')}</div>
            <div><strong>referer</strong>: {t('ByReferer')}</div>
            <div><strong>hostname</strong>: {t('ByHostname')}</div>
            <div><strong>ip-whitelists</strong>: {t('ByIPWhitelists')}</div>
            <div><strong>smart-shield</strong>: {t('BySmartShield')}</div>
          </div>}><InfoCircleOutlined /></Tooltip>
        </span>} data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.proxy.blocked.")).map((key) => metrics[key])
      } />
    </Grid>
  </>)
}

export default ProxyDashboard;