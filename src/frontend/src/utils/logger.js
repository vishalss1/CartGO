/** 
 * System Logger
 * Wraps console methods to ensure they only execute in Development mode.
 * Silences all output in Production to prevent data leaking and log noise.
 */

const isDev = import.meta.env.DEV;

export const logger = {
  info: (...args) => {
    if (isDev) console.log(...args);
  },
  warn: (...args) => {
    if (isDev) console.warn(...args);
  },
  error: (...args) => {
    // We allow errors to reach console in production for easier user-reported debugging,
    // but we ensure they are prefix-matched for system tracking.
    console.error("[System Error]", ...args);
  },
  group: (label) => {
    if (isDev) console.group(label);
  },
  groupEnd: () => {
    if (isDev) console.groupEnd();
  }
};
