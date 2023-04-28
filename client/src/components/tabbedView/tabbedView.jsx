import React, { useState } from 'react';
import { Box, Tab, Tabs, Typography, MenuItem, Select, useMediaQuery } from '@mui/material';
import { styled } from '@mui/system';

const StyledTabs = styled(Tabs)`
  border-right: 1px solid ${({ theme }) => theme.palette.divider};
`;

const TabPanel = (props) => {
  const { children, value, index, ...other } = props;

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
        <Box p={3}>
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

const PrettyTabbedView = ({ tabs }) => {
  const [value, setValue] = useState(0);
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));

  const handleChange = (event, newValue) => {
    setValue(newValue);
  };

  const handleSelectChange = (event) => {
    setValue(event.target.value);
  };

  return (
    <Box display="flex" height="100%" flexDirection={isMobile ? 'column' : 'row'}>
      {isMobile ? (
        <Select value={value} onChange={handleSelectChange} sx={{ minWidth: 120 }}>
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
            <Tab key={index} label={tab.title} {...a11yProps(index)} />
          ))}
        </StyledTabs>
      )}
      {tabs.map((tab, index) => (
        <TabPanel key={index} value={value} index={index}>
          {tab.children}
        </TabPanel>
      ))}
    </Box>
  );
};

export default PrettyTabbedView;