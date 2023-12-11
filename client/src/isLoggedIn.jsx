
import * as API from './api';
import { useEffect } from 'react';
import { redirectToLocal } from './utils/indexs';
import Loadable from './components/Loadable';

const IsLoggedIn = Loadable(() => {
    const urlSearch = encodeURIComponent(window.location.search);
    const redirectToURL = (window.location.pathname + urlSearch);

    API.auth.me().then((data) => {
        if(data.status != 'OK') {
            if(data.status == 'NEW_INSTALL') {
                return '/cosmos-ui/newInstall';
            } else if (data.status == 'error' && data.code == "HTTP004") {
                return '/cosmos-ui/login?redirect=' + redirectToURL;
            } else if (data.status == 'error' && data.code == "HTTP006") {
                return '/cosmos-ui/loginmfa?redirect=' + redirectToURL;
            } else if (data.status == 'error' && data.code == "HTTP007") {
                return '/cosmos-ui/newmfa?redirect=' + redirectToURL;
            }
        }
    }).then(redirectPath => {
        if (window.location.pathname !== redirectPath)
            return redirectToLocal(redirectPath)
    })
});

export default IsLoggedIn;