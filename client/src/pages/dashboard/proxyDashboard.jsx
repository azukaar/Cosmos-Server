import {
  Grid,
  LinearProgress,
} from '@mui/material';

import PlotComponent from './components/plot';
import TableComponent from './components/table';

const ProxyDashboard = ({ xAxis, zoom, setZoom, slot, metrics }) => {
  console.log(metrics)
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


      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title="Reasons For Blocked Requests" data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.proxy.blocked.")).map((key) => metrics[key])
      } />
    </Grid>
  </>)
}

export default ProxyDashboard;