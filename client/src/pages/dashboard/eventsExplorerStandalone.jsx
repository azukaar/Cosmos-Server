import { useEffect, useRef, useState } from 'react';
import localizedFormat from 'dayjs/plugin/localizedFormat'; // import this for localized formatting
import { useTranslation } from 'react-i18next';

// material-ui
import {
    Grid,
    Stack,
    Typography,
} from '@mui/material';


import EventsExplorer from './eventsExplorer';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';

import dayjs from 'dayjs';

dayjs.extend(localizedFormat); // if needed

const EventExplorerStandalone = ({initSearch, initLevel}) => {
  const { t } = useTranslation();
  // one hour ago
  const now = dayjs();
  const [from, setFrom] = useState(now.subtract(1, 'hour'));
  const [to, setTo] = useState(now)

  return (
      <>
      <div style={{zIndex:2, position: 'relative'}}>
          <Grid container rowSpacing={4.5} columnSpacing={2.75} >
              <Grid item xs={12} sx={{ mb: -2.25 }}>
                  <Typography variant="h4">{t('navigation.monitoring.eventsTitle')}</Typography>
                  
                  <Stack direction="row" spacing={2} sx={{ mt: 1.5 }}>
                    <DateTimePicker label={t('navigation.monitoring.events.datePicker.fromLabel')} value={from} onChange={(e) => setFrom(e)} />
                    <DateTimePicker label={t('navigation.monitoring.events.datePicker.toLabel')} value={to} onChange={(e) => setTo(e)} />
                  </Stack>
              </Grid>


              <Grid item xs={12} md={12} lg={12}>
                <EventsExplorer initLevel={initLevel} initSearch={initSearch} from={from} to={to}/>
              </Grid>
          </Grid>
      </div>
    </>
  );
};

export default EventExplorerStandalone;
