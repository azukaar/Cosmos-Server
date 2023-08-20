import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import AddDeviceModal from "./addDevice";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, DesktopOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import IsLoggedIn from "../../isLoggedIn";

export const ConstellationIndex = () => {
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [devices, setDevices] = useState(null);

  const refreshConfig = async () => {
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setDevices((await API.constellation.list()).data || []);
  };

  useEffect(() => {
    refreshConfig();
  }, []);

  const getIcon = (r) => {
    if (r.deviceName.toLowerCase().includes("mobile") || r.deviceName.toLowerCase().includes("phone")) {
      return <MobileOutlined />
    }
    else if (r.deviceName.toLowerCase().includes("laptop") || r.deviceName.toLowerCase().includes("computer")) {
      return <LaptopOutlined />
    } else if (r.deviceName.toLowerCase().includes("desktop")) {
      return <DesktopOutlined />
    } else if (r.deviceName.toLowerCase().includes("tablet")) {
      return <TabletOutlined />
    } else {
      return <CloudOutlined />
    }
  }

  return <>
    <IsLoggedIn />
    {devices && config && <>
      <PrettyTableView 
            data={devices}
            getKey={(r) => r.deviceName}
            buttons={[
              <AddDeviceModal isAdmin={isAdmin} config={config} refreshConfig={refreshConfig} devices={devices} />
            ]}
            columns={[
                {
                    title: '',
                    field: getIcon,
                },
                {
                    title: 'Device Name',
                    field: (r) => <strong>{r.deviceName}</strong>,
                },
                {
                    title: 'Owner',
                    field: (r) => <strong>{r.nickname}</strong>,
                },
                {
                    title: 'Constellation IP',
                    screenMin: 'md', 
                    field: (r) => r.ip,
                },
                {
                  title: '',
                  clickable: true,
                  field: (r) => {
                    return <DeleteButton onDelete={async () => {
                      alert("caca")
                    }}></DeleteButton>
                  }
                }
            ]}
        />
    </>}
  </>
};