import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CheckCircleOutlined, CloudOutlined, CompassOutlined, DesktopOutlined, ExpandOutlined, InfoCircleOutlined, LaptopOutlined, MenuFoldOutlined, MenuOutlined, MinusCircleFilled, MobileOutlined, NodeCollapseOutlined, PlusCircleFilled, PlusCircleOutlined, ReloadOutlined, SettingFilled, TabletOutlined, WarningFilled, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, InputLabel, LinearProgress, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton, TreeItem, TreeView } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { PascalToSnake, isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import { useTheme } from '@mui/material/styles';
import MiniPlotComponent from '../dashboard/components/mini-plot';
import { useTranslation } from 'react-i18next';

const temperatureChip = (temperature) => {
  if (temperature < 45) {
    return "🟢" 
  } else if (temperature < 55) {
    return "🟡"
  } else if (temperature < 65) {
    return "🟠"
  } else {
    return "🔴"
  }
}

const diskColor = (disk) => {
  if (!disk || !disk.smart || !disk.smart.AdditionalData ) {
    return "gray"
  }
  
  let temperature = disk.smart && disk.smart.Temperature;
  
  if(!disk.rota) {
    if (!temperature) {
      return "gray"
    }

    if (temperature < 45) {
      return "green"
    } else if (temperature < 55) {
      return "yellow"
    } else if (temperature < 65) {
      return "orange"
    } else {
      return "red"
    }
  }

  let health = disk.smart && healthStatus(disk, Object.values(disk.smart.AdditionalData.Attrs));
  if (!temperature || !health) {
    return "gray"
  }

  if (temperature < 45 && health > 90) {
    return "green"
  } else if (temperature < 55 && health > 80) {
    return "yellow"
  } else if (temperature < 65 && health > 50) {
    return "orange"
  } else {
    return "red"
  }
}
const diskChip = (disk) => {
  if (!disk) {
    return "⚪"
  } else if (!disk.rota) {
    return  temperatureChip(disk.smart.Temperature);
  } else if (diskColor(disk) == "gray") {
    return "⚪"
  } else if (diskColor(disk) == "green") {
    return "🟢" 
  } else if (diskColor(disk) == "yellow") {
    return "🟡"
  } else if (diskColor(disk) == "orange") {
    return "🟠"
  } else {
    return "🔴"
  }
}

const healthStatus = (disk, fullData) => {
  let health = 100;

  if (!disk.rota) return '--';

  fullData.forEach((data) => {
    let diff = data.Current - data.threshold;
    if (diff < 0 && data.Id != 194) {
      if (data.def.critical) {
        diff = diff * 2;
      }
      health = Math.max(health + diff, 0);
    }
  });

  return health;
}

const healthChip = (disk) => {
  if (!disk || !disk.rota) {
    return "⚪"
  }

  let health = disk.smart && healthStatus(disk, Object.values(disk.smart.AdditionalData.Attrs));

  if (health < 50) {
    return "🔴"
  } else if (health < 80) {
    return "🟠"
  } else if (health < 90) {
    return "🟡"
  } else {
    return "🟢"
  }
}

let smartDef = null;

const CompleteDataSMARTDisk = (disk) => {
  if(disk.rota) {
    if(!disk || !disk.smart || !disk.smart.AdditionalData || !disk.smart.AdditionalData.Attrs || !smartDef) {
      return disk;
    }

    let newAttrs = [];
    Object.values(disk.smart.AdditionalData.Attrs).forEach((data, i) => {
      let def = smartDef.ATA[data.Id];
      let threshold = disk.smart.Thresholds.Thresholds[data.Id] || 50;

      
      if(def && data.Id != 194) {
        newAttrs.push({
          ...data,
          def,
          threshold,
        });
      }
    });

    disk.smart.AdditionalData.Attrs = newAttrs;

    return disk;
      
  } else {
    if(!disk || !disk.smart || !disk.smart.AdditionalData || !smartDef) {
      return disk;
    }

    let newAttrs = [];
    Object.keys(disk.smart.AdditionalData).forEach((dataKey, i) => {
      let defKey = PascalToSnake(dataKey);
      if(defKey == "avail_spare") defKey = "available_spare";
      if(defKey == "crit_warning") defKey = "critical_warning";
      if(dataKey == "CtrlBusyTime") defKey = "controller_busy_time";

      let def = smartDef.NVME[defKey];

      if(def) {
        let val = disk.smart.AdditionalData[dataKey];

        console.log(dataKey, val, typeof val, val[0])

        if(typeof val == 'object') {
          val = val.Val[0];
        }

        newAttrs.push({
          Id: dataKey,
          Current: val,
          def,
          threshold: dataKey == "AvailSpare" ? disk.smart.AdditionalData.SpareThresh : -1,
        });
      }
    });

    disk.smart.AdditionalData = newAttrs;
    console.log(disk)
    return disk;
  }
}

const getSMARTDef = async() => {
  smartDef = (await API.storage.disks.smartDef()).data;
}

const SMARTDialog = ({disk, OnClose}) => {
  const { t } = useTranslation();
  const fullData = disk.smart && disk.smart.AdditionalData && (disk.rota ? Object.values(disk.smart.AdditionalData.Attrs) :
    disk.smart.AdditionalData);

  if (!fullData) {
    return <Dialog open={disk} onClose={() => {
      OnClose && OnClose();
    }}>
      <DialogTitle>{t('mgmt.storage.smart.for')} {disk.name}</DialogTitle>
      <DialogContent>
          <DialogContentText>
            <Alert severity="error">{t('mgmt.storage.smart.noSmartError')}</Alert>
          </DialogContentText>
      </DialogContent>
      <DialogActions>
          <Button onClick={() => {
              OnClose && OnClose();
          }}>{t('global.close')}</Button>
      </DialogActions>
  </Dialog>;
  }

  return <Dialog open={disk} onClose={() => {
      OnClose && OnClose();
    }}>
      <DialogTitle>S.M.A.R.T. {t('mgmt.storage.smart.for')} {disk.name}</DialogTitle>
      <DialogContent>
          <DialogContentText>
            <Stack direction="row" spacing={8}>
              <Stack spacing={2}>
                <div>
                  <InputLabel>{t('mgmt.storage.smart.health')}</InputLabel>
                  {healthChip(disk) + ' ' + healthStatus(disk, disk.rota ? fullData : []) + '%'}
                </div>
                <div>
                  <InputLabel>{t('global.temperature')}</InputLabel>
                  {(disk.smart && disk.smart.Temperature) ? `${temperatureChip(disk.smart.Temperature)} ${disk.smart.Temperature}°C` : '⚪ ?'}
                </div>
                <div style={{maxWidth: '200px'}}>  
                <MiniPlotComponent noLabels agglo metrics={[
                  "cosmos.system.disk-health.temperature." + disk.name,
                ]} labels={{
                  ["cosmos.system.disk-health.temperature." + disk.name]: "Temp", 
                }}/>
                </div>
                {disk && [
                  'name',
                  'model',
                  'serial',
                  'size',
                  'rota',
                  'wwn',
                  'rev',
                ].map((key) => {
                  return <Stack key={key} direction="row" spacing={1}>
                    <div>
                      <InputLabel>{key}</InputLabel>
                      <span>{key == "rota" ? (disk[key] ? 'ATA' : "NVME") : disk[key]}</span>
                    </div>
                  </Stack>
                })}
              </Stack>
              {fullData && <PrettyTableView 
                data={fullData}
                getKey={(r) => `${r.Id}`}
                columns={[
                  {
                    title: t('global.nameTitle'),
                    field: (r) => <Tooltip title={r.def.description}><span><InfoCircleOutlined /> {r.def.display_name}</span></Tooltip>, 
                  },
                  {
                    title: <Tooltip title={t('mgmt.storage.smart.threshholdTooltip')}><span>{t('mgmt.servapps.newContainer.env.envValueInput.envValueLabel')} <InfoCircleOutlined /> </span></Tooltip>,
                    style: {minWidth: '110px'},
                    field: (r) => {
                      let StatusIcon = '';
                      if (r.threshold) {
                        StatusIcon = (r.Current >= r.threshold) ? <CheckCircleOutlined style={{color: 'green'}} /> : (r.def.critical ? <WarningFilled style={{color: 'red'}} /> : <WarningOutlined style={{color: 'orange'}} />)
                      }
                      return r.threshold >= 0 ? <>{StatusIcon} {r.Current + (r.threshold ? ' / ' + r.threshold : '')}</> : r.Current;
                    }
                  },
                ]}
              />}
            </Stack>
          </DialogContentText>
      </DialogContent>
      <DialogActions>
          <Button onClick={() => {
              OnClose && OnClose();
          }}>{t('global.close')}</Button>
      </DialogActions>
  </Dialog>
};

export default SMARTDialog;
export {
  CompleteDataSMARTDisk,
  getSMARTDef,
  diskColor,
  temperatureChip,
  healthStatus,
  healthChip,
  diskChip,
};