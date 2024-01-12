// material-ui
import { CloseSquareOutlined, DeleteOutlined, PlusCircleOutlined, SyncOutlined } from '@ant-design/icons';
import { Button, Chip, CircularProgress, Stack, useTheme } from '@mui/material';
import { useEffect, useState } from 'react';

import * as API from '../../api';
import PrettyTableView from '../../components/tableView/prettyTableView';
import NewNetworkButton from './createNetwork';

export const NetworksColumns = (theme, isDark) => [
  {
    title: 'Network Name',
    field: (r) => <Stack direction='column'>
      <div style={{display:'inline-block', textDecoration: 'inherit', fontSize:'125%', color: isDark ? theme.palette.primary.light : theme.palette.primary.dark}}>{r.Name}</div><br/>
      <div style={{display:'inline-block', textDecoration: 'inherit', fontSize: '90%', opacity: '90%'}}>{r.Driver} driver</div>
    </Stack>,
    search: (r) => r.Name,
  },
  {
    title: 'Properties',
    screenMin: 'md',
    field: (r) => (
      <Stack direction="row" spacing={1}>
        <Chip label={r.Scope} color="primary" />
        {r.Internal && <Chip label="Internal" color="secondary" />}
        {r.Attachable && <Chip label="Attachable" color="success" />}
        {r.Ingress && <Chip label="Ingress" color="warning" />}
      </Stack>
    ),
  },
  {
    title: 'IPAM gateway / mask',
    screenMin: 'lg',
    field: (r) => r.IPAM.Config.map((config, index) => (
      <Stack key={index}>
        <div>{config.Gateway}</div>
        <div>{config.Subnet}</div>
      </Stack>
    )),
  },
  {
    title: 'Created At',
    screenMin: 'lg',
    field: (r) => new Date(r.Created).toLocaleString(),
  },
];

const NetworkManagementList = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [rows, setRows] = useState(null);
  const [tryDelete, setTryDelete] = useState(null);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  function refresh() {
    setIsLoading(true);
    API.docker.networkList()
      .then(data => {
        setRows(data.data);
        setIsLoading(false);
      });
  }

  useEffect(() => {
    refresh();
  }, [])

  return (
    <>
      <Stack direction='row' spacing={1} style={{ marginBottom: '20px' }}>
        <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={refresh}>
          Refresh
        </Button>
      </Stack>

      {rows && (
        <PrettyTableView
          data={rows}
          isLoading={isLoading}
          buttons={[
            <NewNetworkButton refresh={refresh} />,
          ]}
          onRowClick={() => { }}
          getKey={(r) => r.Id}
          columns={[
            ...NetworksColumns(theme, isDark),
            {
              title: '',
              clickable: true,
              field: (r) => (
                <>
                  <Button
                    variant="contained"
                    color="error"
                    startIcon={<DeleteOutlined />}
                    onClick={() => {
                      if (tryDelete === r.Name) {
                        setIsLoading(true);
                        API.docker.networkDelete(r.Name).then(() => {
                          refresh();
                          setIsLoading(false);
                        }).catch(() => {
                          setIsLoading(false);
                          refresh();
                        });
                      } else {
                        setTryDelete(r.Name);
                      }
                    }}
                  >
                    {tryDelete === r.Name ? "Really?" : "Delete"}
                  </Button>
                </>
              ),
            },
          ]}
        />
      )}
    </>
  );
}

export default NetworkManagementList;