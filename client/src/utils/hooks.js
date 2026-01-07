import React from 'react';
import { useCookies } from 'react-cookie';
import { logout } from '../api/authentication';

const isDemo = import.meta.env.MODE === 'demo';

function useClientInfos() {
  const [cookies] = useCookies(['client-infos']);

  if(isDemo) return {
    nickname: "Demo",
    role: "2",
    userRole: "2"
  };

  let clientInfos = null;
  
  try {
    // Try to parse the cookie into a JavaScript object
    clientInfos = cookies['client-infos'].split(',');

    if(clientInfos.length <= 3) {
      window.location.href = '/cosmos-ui/logout';
    }
    
    let res = {
      nickname: clientInfos[0],
      userRole: clientInfos[1],
      role: clientInfos[2]
    };

    // If role is admin, check if the timeout has expired

    if(clientInfos.length > 3) {
      let roleUntil = new Date(parseInt(clientInfos[3], 10) * 1000);
      let currentDate = new Date();

      if(roleUntil < currentDate && res.role == "2") {
        res.role = "1";
      }
    }

    return res;
  } catch (error) {
    console.error('Error parsing client-infos cookie:', error);
    return {
      nickname: "",
      role: "2"
    };
  }
}

export {
  useClientInfos
};
