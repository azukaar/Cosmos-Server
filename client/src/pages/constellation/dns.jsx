import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import AddDeviceModal from "./addDevice";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import IsLoggedIn from "../../isLoggedIn";
import { Alert, Button, CircularProgress, InputLabel, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import { isDomain } from "../../utils/indexs";
import ConfirmModal from "../../components/confirmModal";
import UploadButtons from "../../components/fileUpload";

export const ConstellationDNS = () => {
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);

  const refreshConfig = async () => {
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
  };

  useEffect(() => {
    refreshConfig();
  }, []);

  return <>
    {(config) ? <>
      <Stack spacing={2} style={{maxWidth: "1000px"}}>
      <div>
        <MainCard title={"Constellation Internal DNS"} content={config.constellationIP}>
          <Stack spacing={2}>

          <Formik
            initialValues={{
              Fallback: config.ConstellationConfig.DNSFallback,
              DNSBlockBlacklist: config.ConstellationConfig.DNSBlockBlacklist,
              DNSAdditionalBlocklists: config.ConstellationConfig.DNSAdditionalBlocklists || [],
              CustomDNSEntries: config.ConstellationConfig.CustomDNSEntries || []
            }}
            onSubmit={(values) => {
              let newConfig = { ...config };
              newConfig.ConstellationConfig.DNSFallback = values.Fallback;
              newConfig.ConstellationConfig.DNSBlockBlacklist = values.DNSBlockBlacklist;
              newConfig.ConstellationConfig.DNSAdditionalBlocklists = values.DNSAdditionalBlocklists;
              newConfig.ConstellationConfig.CustomDNSEntries = values.CustomDNSEntries;
              
              return API.config.set(newConfig);
            }}
          >
            {(formik) => (
              <form onSubmit={formik.handleSubmit}>
                <Stack spacing={2}>        
                  <Alert severity="info">This is a DNS that runs inside your Constellation network. It automatically
                  rewrites your domains DNS entries to be local to your network, and also allows you to do things like block ads
                  and trackers on all devices connected to your network. You can also add custom DNS entries to resolve to specific
                  IP addresses. This DNS server is only accessible from inside your network.</Alert>

                  <CosmosInputText formik={formik} name="Fallback" label="DNS Fallback" placeholder={'8.8.8.8:53'} />
                  
                  <CosmosFormDivider title={"DNS Blocklists"} />

                  <CosmosCheckbox formik={formik} name="DNSBlockBlacklist" label="Use Blacklists to block domains" />

                  <Alert severity="warning">When changing your DNS records, always use private mode on your browser and allow some times for various caches to expire.</Alert>

                  <InputLabel>DNS Blocklist URLs</InputLabel>
                  {formik.values.DNSAdditionalBlocklists && formik.values.DNSAdditionalBlocklists.map((item, index) => (
                    <Stack direction={"row"} spacing={2} key={`DNSAdditionalBlocklists${item}`} width={"100%"}>
                      <DeleteButton onDelete={() => {
                        formik.setFieldValue("DNSAdditionalBlocklists", [...formik.values.DNSAdditionalBlocklists.slice(0, index), ...formik.values.DNSAdditionalBlocklists.slice(index + 1)]);
                      }} />
                      <div style={{flexGrow: 1}}>
                        <CosmosInputText
                          value={item} 
                          name={`DNSAdditionalBlocklists${index}`}
                          placeholder={'https://example.com/blocklist.txt'} 
                          onChange={(e) => {
                            formik.setFieldValue("DNSAdditionalBlocklists", [...formik.values.DNSAdditionalBlocklists.slice(0, index), e.target.value, ...formik.values.DNSAdditionalBlocklists.slice(index + 1)]);
                          }}
                        />
                      </div>
                    </Stack>
                  ))}
                  <Stack direction="row" spacing={2}>
                    <Button variant="outlined" onClick={() => {
                      formik.setFieldValue("DNSAdditionalBlocklists", [...formik.values.DNSAdditionalBlocklists, ""]);
                    }}>Add</Button>
                    <Button variant="outlined" onClick={() => {
                      formik.setFieldValue("DNSAdditionalBlocklists", [
                        "https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
                        "https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
                        "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
                        "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-only/hosts"
                      ]);
                    }}>Reset Default</Button>
                  </Stack>

                  <CosmosFormDivider title={"DNS Custom Entries"} />

                  <InputLabel>DNS Custom Entries</InputLabel>
                  {formik.values.CustomDNSEntries && formik.values.CustomDNSEntries.map((item, index) => (
                    <Stack direction={"row"} spacing={2} key={`CustomDNSEntries${item}`} width={"100%"}>
                      <DeleteButton onDelete={() => {
                        formik.setFieldValue("CustomDNSEntries", [...formik.values.CustomDNSEntries.slice(0, index), ...formik.values.CustomDNSEntries.slice(index + 1)]);
                      }} />
                      <div style={{flexGrow: 1}}>
                        <CosmosInputText
                          value={item.Key} 
                          name={`CustomDNSEntries${index}-key`}
                          placeholder={'domain.com'} 
                          onChange={(e) => {
                            const updatedCustomDNSEntries = [...formik.values.CustomDNSEntries];
                            updatedCustomDNSEntries[index].Key = e.target.value;
                            formik.setFieldValue("CustomDNSEntries", updatedCustomDNSEntries);
                          }}
                        />
                      </div>
                      <div style={{flexGrow: 1}}>
                        <CosmosInputText
                          value={item.Value} 
                          name={`CustomDNSEntries${index}-value`}
                          placeholder={'1213.123.123.123'} 
                          onChange={(e) => {
                            const updatedCustomDNSEntries = [...formik.values.CustomDNSEntries];
                            updatedCustomDNSEntries[index].Value = e.target.value;
                            formik.setFieldValue("CustomDNSEntries", updatedCustomDNSEntries);
                          }}
                        />
                      </div>
                    </Stack>
                  ))}
                  <Stack direction="row" spacing={2}>
                    <Button variant="outlined" onClick={() => {
                      formik.setFieldValue("CustomDNSEntries", [...formik.values.CustomDNSEntries, {
                        Key: "",
                        Value: "",
                        Type: "A"
                      }]);
                    }}>Add</Button>
                    <Button variant="outlined" onClick={() => {
                      formik.setFieldValue("CustomDNSEntries", [
                      ]);
                    }}>Reset</Button>
                  </Stack>

                  <LoadingButton
                      disableElevation
                      loading={formik.isSubmitting}
                      type="submit"
                      variant="contained"
                      color="primary"
                    >
                      Save
                  </LoadingButton>
                </Stack>
              </form>
            )}
          </Formik>
          </Stack>
        </MainCard>
      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};