import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Stack,
  Typography,
  Paper,
  Box,
  Alert,
  AlertTitle,
} from '@mui/material';
import * as API from '../../api';
import { CosmosCollapse } from './users/formShortcuts';
import { DownloadFile } from '../../api/downloadButton';

const PlatformGuide = ({ platform, children }) => {
  const { t } = useTranslation();
  return (
    <Box sx={{ mb: 2 }}>
      <Typography variant="h6" sx={{ mb: 1 }}>
        {platform}
      </Typography>
      {children}
    </Box>
  );
};

const TrustPage = () => {
  const { t } = useTranslation();
  const [status, setStatus] = useState(null);
  const [userPlatform, setUserPlatform] = useState('');

  useEffect(() => {
    // Detect user's platform
    const platform = navigator.platform.toLowerCase();
    if (platform.includes('win')) setUserPlatform('Windows');
    else if (platform.includes('mac')) setUserPlatform('MacOS');
    else if (platform.includes('linux')) setUserPlatform('Linux');
    else if (/iphone|ipad|ipod/.test(navigator.userAgent.toLowerCase())) setUserPlatform('iOS');
    else if (/android/.test(navigator.userAgent.toLowerCase())) setUserPlatform('Android');
    else setUserPlatform('Unknown');

    // Fetch certificate status
    API.getStatus().then((res) => {
      setStatus(res.data);
    });
  }, []);

  if (!status) return null;

  return status.CACert ? (
    <Stack spacing={4} sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
      {/* Introduction Section */}
      <Paper sx={{ p: 3 }}>
        <Alert severity="info" sx={{ mb: 2 }}>
          {t('trust.selfSignedInfo')}
        </Alert>
        <Typography variant="h4" gutterBottom>
          {t('trust.title')}
        </Typography>
        <Typography paragraph>
          {t('trust.intro.first')}
        </Typography>
        <Typography paragraph>
          {t('trust.intro.second')}
        </Typography>
        <ul>
          {t('trust.benefits', { returnObjects: true }).map((benefit, index) => (
            <li key={index}>
              <Typography>{benefit}</Typography>
            </li>
          ))}
        </ul>
      </Paper>

      {/* Certificate Download Section */}
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>
          {t('trust.installation.title')}
        </Typography>
        <Alert severity="info" sx={{ mb: 2 }}>
          <AlertTitle>{t('trust.installation.detectedPlatform', { platform: userPlatform })}</AlertTitle>
          {t('trust.installation.followInstructions')}
        </Alert>
        
        <DownloadFile 
          filename="ca.crt" 
          content={status.CACert} 
          label={t('trust.installation.downloadButton')} 
        />

        <Typography variant="h6" sx={{ mt: 3 }} gutterBottom>
          {t('trust.installation.stepsFor', { platform: userPlatform })}
        </Typography>
        <PlatformInstructions platform={userPlatform} />
      </Paper>

      {/* Other Platforms Section */}
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>
          {t('trust.otherPlatforms.title')}
        </Typography>
        
        {userPlatform !== 'Windows' && (
          <CosmosCollapse title="Windows">
            <WindowsInstructions />
          </CosmosCollapse>
        )}
        
        {userPlatform !== 'MacOS' && (
          <CosmosCollapse title="MacOS">
            <MacOSInstructions />
          </CosmosCollapse>
        )}
        
        {userPlatform !== 'Linux' && (
          <CosmosCollapse title="Linux">
            <LinuxInstructions />
          </CosmosCollapse>
        )}
        
        {userPlatform !== 'iOS' && (
          <CosmosCollapse title="iOS">
            <IOSInstructions />
          </CosmosCollapse>
        )}
        
        {userPlatform !== 'Android' && (
          <CosmosCollapse title="Android">
            <AndroidInstructions />
          </CosmosCollapse>
        )}
      </Paper>
    </Stack>
  ) : (
    <Stack spacing={4} sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
      <Typography variant="h4" gutterBottom>
        {t('trust.noInstall.title')}
      </Typography>
      <Typography>
        {t('trust.noInstall.message')}
      </Typography>
    </Stack>
  );
};

// Platform-specific instruction components
const WindowsInstructions = () => {
  const { t } = useTranslation();
  return (
    <Stack spacing={1}>
      {t('trust.platforms.windows', { returnObjects: true }).map((step, index) => (
        <Typography key={index}>{`${index + 1}. ${step}`}</Typography>
      ))}
    </Stack>
  );
};

const MacOSInstructions = () => {
  const { t } = useTranslation();
  return (
    <Stack spacing={1}>
      {t('trust.platforms.macos', { returnObjects: true }).map((step, index) => (
        <Typography key={index}>{`${index + 1}. ${step}`}</Typography>
      ))}
    </Stack>
  );
};

const LinuxInstructions = () => {
  const { t } = useTranslation();
  return (
    <Stack spacing={1}>
      {t('trust.platforms.linux', { returnObjects: true }).map((step, index) => (
        <Typography key={index}>{`${index + 1}. ${step}`}</Typography>
      ))}
    </Stack>
  );
};

const IOSInstructions = () => {
  const { t } = useTranslation();
  return (
    <Stack spacing={1}>
      {t('trust.platforms.ios', { returnObjects: true }).map((step, index) => (
        <Typography key={index}>{`${index + 1}. ${step}`}</Typography>
      ))}
    </Stack>
  );
};

const AndroidInstructions = () => {
  const { t } = useTranslation();
  return (
    <Stack spacing={1}>
      {t('trust.platforms.android', { returnObjects: true }).map((step, index) => (
        <Typography key={index}>{`${index + 1}. ${step}`}</Typography>
      ))}
    </Stack>
  );
};

const PlatformInstructions = ({ platform }) => {
  const { t } = useTranslation();
  switch (platform) {
    case 'Windows':
      return <WindowsInstructions />;
    case 'MacOS':
      return <MacOSInstructions />;
    case 'Linux':
      return <LinuxInstructions />;
    case 'iOS':
      return <IOSInstructions />;
    case 'Android':
      return <AndroidInstructions />;
    default:
      return (
        <Typography color="error">
          {t('trust.platforms.notAvailable')}
        </Typography>
      );
  }
};

export default TrustPage;