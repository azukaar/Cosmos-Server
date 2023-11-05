import React, { useState } from 'react';
import { Box, Tab, Tabs, Typography, MenuItem, Select, useMediaQuery, CircularProgress } from '@mui/material';
import { styled } from '@mui/system';

const StyledTabs = styled(Tabs)`
  border-right: 1px solid ${({ theme }) => theme.palette.divider};
`;

const TabPanel = (props) => {
  const { children, value, index, ...other } = props;
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('md'));

  return (
    <div
      role="tabpanel"
      style={{
        width: '100%',
      }}
      hidden={value !== index}
      id={`vertical-tabpanel-${index}`}
      aria-labelledby={`vertical-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box p={isMobile ? 1 : 3}>
          <Typography>{children}</Typography>
        </Box>
      )}
    </div>
  );
};

const a11yProps = (index) => {
  return {
    id: `vertical-tab-${index}`,
    'aria-controls': `vertical-tabpanel-${index}`,
  };
};

const PrettyTabbedView = ({ tabs, isLoading, currentTab, setCurrentTab }) => {
  const [value, setValue] = useState(0);
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('md'));
  
  if((currentTab != null && typeof currentTab === 'number') && value !== currentTab)
    setValue(currentTab);

  const handleChange = (event, newValue) => {
    setValue(newValue);
    setCurrentTab && setCurrentTab(newValue);
  };

  const handleSelectChange = (event) => {
    setValue(event.target.value);
    setCurrentTab && setCurrentTab(event.target.value);
  };

  return (
    <Box display="flex" height="100%" flexDirection={isMobile ? 'column' : 'row'}>
      {(isMobile) ? (
        <Select value={value} onChange={handleSelectChange} sx={{ minWidth: 120, marginBottom: '15px' }}>
          {tabs.map((tab, index) => (
            <MenuItem key={index} value={index}>
              {tab.title}
            </MenuItem>
          ))}
        </Select>
      ) : (
        <StyledTabs
          orientation="vertical"
          variant="scrollable"
          value={value}
          onChange={handleChange}
          aria-label="Vertical tabs"
        >
          {tabs.map((tab, index) => (
            <Tab 
              style={{fontWeight: !tab.children ? '1000' : '', }}
              disabled={tab.disabled || !tab.children} key={index}
              label={tab.title} {...a11yProps(index)}
            />
          ))}
        </StyledTabs>
      )}
      {!isLoading && tabs.map((tab, index) => (
        <TabPanel key={index} value={value} index={index}>
          {tab.children}
        </TabPanel>
      ))}
      {isLoading && (
        <Box
          display="flex"
          alignItems="center" 
          justifyContent="center" 
          width="100%"  
          height="100%" 
          color="text.primary"  
          p={2}
        >
          <CircularProgress />
        </Box>
      )}
    </Box>
  );
};

export default PrettyTabbedView;