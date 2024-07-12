// material-ui
import { CloseSquareOutlined, DeleteOutlined, PlusCircleOutlined, SyncOutlined } from '@ant-design/icons';
import { Button, Chip, CircularProgress, Stack, useTheme } from '@mui/material';
import { useEffect, useState } from 'react';

import * as API from '../../api';
import PrettyTableView from '../../components/tableView/prettyTableView';
import NewNetworkButton from './createNetwork';
import { useTranslation } from 'react-i18next';

export const NetworksColumns = (theme, isDark, t) => [
  {
    title: t('mgmt.servapps.networks.list.networkName'),
    field: (r) => <Stack direction='column'>
      <div style={{display:'inline-block', textDecoration: 'inherit', fontSize:'125%', color: isDark ? theme.palette.primary.light : theme.palette.primary.dark}}>{r.Name}</div><br/>
      <div style={{display:'inline-block', textDecoration: 'inherit', fontSize: '90%', opacity: '90%'}}>{r.Driver} driver</div>
    </Stack>,
    search: (r) => r.Name,
  },
  {
    title: t('mgmt.servapps.networks.list.networkproperties'),
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
    title: t('mgmt.servapps.networks.list.networkIpam'),
    screenMin: 'lg',
    field: (r) => r.IPAM.Config ? r.IPAM.Config.map((config, index) => (
      <Stack key={index}>
        <div>{config.Gateway}</div>
        <div>{config.Subnet}</div>
      </Stack>
    )) : t('mgmt.servapps.networks.list.networkNoIp'),
  },
  {
    title: t('global.createdAt'),
    screenMin: 'lg',
    field: (r) => r.Created ? new Date(r.Created).toLocaleString() : '-',
  },
];

const NetworkManagementList = () => {
  const { t } = useTranslation();
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
          {t('global.refresh')}
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
            ...NetworksColumns(theme, isDark, t),
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
                        }).catch(() => {
                          setIsLoading(false);
                          refresh();
                        });
                      } else {
                        setTryDelete(r.Id);
                      }
                    }}
                  >
                    {tryDelete === r.Id ? t('global.confirmDeletion') : t('global.delete')}
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