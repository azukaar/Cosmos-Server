// project import
import pages from './pages';
import dashboard from './dashboard';
import constellation from './constellation';
import support from './support';
import { isProBuild } from '../utils/pro';
import { version } from '../../../package.json';

// ==============================|| MENU ITEMS ||============================== //

const UNSTABLE_ONLY = ['registries', 'functions'];

const buildItems = () => {
    if (!isProBuild()) {
        const network = { ...constellation.children[0], title: 'menu-items.management.constellation' };
        const children = [...pages.children];
        children.splice(children.findIndex((c) => c.id === 'users'), 0, network);
        return [dashboard, { ...pages, children }, support];
    }

    const group = version.includes('-unstable') ? constellation : {
        ...constellation,
        children: constellation.children.filter((c) => !UNSTABLE_ONLY.includes(c.id)),
    };
    return [dashboard, group, pages, support];
};

const menuItems = {
    items: buildItems()
};

export default menuItems;
