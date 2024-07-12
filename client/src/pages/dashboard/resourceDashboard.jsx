import {
  Grid,
  LinearProgress,
} from '@mui/material';
import { useTranslation } from 'react-i18next';

import PlotComponent from './components/plot';
import TableComponent from './components/table';

const ResourceDashboard = ({ xAxis, zoom, setZoom, slot, metrics }) => {
  const { t } = useTranslation();
  return (<>

    <Grid container rowSpacing={4.5} columnSpacing={2.75} >
      <Grid item xs={12} md={7} lg={8}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('navigation.monitoring.resourcesTitle')} data={[metrics["cosmos.system.cpu.0"], metrics["cosmos.system.ram"]]} />
      </Grid>

      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('navigation.monitoring.resourceDashboard.averageResourcesTitle')} data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.system.docker.cpu") || key.startsWith("cosmos.system.docker.ram")).map((key) => metrics[key])
      } />

      <Grid item xs={12} md={7} lg={8}>
        <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('global.network')} data={[metrics["cosmos.system.netTx"], metrics["cosmos.system.netRx"]]} />
      </Grid>

      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('navigation.monitoring.resourceDashboard.averageNetworkTitle')} data={
        Object.keys(metrics).filter((key) => key.startsWith("cosmos.system.docker.net")).map((key) => metrics[key])
      } />

      <TableComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('navigation.monitoring.resourceDashboard.diskUsageTitle')} displayMax={true}
        render={(metric, value, formattedValue) => {
          let percent = value / metric.Max * 100;
          return <span>
            {formattedValue}
            <LinearProgress
              variant="determinate"
              color={percent > 95 ? 'error' : (percent > 75 ? 'warning' : 'info')}
              value={percent} />
          </span>
        }}
        data={
          Object.keys(metrics).filter((key) => key.startsWith("cosmos.system.disk.")).map((key) => metrics[key])
        } />

      <Grid item xs={12} md={7} lg={8}>
        <PlotComponent
          zoom={zoom} setZoom={setZoom}
          xAxis={xAxis}
          slot={slot}
          title={t('global.temperature')}
          withSelector={'cosmos.system.temp.all'}
          SimpleDesign
          data={Object.keys(metrics).filter((key) => key.startsWith("cosmos.system.temp")).map((key) => metrics[key])}
        />
      </Grid>
    </Grid>
  </>)
}

export default ResourceDashboard;