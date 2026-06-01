<#-- LKFL header-content override — injects theme toggle button
 *
 * Place theme toggle button in header area.
 * Used by Keycloak 25+ freemarker template override.
 -->
<div id="lkfl-theme-toggle-wrapper">
  <button id="lkfl-theme-toggle" type="button"
    style="width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--brand-border, #dee2e6);
           background: var(--brand-bg, #fff); cursor: pointer; font-size: 16px;
           display: flex; align-items: center; justify-content: center;
           box-shadow: var(--brand-shadow-card, 0 1px 4px rgba(0,0,0,0.06));"
    title="Переключить тему">
    <span id="lkfl-theme-icon">&#127769;</span>
  </button>
</div>
