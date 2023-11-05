import {
  Grid,
  LinearProgress,
  Tooltip,
} from '@mui/material';

import PlotComponent from './components/plot';
import TableComponent from './components/table';
import { InfoCircleOutlined } from '@ant-design/icons';

const ProxyDashboard = ({ xAxis, zoom, setZoom, slot, metrics }) => {
  return (<>

    <Grid container rowSpacing={4.5} columnSpacing={2.75} >
      <Grid item xs={12} md={6} lg={6}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={'Requests Resources'} data={[
          metrics["cosmos.proxy.all.time"],
          metrics["cosmos.proxy.all.bytes"],
        ]} />
      </Grid>

      <Grid item xs={12} md={6} lg={6}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={'Requests Responses'} data={[
          metrics["cosmos.proxy.all.success"],
          metrics["cosmos.proxy.all.error"],
        ]} />
      </Grid>

      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title="Requests Per URLs" data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.proxy.route.")).map((key) => metrics[key])
      } />

      <Grid item xs={12} md={4} lg={4}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={'Blocked Requests'} data={[
          metrics["cosmos.proxy.all.blocked"],
        ]} />
      </Grid>


      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={
        <span>
          Reasons For Blocked Requests <Tooltip title={<div>
            <div><strong>bots</strong>: Bots</div>
            <div><strong>geo</strong>: By Geolocation (blocked countries)</div>
            <div><strong>referer</strong>: By Referer</div>
            <div><strong>hostname</strong>: By Hostname (usually IP scanning threat)</div>
            <div><strong>ip-whitelists</strong>: By IP Whitelists (Including restricted to Constellation)</div>
            <div><strong>smart-shield</strong>: Smart Shield (various abuse metrics such as time, size, brute-force, concurrent requests, etc...). It does not include blocking for banned IP to save resources in case of potential attacks</div>
          </div>}><InfoCircleOutlined /></Tooltip>
        </span>} data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.proxy.blocked.")).map((key) => metrics[key])
      } />
    </Grid>
  </>)
}

export default ProxyDashboard;