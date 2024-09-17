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

export const crontabToText = (crontab, t) => {
  const parts = crontab.split(' ');

  if (parts.length !== 6) {
      return t('mgmt.cron.invalidCron');
  }

  const [second, minute, hour, dayOfMonth, month, dayOfWeek] = parts;

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

  return intro + text + dateText + timeText;
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