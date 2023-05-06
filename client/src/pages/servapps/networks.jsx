// material-ui
import { CloseSquareOutlined, DeleteOutlined, PlusCircleOutlined, SyncOutlined } from '@ant-design/icons';
import { Button, Chip, CircularProgress, Stack, useTheme } from '@mui/material';
import { useEffect, useState } from 'react';

import * as API from '../../api';
import PrettyTableView from '../../components/tableView/prettyTableView';

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

      {isLoading && (<div style={{ height: '550px' }}>
        <center>
          <br />
          <CircularProgress />
        </center>
      </div>
      )}

      {!isLoading && rows && (
        <PrettyTableView
          data={rows}
          onRowClick={() => { }}
          getKey={(r) => r.Id}
          columns={[
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
                <div key={index}>
                  {config.Gateway} / {config.Subnet}
                </div>
              )),
            },
            {
              title: 'Created At',
              screenMin: 'lg',
              field: (r) => new Date(r.Created).toLocaleString(),
            },
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
                      if (tryDelete === r.Id) {
                        setIsLoading(true);
                        API.docker.networkDelete(r.Id).then(() => {
                          refresh();
                          setIsLoading(false);
                        });
                      } else {
                        setTryDelete(r.Id);
                      }
                    }}
                  >
                    {tryDelete === r.Id ? "Really?" : "Delete"}
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