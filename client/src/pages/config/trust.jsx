import React, { useState, useEffect } from 'react';
import {
  Stack,
  Typography,
  Paper,
  Button,
  Box,
  Alert,
  AlertTitle,
} from '@mui/material';
import * as API from '../../api';
import { CosmosCollapse } from './users/formShortcuts';
import { DownloadFile } from '../../api/downloadButton';

const PlatformGuide = ({ platform, children }) => (
  <Box sx={{ mb: 2 }}>
    <Typography variant="h6" sx={{ mb: 1 }}>
      {platform}
    </Typography>
    {children}
  </Box>
);

const TrustPage = () => {
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
          This is only needed if you use a self-signed certificate, or if you have .local / IP based routes. Otherwise, you can ignore this!
        </Alert>
        <Typography variant="h4" gutterBottom>
          Understanding Self-Signed Certificates
        </Typography>
        <Typography paragraph>
          HTTPS uses certificates to encrypt data between your browser and the sites you visit. 
          In your home server, you have the choices between a public certificate from a Certificate Authority (CA)
          which is trusted by all browsers, or a self-signed certificate which is not trusted by default.
          A public certificate requires a full domain name, that is why Cosmos fallsback to a self-signed certificate
          when an IP, or a .local domain is used. In order to ensure maximum security, you can setup Cosmos as a CA,
          in case you need to use a self-signed certificate. This will tell your browser to trust the certificates issued by Cosmos.
        </Typography>
        <Typography paragraph>
          While self-signed certificates aren't usually recommended for public-facing websites,
          they can be a secure choice for internal applications and development environments.
          By properly installing our Certificate Authority (CA) certificate, you can:
        </Typography>
        <ul>
          <li>
            <Typography>
              Ensure encrypted communication between your browser and our services
            </Typography>
          </li>
          <li>
            <Typography>
              Eliminate security warnings while maintaining a secure connection
            </Typography>
          </li>
          <li>
            <Typography>
              Have full control over your certificate infrastructure
            </Typography>
          </li>
          <li>
            <Typography>
              Fix issues with IOS apps such as Emby that don't support self-signed certificates
            </Typography>
          </li>
        </ul>
      </Paper>

      {/* Certificate Download Section */}
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>
          Certificate Installation
        </Typography>
        <Alert severity="info" sx={{ mb: 2 }}>
          <AlertTitle>Detected Platform: {userPlatform}</AlertTitle>
          Follow the instructions below to install the certificate on your device.
        </Alert>
        
        {/* Certificate download */}
        <DownloadFile 
          filename="ca.crt" 
          content={status.CACert} 
          label="Download CA Certificate" 
        />

        {/* Current Platform Instructions */}
        <Typography variant="h6" sx={{ mt: 3 }} gutterBottom>
          Installation Steps for {userPlatform}
        </Typography>
        <PlatformInstructions platform={userPlatform} />
      </Paper>

      {/* Other Platforms Section */}
      <Paper sx={{ p: 3 }}>
        <Typography variant="h5" gutterBottom>
          Instructions for Other Platforms
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
        Nothing to Install
      </Typography>
      <Typography>
        You are not currently using any self-signed certificate. There is no need to install any certificate.
      </Typography>
    </Stack>
  );
};

// Platform-specific instruction components
const WindowsInstructions = () => (
  <Stack spacing={1}>
    <Typography>1. Double-click the downloaded certificate file</Typography>
    <Typography>2. Click "Install Certificate"</Typography>
    <Typography>3. Select "Local Machine" and click "Next"</Typography>
    <Typography>4. Choose "Place all certificates in the following store"</Typography>
    <Typography>5. Click "Browse" and select "Trusted Root Certification Authorities"</Typography>
    <Typography>6. Click "Next" and then "Finish"</Typography>
    <Typography>7. Restart your browser</Typography>
  </Stack>
);

const MacOSInstructions = () => (
  <Stack spacing={1}>
    <Typography>1. Double-click the downloaded certificate file</Typography>
    <Typography>2. Keychain Access will open automatically</Typography>
    <Typography>3. Add the certificate to the "System" keychain</Typography>
    <Typography>4. Double-click the imported certificate</Typography>
    <Typography>5. Expand the "Trust" section</Typography>
    <Typography>6. Set "When using this certificate" to "Always Trust"</Typography>
    <Typography>7. Restart your browser</Typography>
  </Stack>
);

const LinuxInstructions = () => (
  <Stack spacing={1}>
    <Typography>1. Open terminal</Typography>
    <Typography>2. Copy the certificate to the correct directory:</Typography>
    <Typography sx={{ fontFamily: 'monospace', bgcolor: 'black', p: 1, color: 'white' }}>
      sudo cp ca-certificate.crt /usr/local/share/ca-certificates/
    </Typography>
    <Typography>3. Update the certificate store:</Typography>
    <Typography sx={{ fontFamily: 'monospace', bgcolor: 'black', p: 1, color: 'white' }}>
      sudo update-ca-certificates
    </Typography>
    <Typography>4. Restart your browser</Typography>
  </Stack>
);

const IOSInstructions = () => (
  <Stack spacing={1}>
    <Typography>1. Download the certificate</Typography>
    <Typography>2. Go to Settings</Typography>
    <Typography>3. Tap on "Profile Downloaded" near the top</Typography>
    <Typography>4. Tap "Install" in the top right</Typography>
    <Typography>5. Enter your device passcode</Typography>
    <Typography>6. Tap "Install" again to confirm</Typography>
    <Typography>7. Go to Settings - General - About - Certificate Trust Settings</Typography>
    <Typography>8. Enable trust for the installed certificate</Typography>
    <Typography>9. Restart your browser</Typography>
  </Stack>
);

const AndroidInstructions = () => (
  <Stack spacing={1}>
    <Typography>1. Download the certificate</Typography>
    <Typography>2. Go to Settings - Security</Typography>
    <Typography>3. Tap "Install from storage" (might vary by device)</Typography>
    <Typography>4. Find and select the downloaded certificate</Typography>
    <Typography>5. Enter a name for the certificate</Typography>
    <Typography>6. Select "VPN and apps" or "CA certificate"</Typography>
    <Typography>7. Tap OK to confirm installation</Typography>
    <Typography>8. Restart your browser</Typography>
  </Stack>
);

const PlatformInstructions = ({ platform }) => {
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
          Platform-specific instructions not available. Please check the instructions below for your platform.
        </Typography>
      );
  }
};

export default TrustPage;