import { Button } from "@mui/material";
import { useTheme } from '@mui/material/styles';

export const randomString = (length) => {
  let text = "";
  const possible =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  for (let i = 0; i < length; i++)
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  return text;
}

export function isDomain(hostname) {
  // Regular expression to check if it's an IP address
  const ipPattern = /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/;

  // Remove port if there is one
  hostname = hostname.replace(/:\d+$/, '');

  // Check if the hostname is an IP address
  if (ipPattern.test(hostname)) {
      return false;
  }

  // Check if the hostname is "localhost"
  if (hostname === 'localhost') {
      return false;
  }

  return true;
}

export const debounce = (func, wait) => {
    let timeout;
    return function (...args) {
      const context = this;
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(context, args), wait);
    };
  };

export const redirectTo = (url) => {
  window.location.href = url;
}

export const redirectToLocal = (url) => {
  let redirectUrl = new URL(url, window.location.href);
  let currentLocation = window.location;
  if (redirectUrl.origin != currentLocation.origin){
    throw new Error("URL must be local");
  }
  window.location.href = url;
}

// Validates a crontab expression. Accepts:
//   - 5-field standard:  minute hour dom month dow
//   - 6-field seconds:   second minute hour dom month dow
//   - descriptors:       @daily, @hourly, @yearly, @weekly, @monthly, @every 10m
//   - timezone prefix:   CRON_TZ=Europe/Paris <expr>
// Returns true when the field count and values are valid cron syntax.
export const isValidCrontab = (crontab) => {
  if (!crontab) return false;
  let expr = crontab.trim();
  if (expr.startsWith('@')) return true;

  // strip a leading TZ= / CRON_TZ= prefix
  const m = expr.match(/^(TZ|CRON_TZ)=[^\s]+\s+(.*)$/);
  if (m) expr = m[2].trim();

  const parts = expr.split(/\s+/).filter(Boolean);
  if (parts.length !== 5 && parts.length !== 6) return false;

  const ranges = {
    sec: [0, 59],
    min: [0, 59],
    hour: [0, 23],
    dom: [1, 31],
    month: [1, 12],
    dow: [0, 7], // 0 and 7 both mean Sunday
  };
  const names = parts.length === 6
    ? ['sec','min','hour','dom','month','dow']
    : ['min','hour','dom','month','dow'];

  const parseRange = (field) => {
    if (field === '*') return true;
    // steps: */5 or 1-10/2 or 5/15
    const base = field.split('/')[0];
    if (base === '*') return true;
    if (base.includes(',')) return base.split(',').every(parseRange);
    if (base.includes('-')) {
      const [a, b] = base.split('-');
      return !isNaN(a) && !isNaN(b) && parseInt(a) >= 0 && parseInt(b) >= 0 && parseInt(a) <= parseInt(b);
    }
    return !isNaN(base);
  };

  return parts.every((p, i) => {
    if (!parseRange(p)) return false;
    // validate literal values are within range (ignore wildcards/steps/ranges)
    const range = ranges[names[i]];
    const vals = p.split('/')[0];
    if (vals.includes(',')) {
      return vals.split(',').every(v => {
        if (v.includes('-')) return true; // range checked above
        const n = parseInt(v);
        return !isNaN(n) && n >= range[0] && n <= range[1];
      });
    }
    if (vals.includes('-') || vals === '*') return true;
    const n = parseInt(vals);
    return !isNaN(n) && n >= range[0] && n <= range[1];
  });
};

export const crontabToText = (crontab, t) => {
  if (!crontab) return t('mgmt.cron.invalidCron');

  let expr = crontab.trim();
  let tzPrefix = '';
  const m = expr.match(/^(TZ|CRON_TZ)=([^\s]+)\s+(.*)$/);
  if (m) {
    tzPrefix = `${m[1]}=${m[2]} `;
    expr = m[3].trim();
  }

  // Descriptors
  if (expr.startsWith('@')) {
    const map = {
      '@yearly': t('mgmt.cron.descriptor.yearly'),
      '@annually': t('mgmt.cron.descriptor.yearly'),
      '@monthly': t('mgmt.cron.descriptor.monthly'),
      '@weekly': t('mgmt.cron.descriptor.weekly'),
      '@daily': t('mgmt.cron.descriptor.daily'),
      '@midnight': t('mgmt.cron.descriptor.daily'),
      '@hourly': t('mgmt.cron.descriptor.hourly'),
      '@every': t('mgmt.cron.descriptor.every'),
    };
    const key = expr.startsWith('@every ') ? '@every' : expr;
    return tzPrefix + (map[key] ? (key === '@every' ? `${map[key]} ${expr.slice(7)}` : map[key]) : expr);
  }

  const parts = expr.split(/\s+/).filter(Boolean);
  if (parts.length !== 5 && parts.length !== 6) {
    return t('mgmt.cron.invalidCron');
  }

  // Canonical 6-field view: seconds first. 5-field gets an implicit "0" seconds.
  const [second, minute, hour, dayOfMonth, month, dayOfWeek] = parts.length === 6 ? parts : ['0', ...parts];

  const parseField = (field, unit = "", date=false) => {
    const count = (nb) => {
      if(date) {
        nb = parseInt(nb);

        if(nb === 1 || nb === 21 || nb === 31) {
          return `${nb}st `
        }
        if(nb === 2 || nb === 22) {
          return `${nb}nd `
        }
        if(nb === 3 || nb === 23) {
          return `${nb}rd `
        }
        return `${nb}th `
      }
      return `${nb} `;
    }
    
    const plur = (field, force=false) => {
      if (!force && date) {
        return '';
      }
      if (field === '1' || field === '0') {
        return '';
      } else {
        return 's';
      }
    }

    if (field === '*') {
        return `every ${unit}s`;
    } else if (field.includes('-')) {
        const [start, end] = field.split('-');
        return `from ${start}${count(start)}${unit}${plur(field)} to ${end}${count(end)}${unit}${plur(field)}`;
    } else if (field.includes(',')) {
        return `${field.split(',')
        .join(', ')}`;
    } else if (field.includes('/')) {
        const [start, step] = field.split('/');
        return `every ${step} ${unit}${plur(step, true)}, starting at ${unit} ${start}`;
    } else {
        return `${count(field)}${unit}${plur(field)}`;
    }
  };

  let text = '';
  let timeText = '';
  let timeTextArray = [];
  let dateText = '';
  let dateTextArray = [];

  // Handle date fields
  if (dayOfMonth !== '*') {
      const dayOfMonthText = parseField(dayOfMonth, "day", true);
      dateTextArray.push(`${dayOfMonthText} of the month`)
  }

  if (month !== '*') {
      const monthText = parseField(month, "month", true);
      dateTextArray.push(`${monthText}`);
  }

  if (dayOfWeek !== '*') {
      const dayOfWeekText = parseField(dayOfWeek, "day", true);
      dateTextArray.push(`${dayOfWeekText} of the week`);
  }
  
  if (hour !== '*') {
    timeTextArray.push(parseField(hour, "hour"));
  }
  if (minute !== '*') {
    timeTextArray.push(`${parseField(minute, "min")}`);
  }
  if (second !== '*') {
    timeTextArray.push(`${parseField(second, "sec")}`);
  }


  if (dateTextArray.length > 0) {
    dateText = `${dateTextArray.join(' and ')}`;
    if(!dateText.startsWith("from")) {
      dateText = "On " + dateText
    }
  }
  if (timeTextArray.length > 0) {
    timeText = ` at ${timeTextArray.join(' and ')}`;
    if(dateText == "") {
      timeText = "Every day " + timeText
    }
  }

  let intro = '';
  // get first * field
  if (second === '*') {
    intro = 'Every second, ';
  } else if (minute === '*') {
    intro = 'Every minute , ';
  } else if (hour === '*') {
    intro = 'Every hour, ';
  } else if (dayOfMonth === '*') {
    intro = 'Every day, ';
  } else if (month === '*') {
    intro = 'Every month, ';
  } else if (dayOfWeek === '*') {
    intro = 'Every day, ';
  }


  return tzPrefix + intro + text + dateText + timeText;
}

export const PascalToSnake = (str) => {
  return str.replace(/[\w]([A-Z])/g, function(m) {
    return m[0] + "_" + m[1];
  }).toLowerCase();
}

export const getCurrencyFromLanguage = () => {
  let language = window.navigator.userLanguage || window.navigator.language;
  // language = language.split('-')[0]; // Get language code without region
  
  const currencyMap = {
    en: 'USD', // English (assuming US English as default)
    'en-US': 'USD', // US English
    'en-GB': 'GBP', // British English
    de: 'EUR', // German
    fr: 'EUR', // French
    es: 'EUR', // Spanish
    it: 'EUR', // Italian
    pt: 'EUR', // Portuguese
    nl: 'EUR', // Dutch
  };

  return currencyMap[language] || 'USD'; // Default to USD if no match
};


export const DarkModeSwitch = ({light, dark}) => {
  const theme = useTheme();
  const isLight = theme.palette.mode === 'light';

  return isLight ? light : dark;
}