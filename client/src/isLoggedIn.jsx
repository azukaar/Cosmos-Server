
import * as API from './api';
import { useEffect } from 'react';
import { redirectToLocal } from './utils/indexs';

const IsLoggedIn = () => useEffect(() => {
    console.log("CHECK LOGIN")
    const urlSearch = encodeURIComponent(window.location.search);
    const redirectToURL = (window.location.pathname + urlSearch);

    API.auth.me().then((data) => {
        if(data.status != 'OK') {
            if(data.status == 'NEW_INSTALL') {
                redirectToLocal('/cosmos-ui/newInstall');
            } else if (data.status == 'error' && data.code == "HTTP004") {
                redirectToLocal('/cosmos-ui/login?redirect=' + redirectToURL);
            } else if (data.status == 'error' && data.code == "HTTP006") {
                redirectToLocal('/cosmos-ui/loginmfa?redirect=' + redirectToURL);
            } else if (data.status == 'error' && data.code == "HTTP007") {
                redirectToLocal('/cosmos-ui/newmfa?redirect=' + redirectToURL);
            }
        }
    })
}, []);

export default IsLoggedIn;