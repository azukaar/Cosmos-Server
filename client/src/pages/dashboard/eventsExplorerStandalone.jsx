import { useEffect, useRef, useState } from 'react';
import localizedFormat from 'dayjs/plugin/localizedFormat'; // import this for localized formatting
import 'dayjs/locale/en-gb';

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
dayjs.locale('en-gb');

const EventExplorerStandalone = ({initSearch, initLevel}) => {
  // one hour ago
  const now = dayjs();
  const [from, setFrom] = useState(now.subtract(1, 'hour'));
  const [to, setTo] = useState(now)

  return (
      <>
      <div style={{zIndex:2, position: 'relative'}}>
          <Grid container rowSpacing={4.5} columnSpacing={2.75} >
              <Grid item xs={12} sx={{ mb: -2.25 }}>
                  <Typography variant="h4">Events</Typography>
                  
                  <Stack direction="row" spacing={2} sx={{ mt: 1.5 }}>
                    <DateTimePicker label="From" value={from} onChange={(e) => setFrom(e)} />
                    <DateTimePicker label="To" value={to} onChange={(e) => setTo(e)} />
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
