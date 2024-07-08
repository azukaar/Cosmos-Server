
import * as API from './api';

async function isLoggedIn() {
    const urlSearch = encodeURIComponent(window.location.search);
    const redirectToURL = (window.location.pathname + urlSearch);
    const data = await API.auth.me();

    if (data.status == 'NEW_INSTALL') {
        return '/resios-ui/newInstall';
    } else if (data.status == 'error' && data.code == "HTTP004") {
        return '/resios-ui/login?redirect=' + redirectToURL;
    } else if (data.status == 'error' && data.code == "HTTP006") {
        return '/resios-ui/loginmfa?redirect=' + redirectToURL;
    } else if (data.status == 'error' && data.code == "HTTP007") {
        return '/resios-ui/newmfa?redirect=' + redirectToURL;
    } else if (data.status == 'OK') {
        return data.status
    } else {
        console.warn(`Status "${data.status}" does not have a navigation handler, will be interpreted as OK!`)
        return 'OK'
    }
};

export default isLoggedIn;
