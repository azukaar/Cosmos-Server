import React from 'react';
import { useCookies } from 'react-cookie';
import { SUDO_PERMISSIONS } from './permissions';

const isDemo = import.meta.env.MODE === 'demo';

function useClientInfos() {
  const [cookies] = useCookies(['client-infos']);

  if(isDemo) return {
    nickname: "Demo",
    hasPermission: () => true,
    hasRolePermission: () => true,
    needsSudo: false,
  };

  try {
    const parts = cookies['client-infos'].split(',');
    // Format: nickname,perm1-perm2-...,sudoUntil
    if(parts.length < 3) {
      window.location.href = '/cosmos-ui/logout';
      return {
        nickname: "",
        hasPermission: () => true,
        hasRolePermission: () => true,
        needsSudo: false,
      };
    }

    const nickname = parts[0];
    const permissions = parts[1].split('-').map(Number);
    const sudoUntilTs = parseInt(parts[2], 10);

    // Determine sudo state from timestamp
    let isSudoed = false;
    if(sudoUntilTs > 0) {
      isSudoed = new Date(sudoUntilTs * 1000) > new Date();
    }

    const hasSudoPerms = permissions.some(p => SUDO_PERMISSIONS.includes(p));
    const needsSudo = hasSudoPerms && !isSudoed;

    return {
      nickname,
      hasPermission: (perm) => {
        if(!permissions.includes(perm)) return false;
        if(SUDO_PERMISSIONS.includes(perm)) return isSudoed;
        return true;
      },
      hasRolePermission: (perm) => permissions.includes(perm),
      needsSudo,
    };
  } catch (error) {
    console.error('Error parsing client-infos cookie:', error);
    return {
      nickname: "",
      hasPermission: () => true,
      hasRolePermission: () => true,
      needsSudo: false,
    };
  }
}

export {
  useClientInfos
};
