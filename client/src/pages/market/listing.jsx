import { Box, CircularProgress, Input, InputAdornment, Stack, Tooltip, Alert } from "@mui/material";
import { HomeBackground, TransparentHeader } from "../home";
import { useEffect, useState } from "react";
import * as API from "../../api";
import { useTheme } from "@emotion/react";
import Grid2 from "@mui/material/Unstable_Grid2/Grid2";
import { useParams } from "react-router";
import Carousel from 'react-material-ui-carousel'
import { Paper, Button, Chip } from '@mui/material'
import { Link } from "react-router-dom";
import { Link as LinkMUI } from '@mui/material'
import DockerComposeImport from '../servapps/containers/docker-compose';
import { AppstoreAddOutlined, SearchOutlined, WarningOutlined } from "@ant-design/icons";
import ResponsiveButton from "../../components/responseiveButton";
import { useClientInfos } from "../../utils/hooks";
import EditSourcesModal from "./sources";
import { PersistentCheckbox } from "../../components/persistentInput";
import { useTranslation } from 'react-i18next'; 

function Screenshots({ screenshots }) {
  const aspectRatioContainerStyle = {
    position: 'relative',
    overflow: 'hidden',
    paddingTop: '56.25%', // 9 / 16 = 0.5625 or 56.25%
    height: 0,
  };

  // This will position the image correctly within the aspect ratio container
  const imageStyle = {
    position: 'absolute',
    top: 0,
    left: 0,
    bottom: 0,
    right: 0,
    width: '100%',
    height: '100%',
    objectFit: 'cover', // This will cover the area without losing the aspect ratio
  };

  return screenshots.length > 1 ? (
    <Carousel animation="slide" navButtonsAlwaysVisible={false} fullHeightHover="true" swipe={false}>
      {
        screenshots.map((item, i) => <div style={{height: "400px"}}>
          <img style={{ maxHeight: '100%', width: '100%' }} key={i} src={item} />
        </div>)
      }
    </Carousel>)
    : <div style={{height: "400px"}}>
        <img src={screenshots[0]} style={{ maxHeight: '100%', width: '100%' }} />
      </div>
}

function Showcases({ showcase, isDark, isAdmin }) {
  return (
    <Carousel animation="slide" navButtonsAlwaysVisible={false} fullHeightHover="true" swipe={false}>
      {
        showcase.map((item, i) => <ShowcasesItem isDark={isDark} key={i} item={item} isAdmin={isAdmin} />)
      }
    </Carousel>
  )
}

function ShowcasesItem({ isDark, item, isAdmin }) {
  const { t, i18n } = useTranslation();
  return (
    <Paper style={{
      position: 'relative',
      background: 'url(' + item.screenshots[0] + ')',
      height: '31vh',
      backgroundSize: 'auto 100%',
      maxWidth: '120vh',
      margin: 'auto',
    }}>
      <Stack direction="row" spacing={2} style={{ height: '100%', overflow: 'hidden' }} justifyContent="flex-end">
        <Stack direction="column" spacing={2} style={{ height: '100%' }} sx={{
          backgroundColor: isDark ? '#1A2027' : '#fff',
          padding: '20px 100px',
          width: '50%',
          filter: 'drop-shadow(-20px 0px 20px rgba(0, 0, 0, 1))',

          '@media (max-width: 1100px)': {
            width: '70%',
            padding: '20px 40px',
          },

          '@media (max-width: 600px)': {
            width: '80%',
            padding: '20px 20px',
          }
        }}>
          <Stack direction="row" spacing={2}>
            <img src={item.icon} style={{ width: '36px', height: '36px' }} />
            <h2>{item.name}</h2>
          </Stack>
          <p dangerouslySetInnerHTML={{ __html: item?.translation?.[i18n?.resolvedLanguage]?.longDescription || item?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.longDescription || item.longDescription }} style={{
            overflow: 'hidden',
          }}></p>
          <Stack direction="row" spacing={2} justifyContent="flex-start">
            {isAdmin && <div>
              <DockerComposeImport installerInit defaultName={item.name} dockerComposeInit={item.compose} />
            </div>}
            <Link to={"/cosmos-ui/market-listing/cosmos-cloud/" + item.name} style={{
              textDecoration: 'none',
            }}>
              <Button className="CheckButton" color="primary" variant="outlined">
                {t('navigation.market.viewButton')}
              </Button> 
            </Link>
          </Stack>
        </Stack>
      </Stack>
    </Paper>
  )
}

const appCardStyle = (theme) => ({
  width: '100%',
  backgroundColor: theme.palette.mode === 'dark' ? '#1A2027' : '#fff',
  ...theme.typography.body2,
  padding: theme.spacing(1),
  color: theme.palette.text.secondary,
})

const gridAnim = {
  transition: 'all 0.2s ease',
  opacity: 1,
  transform: 'translateY(0px)',
  '&.MuiGrid2-item--hidden': {
    opacity: 0,
    transform: 'translateY(-20px)',
  },
};

const MarketPage = () => {
  const { t, i18n } = useTranslation();
  const [apps, setApps] = useState([]);
  const [showcase, setShowcase] = useState([]);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { appName, appStore } = useParams();
  const [search, setSearch] = useState("");
  const [filterDups, setFilterDups] = useState(false);
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  const backgroundStyle = isDark ? {
    backgroundColor: 'rgb(0,0,0)',
    // borderTop: '1px solid #595959'
  } : {
    backgroundColor: 'rgb(255,255,255)',
    // borderTop: '1px solid rgb(220,220,220)'
  };

  const refresh = () => {
    API.market.list().then((res) => {
      setApps(res.data.all);
      setShowcase(res.data.showcase);
    });
  };

  useEffect(() => {
    refresh();
  }, []);

  let openedApp = null;
  if (appName && Object.keys(apps).length > 0) {
    openedApp = apps[appStore].find((app) => app.name === appName);
    openedApp.appstore = appStore;
  }

  let appList = apps && Object.keys(apps).reduce((acc, appstore) => {
    const a = apps[appstore].map((app) => {
      app.appstore = appstore;
      return app;
    });

    return acc.concat(a);
  }, []);

  appList.sort((a, b) => {
    if (a.name > b.name) {
      return 1;
    }

    if (a.name < b.name) {
      return -1;
    }

    if (a.appstore > b.appstore || b.appstore === 'cosmos-cloud') {
      return 1;
    }

    if (a.appstore < b.appstore || a.appstore === 'cosmos-cloud') {
      return -1;
    }

    return 0;
  });

  let filteredAppList = appList.filter((app) => {
    if (!search || search.length <= 2) {
      return true;
    }
    return app.name.toLowerCase().includes(search.toLowerCase()) ||
      app.tags.join(' ').toLowerCase().includes(search.toLowerCase());
  })
  .filter((app) => {
    if (!filterDups) {
      return true;
    } else if(app.appstore === 'cosmos-cloud') {
      return true;
    } else if(appList.filter((a) => a.name === app.name && a.appstore === 'cosmos-cloud').length > 0) {
      return false;
    }

    return true;
  });

  return <>
    <HomeBackground />
    <TransparentHeader />
    {openedApp && <Box style={{
      position: 'fixed',
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      zIndex: 1300,
      backgroundColor: 'rgba(0,0,0,0.5)',
    }}>
      <Link to="/cosmos-ui/market-listing" as={Box}
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
        }}></Link>

      <Stack direction="row" spacing={2} style={{ height: '100%' }} justifyContent="flex-end">
        <Stack direction="column" spacing={3} style={{ height: '100%', overflow: "auto" }} sx={{
          backgroundColor: isDark ? '#1A2027' : '#fff',
          padding: '80px 80px',
          width: '100%',
          maxWidth: '800px',
          filter: 'drop-shadow(-20px 0px 20px rgba(0, 0, 0, 1))',

          '@media (max-width: 700px)': {
            padding: '60px 40px',
          },

          '@media (max-width: 500px)': {
            padding: '40px 20px',
          }
        }}>

          <Link to="/cosmos-ui/market-listing" style={{
            textDecoration: 'none',
          }}>
            <Button className="CheckButton" color="primary" variant="outlined">
              {t('global.close')}
            </Button>
          </Link>

          <div style={{ textAlign: 'center' }}>
            <Screenshots screenshots={openedApp.screenshots} isAdmin={isAdmin}/>
          </div>

          <Stack direction="row" spacing={2}>
            <img src={openedApp.icon} style={{ width: '36px', height: '36px' }} />
            <h2>{openedApp.name} <span style={{color:'grey'}}>{openedApp.appstore != 'cosmos-cloud' ? (' @ '+openedApp.appstore) : ''}</span></h2>
          </Stack>

          <div>
            {openedApp.tags && openedApp.tags.slice(0, 8).map((tag) => <Chip label={tag} />)}
          </div>

          <div>
            {openedApp.supported_architectures && openedApp.supported_architectures.slice(0, 8).map((tag) => <Chip label={tag} />)}
          </div>

          {openedApp.appstore != 'cosmos-cloud' && <div>
            <div>
            <Tooltip title={t('navigation.market.unofficialMarketTooltip')}>
                <WarningOutlined />
              </Tooltip> <strong>{t('global.source')}:</strong> {openedApp.appstore} 
            </div>
          </div>}
          
          <div>
            <div><strong>{t('navigation.market.repository')}:</strong> <LinkMUI href={openedApp.repository}>{openedApp.repository}</LinkMUI></div>
            <div><strong>{t('navigation.market.image')}:</strong> <LinkMUI href={openedApp.image}>{openedApp.image}</LinkMUI></div>
            <div><strong>{t('navigation.market.compose')}:</strong> <LinkMUI href={openedApp.compose}>{openedApp.compose}</LinkMUI></div>
          </div>

          <div dangerouslySetInnerHTML={{ __html: openedApp?.translation?.[i18n?.resolvedLanguage]?.longDescription || openedApp?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.longDescription || openedApp.longDescription }}></div>

          {isAdmin ? <div>
            <DockerComposeImport installerInit defaultName={openedApp.name} dockerComposeInit={openedApp.compose} />
          </div> : <div style={{
            color: 'orange',
            fontStyle: 'italic',
          }}>{t('navigation.market.mustBeAdmin')}</div>}
        </Stack>
      </Stack>
    </Box>}

    <Stack style={{ position: 'relative' }} spacing={1}>
      <Stack style={{ height: '35vh' }} spacing={1}>
        {(!showcase || !Object.keys(showcase).length) && <Box style={{
          width: '100%',
          height: '100%',
          zIndex: 1000,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}>
          <CircularProgress
            size={100}
          />
        </Box>}
        {showcase && showcase.length > 0 && <Showcases showcase={showcase} isDark={isDark} isAdmin={isAdmin} />}
      </Stack>

      <Stack spacing={1} style={{
        ...backgroundStyle,
        marginLeft: "-24px",
        marginRight: "-24px",
        marginBottom: "-24px",
        minHeight: 'calc(65vh - 80px)',
        padding: '24px',
      }}>
        <h2>{t('navigation.market.applicationsTitle')}</h2>
        <Stack direction="row" spacing={2}>
          <Input placeholder={t('navigation.market.search',  {count: filteredAppList.length})}
            value={search}
            style={{ maxWidth: '400px' }}
            startAdornment={
              <InputAdornment position="start">
                <SearchOutlined />
              </InputAdornment>
            }
            onChange={(e) => {
              setSearch(e.target.value);
            }}
          />

          <Link to="/cosmos-ui/servapps/new-service">
            <ResponsiveButton
              variant="contained"
              startIcon={<AppstoreAddOutlined />}
            >{t('navigation.market.startServAppButton')}</ResponsiveButton>
          </Link>
          <DockerComposeImport refresh={() => { }} />
          <EditSourcesModal onSave={refresh} />
          <PersistentCheckbox name="filterDups" label={t('navigation.market.filterDuplicateCheckbox')} value={filterDups} onChange={setFilterDups} />
        </Stack>
        {(!apps || !Object.keys(apps).length) && <Box style={{
          width: '100%',
          height: '100%',
          zIndex: 1000,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginTop: '150px',
        }}>
          <CircularProgress
            size={100}
          />
        </Box>}

        {apps && Object.keys(apps).length > 0 && <Grid2 container spacing={{ xs: 1, sm: 1, md: 2 }}>
          {filteredAppList.map((app) => {
              return <Grid2
              style={{
                ...gridAnim,
                cursor: 'pointer',
              }} xs={12} sm={12} md={6} lg={4} xl={3} key={app.name + app.appstore} item><Link to={"/cosmos-ui/market-listing/" + app.appstore + "/" + app.name} style={{
                textDecoration: 'none',
              }}>
                  <div key={app.name} style={appCardStyle(theme)}>
                    <Stack spacing={3} direction={'row'} alignItems={'center'} style={{ padding: '0px 15px' }}>
                      <img src={app.icon} style={{ width: 64, height: 64 }} />
                      <Stack spacing={1}>
                        <div style={{ fontWeight: "bold" }}>{app.name}<span style={{color:'grey'}}>{app.appstore != 'cosmos-cloud' ? (' @ '+app.appstore) : ''}</span></div>
                        <div style={{
                          height: '40px',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'pre-wrap',
                        }}
                        >{ app?.translation?.[i18n?.resolvedLanguage]?.description || app?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.description || app.description }</div>
                        <Stack direction={'row'} spacing={1}>
                          <div style={{
                            fontStyle: "italic", opacity: 0.7,
                            overflow: 'hidden',
                            height: '21px',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'pre-wrap',
                          }}>{app.tags.slice(0, 3).join(", ")}</div>
                        </Stack>
                      </Stack>
                    </Stack>
                  </div>

                </Link>
              </Grid2>
            })}
        </Grid2>}
      </Stack>
    </Stack>
  </>
};

export default MarketPage;