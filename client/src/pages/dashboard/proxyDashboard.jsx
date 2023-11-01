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
      <Grid item xs={12} md={7} lg={8}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={'Requests'} data={[
          metrics["cosmos.proxy.all.time"],
          metrics["cosmos.proxy.all.success"],
          metrics["cosmos.proxy.all.error"],
        ]} />
      </Grid>

      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title="Containers - Resources" data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.proxy.route.")).map((key) => metrics[key])
      } />

    </Grid>
  </>)
}

export default ProxyDashboard;