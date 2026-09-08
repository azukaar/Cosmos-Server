import * as React from 'react';
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import { useClientInfos } from '../../utils/hooks';
import { PERM_RESOURCES_READ } from '../../utils/permissions';
import { useTranslation } from 'react-i18next';
import { useHasLicence } from '../../utils/pro';
import EventExplorerStandalone from '../dashboard/eventsExplorerStandalone';

import { ConstellationVPN } from './vpn';
import { ConstellationDNS } from './dns';

const ConstellationIndex = () => {
  const { t } = useTranslation();
  const { hasPermission } = useClientInfos();
  const isAdmin = hasPermission(PERM_RESOURCES_READ);
  const licence = useHasLicence();

  // Hold the free/paid decision until /status answers so a licensed user never sees the sales page flash.
  const freeVersion = licence === false;

  const tabs = [
    {
      title: 'VPN',
      children: <ConstellationVPN freeVersion={freeVersion} />,
      url: '/',
    },
    {
      title: 'DNS',
      children: <ConstellationDNS />,
      url: '/dns',
    },
    {
      title: t('navigation.monitoring.eventsTitle'),
      children: <EventExplorerStandalone
        initLevel="info"
        initSearch={'{"$or":[{"eventId":{"$regex":"^cosmos\\\\.constellation\\\\."}},{"object":{"$regex":"^device@"}}]}'}
      />,
      url: '/events',
    },
  ];

  if (!isAdmin || freeVersion) {
    return <ConstellationVPN freeVersion={freeVersion} />;
  }

  return <div>
    <PrettyTabbedView rootURL="/cosmos-ui/constellation" tabs={tabs} />
  </div>;
}

export default ConstellationIndex;
