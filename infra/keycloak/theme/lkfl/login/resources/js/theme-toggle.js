/* LKFL Theme Toggle — cookie-based dark/light mode
 *
 * Priority: cookie > system preference
 * Cookie: lkfl-theme (dark|light|auto)
 * Loaded via theme.properties scripts directive.
 */
(function () {
  'use strict';

  var COOKIE_NAME = 'lkfl-theme';
  var COOKIE_DAYS = 365;

  function setCookie(name, value, days) {
    var expires = new Date(Date.now() + days * 864e5).toUTCString();
    document.cookie = name + '=' + value + '; expires=' + expires + '; path=/; SameSite=Lax';
  }

  function getCookie(name) {
    var match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return match ? match[2] : null;
  }

  function isSystemDark() {
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
  }

  function applyTheme(dark) {
    var icon = document.getElementById('lkfl-theme-icon');
    if (dark) {
      document.documentElement.classList.add('theme-dark');
      document.documentElement.classList.remove('theme-light');
      if (icon) icon.textContent = '\u2600\ufe0f';
    } else {
      document.documentElement.classList.remove('theme-dark');
      document.documentElement.classList.add('theme-light');
      if (icon) icon.textContent = '\ud83c\udf19';
    }
  }

  function getEffectiveMode() {
    var cookie = getCookie(COOKIE_NAME);
    if (cookie === 'dark') return 'dark';
    if (cookie === 'light') return 'light';
    // auto (default)
    return isSystemDark() ? 'dark' : 'light';
  }

  function cycleMode() {
    var current = getEffectiveMode();
    var next;
    if (current === 'light') {
      next = 'dark';
    } else {
      next = 'light';
    }
    setCookie(COOKIE_NAME, next, COOKIE_DAYS);
    applyTheme(next === 'dark');
  }

  // Initial apply
  applyTheme(getEffectiveMode() === 'dark');

  // Button handler
  var btn = document.getElementById('lkfl-theme-toggle');
  if (btn) {
    btn.addEventListener('click', cycleMode);
  }

  // System preference listener (only if cookie = auto)
  if (window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function (e) {
      var cookie = getCookie(COOKIE_NAME);
      if (!cookie || cookie === 'auto') {
        applyTheme(e.matches);
      }
    });
  }
})();
