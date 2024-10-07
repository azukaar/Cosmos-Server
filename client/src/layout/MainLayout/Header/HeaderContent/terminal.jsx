import React, { useEffect } from "react";
import { BorderOuterOutlined, ClockCircleOutlined, ConsoleSqlOutlined, MonitorOutlined, WindowsOutlined } from "@ant-design/icons";
import { Badge, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, IconButton, Tooltip } from "@mui/material";
import TerminalComponent from "../../../../components/terminal";

import * as API from '../../../../api';
import { useTranslation } from "react-i18next";

const TerminalHeader = () => {
	const [open, setOpen] = React.useState(false);
	const [status, setStatus] = React.useState(false);
  const { t, Trans } = useTranslation();

	useEffect(() => {
    API.getStatus().then((res) => {
      setStatus(res.data);
    });
	}, []);

	const onopen = () => {
			setOpen((prevOpen) => !prevOpen);
	};

	return (
		<>
	  <IconButton
				disableRipple
				color="secondary"
				sx={{ color: 'text.primary' }}
				aria-label="open profile"
				aria-haspopup="true"
				onClick={onopen}
				disabled={!status || status.containerized}
		>
			<Tooltip title={(!status || status.containerized) ?
				t('mgmt.servapps.containers.terminal.disabled') :
				t('mgmt.servapps.containers.terminal.enabled')
			 }>
				<BorderOuterOutlined />
			</Tooltip>
		</IconButton>
		
		<Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth={true}>
			<DialogTitle>{t('mgmt.servapps.containers.terminal.enabled')}</DialogTitle>
			<DialogContent>
          <DialogContentText>
						<div style={{display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column'}}>
							<TerminalComponent refresh={() => {}} connectButtons={
								[
									{
										label: t('global.connect'),
										onClick: (connect) => connect(() => API.terminal(), 'Server Terminal')
									},
								]} />
						</div>
					</DialogContentText>
        </DialogContent>
        <DialogActions>
            <Button onClick={() => {
              setOpen(false)
            }}>Close</Button>
        </DialogActions>
		</Dialog>
		</>
	);
}

export default TerminalHeader;