import wrap, { type ApiResponse, type ApiFetch } from './wrap';
export interface Route {
  Name: string;
}

export type Operation = 'replace' | 'move_up' | 'move_down' | 'delete' | 'add';

export default function createConfigAPI(apiFetch: ApiFetch) {
  function get() {
    return wrap(apiFetch('/cosmos/api/config', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function set(values) {
    return wrap(apiFetch('/cosmos/api/config', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function restart() {
    return apiFetch('/cosmos/api/restart', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    })
  }

  function canSendEmail() {
    return apiFetch('/cosmos/api/can-send-email', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }).then((response) => {
      return response.json();
    });
  }

  async function rawUpdateRoute(routeName: string, operation: Operation, newRoute?: Route): Promise<ApiResponse> {
    const payload = {
      routeName,
      operation,
      newRoute,
    };

    if (operation === 'replace') {
      if (!newRoute) throw new Error('newRoute must be provided for replace operation');
    }

    return wrap(apiFetch('/cosmos/api/config', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    }));
  }

  async function replaceRoute(routeName: string, newRoute: Route): Promise<ApiResponse> {
    return rawUpdateRoute(routeName, 'replace', newRoute);
  }

  async function moveRouteUp(routeName: string): Promise<ApiResponse> {
    return rawUpdateRoute(routeName, 'move_up');
  }

  async function moveRouteDown(routeName: string): Promise<ApiResponse> {
    return rawUpdateRoute(routeName, 'move_down');
  }

  async function deleteRoute(routeName: string): Promise<ApiResponse> {
    return rawUpdateRoute(routeName, 'delete');
  }
  async function addRoute(newRoute: Route): Promise<ApiResponse> {
    return rawUpdateRoute("", 'add', newRoute);
  }

  function getBackup() {
    return wrap(apiFetch('/cosmos/api/get-backup', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function getDashboard() {
    return wrap(apiFetch('/cosmos/api/dashboard', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function updateDNS(dnsConfig: {
    dnsPort?: string;
    dnsFallback?: string;
    dnsBlockBlacklist?: boolean;
    dnsAdditionalBlocklists?: string[];
    customDNSEntries?: { Type: string; Key: string; Value: string }[];
  }) {
    return wrap(apiFetch('/cosmos/api/config/dns', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(dnsConfig),
    }))
  }

  return {
    get,
    set,
    restart,
    rawUpdateRoute,
    replaceRoute,
    moveRouteUp,
    moveRouteDown,
    deleteRoute,
    addRoute,
    canSendEmail,
    getBackup,
    getDashboard,
    updateDNS,
  };
}
