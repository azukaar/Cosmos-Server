
import * as API from './api';
import { useEffect } from 'react';

const IsLoggedIn = () => useEffect(() => {
    console.log("CHECK LOGIN")
    const urlSearch = encodeURIComponent(window.location.search);
    const redirectTo = (window.location.pathname + urlSearch);

    API.auth.me().then((data) => {
        if(data.status != 'OK') {
            if(data.status == 'NEW_INSTALL') {
                window.location.href = '/ui/newInstall';
            } else if (data.status == 'error' && data.code == "HTTP004") {
                window.location.href = '/ui/login?redirect=' + redirectTo;
            } else if (data.status == 'error' && data.code == "HTTP006") {
                window.location.href = '/ui/loginmfa?redirect=' + redirectTo;
            } else if (data.status == 'error' && data.code == "HTTP007") {
                window.location.href = '/ui/newmfa?redirect=' + redirectTo;
            }
        }
    })
}, []);

export default IsLoggedIn;