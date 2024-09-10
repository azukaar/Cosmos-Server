import { useEffect, useState } from "react";
import { toUTC } from "./components/utils";
import * as API from '../../api';
import { Button, CircularProgress, Stack, TextField } from "@mui/material";
import { CosmosCollapse, CosmosSelect } from "../config/users/formShortcuts";
import MainCard from '../../components/MainCard';
import { register, format } from 'timeago.js';
import de from "timeago.js/lib/lang/de";
import { ExclamationOutlined, SettingOutlined } from "@ant-design/icons";
import { Alert } from "@mui/material";
import { DownloadFile } from "../../api/downloadButton";
import { Trans, useTranslation } from 'react-i18next';

const EventsExplorer = ({from, to, xAxis, zoom, slot, initLevel, initSearch = ''}) => {
	register('de', de);
	const { t, i18n } = useTranslation();
	const [events, setEvents] = useState([]);
	const [loading, setLoading] = useState(true);
	const [search, setSearch] = useState(initSearch);
	const [debouncedSearch, setDebouncedSearch] = useState(search);
	const [total, setTotal] = useState(0);
	const [remains, setRemains] = useState(0);
	const [page, setPage] = useState(0);
	const [logLevel, setLogLevel] = useState(initLevel || 'success');

	if(typeof from != 'undefined' && typeof to != 'undefined') {
		from = new Date(from);
		to = new Date(to);
	} else {
		const zoomedXAxis = xAxis
		.filter((date, index) => {
			if (zoom && zoom.xaxis && zoom.xaxis.min && zoom.xaxis.max) {
				return index >= zoom.xaxis.min && index <= zoom.xaxis.max;
			}
			return true;
		})
		.map((date) => {
			if (slot === 'hourly' || slot === 'daily') {
				let k = toUTC(date, slot === 'hourly');
				return k;
			} else {
				let realIndex = xAxis.length - 1 - date
				return realIndex;
			}
		})

		const firstItem = zoomedXAxis[0];
		const lastItem = zoomedXAxis[zoomedXAxis.length - 1];

		if (slot === 'hourly' || slot === 'daily') {
			from = new Date(firstItem);
			to = new Date(lastItem);

			if (slot === 'daily') {
				to.setHours(23);
				to.setMinutes(59);
				to.setSeconds(59);
				to.setMilliseconds(999);
			} else {
				to.setMinutes(to.getMinutes() + 59);
				to.setSeconds(to.getSeconds() + 59);
				to.setMilliseconds(to.getMilliseconds() + 999);
			}
		} else {
			const now = new Date();
			// round to 30 seconds
			now.setSeconds(now.getSeconds() - now.getSeconds() % 30);
			// remove microseconds
			now.setMilliseconds(0);

			from = new Date(now.getTime() - lastItem * 30000);
			to = new Date(now.getTime() - firstItem * 30000);

			// add 29 seconds to the end
			to.setSeconds(to.getSeconds() + 29);
		}
	}

	const refresh = (_page) => {
		setLoading(true);
		let _search = debouncedSearch;
		let _query = "";
		if (_search.startsWith('{') || _search.startsWith('[')) {
			_search = ""
			_query = debouncedSearch;
		}
		return API.metrics.events(from.toISOString(), to.toISOString(), _search, _query, _page, logLevel).then((res) => {
			setEvents(res.data);
			setLoading(false);
			setTotal(res.total);
			setRemains(res.total - res.data.length);
		});
	}

	useEffect(() => {
		setPage(0);
		if (debouncedSearch.length === 0 || debouncedSearch.length > 2) {
			refresh("");
		}
	}, [debouncedSearch, xAxis, zoom, slot, logLevel]);

	useEffect(() => {
		// Set a timeout to update debounced search after 1 second
		const handler = setTimeout(() => {
			setDebouncedSearch(search);
		}, 500);
	
		// Clear the timeout if search changes before the 1 second has passed
		return () => {
			clearTimeout(handler);
		};
	}, [search]); // Only re-run if search changes

	useEffect(() => {
		setLoading(true);
		refresh(page);
	}, [page]);

	return (<div>
		<Stack spacing={2} direction="column" style={{width: '100%'}}>
			<Stack spacing={2} direction="row" style={{width: '100%'}}>
				<div>
					<Button variant='contained' onClick={() => {
						refresh("");
					}} style={{height: '42px'}}>{t('global.refresh')}</Button>
				</div>
				<div>
					<DownloadFile filename='events-export.json' content={
						JSON.stringify(events, null, 2)
					}  style={{height: '42px'}} label='export' />
				</div>
				<div>
					<CosmosSelect
						name={'level'}
						formik={{
							values: {
								level: logLevel
							},
							touched: {},
							errors: {},
							setFieldValue: () => {},
							handleChange: () => {}
						}}
						options={[
							['debug', 'Debug'],
							['info', 'Info'],
							['success', 'Success'],
							['warning', 'Warning'],
							['important', 'Important'],
							['error', 'Error'],
						]}
						onChange={(e) => {
							setLogLevel(e.target.value);
						}}
						style={{
							width: '100px',
							margin:0,
						}}
					/>
				</div>
				<TextField fullWidth value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('navigation.monitoring.events.searchInput.searchPlaceholder')} />
			</Stack>
			<div>
				{!loading &&
					<Trans i18nKey="navigation.monitoring.events.eventsFound" count={total} values={{from: from.toLocaleString(), to: to.toLocaleString()}}/>
				}
			</div>
			<div>
				{events && <Stack spacing={1}>
					{events.map((event) => {
						return <div key={event.id} style={eventStyle(event)}>
							<CosmosCollapse title={
								<Alert severity={event.level} icon={
									event.level == "debug" ? <SettingOutlined /> : event.level == "important" ? <ExclamationOutlined /> : undefined
								}>
									<div style={{fontWeight: 'bold', fontSize: '120%'}}>{event.label}</div>
									<div>{(new Date(event.date)).toLocaleString()} - {format(event.date, i18n.resolvedLanguage)}</div>
									<div>{event.eventId} - {event.object}</div>
								</Alert>}>
								<div style={{overflow: 'auto'}}>
										<pre style={{
											whiteSpace: 'pre-wrap',
											overflowX: 'auto',
											maxWidth: '100%',
											maxHeight: '400px',
									}}>
										{JSON.stringify(event, null, 2)}
									</pre>
								</div>
							</CosmosCollapse>
						</div>
					})}
					{loading && <div style={{textAlign: "center"}}>
						<CircularProgress />
					</div>}
					{!loading && (remains > 0) && <MainCard>
						<Button variant='contained' fullWidth onClick={() => {
							// set page to last element's id
							setPage(events[events.length - 1].id);
						}}>{t('navigation.monitoring.events.loadMoreButton')}</Button>
					</MainCard>}
				</Stack>}
			</div>
		</Stack>
		
	</div>);
}

export default EventsExplorer;

const eventStyle = (event) => ({
	padding: 4,
	borderRadius: 4,
});