import * as API from "./api/index"

async function isLoggedIn() {
    const urlSearch = encodeURIComponent(window.location.search);
    const redirectToURL = (window.location.pathname + urlSearch);

    const data = await API.auth.me();

    if (data.status != 'OK') {
        if (data.status == 'NEW_INSTALL') {
            return '/cosmos-ui/newInstall';
        } else if (data.status == 'error' && data.code == "HTTP004") {
            return '/cosmos-ui/login?redirect=' + redirectToURL;
        } else if (data.status == 'error' && data.code == "HTTP006") {
            return '/cosmos-ui/loginmfa?redirect=' + redirectToURL;
        } else if (data.status == 'error' && data.code == "HTTP007") {
            return '/cosmos-ui/newmfa?redirect=' + redirectToURL;
        }
    } else return data.status
};

export default isLoggedIn