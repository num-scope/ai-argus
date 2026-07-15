(() => {
  const protocol = document.querySelector("[data-protocol]");
  const customFields = document.querySelector("[data-custom-fields]");
  const targetAdvanced = document.querySelector("[data-target-advanced]");
  if (protocol && customFields) {
    const syncProtocol = () => {
      const custom = protocol.value === "custom";
      customFields.hidden = !custom;
      if (custom && targetAdvanced) targetAdvanced.open = true;
    };
    protocol.addEventListener("change", syncProtocol);
    syncProtocol();
  }

  const runPage = document.querySelector("[data-run-page]");
  if (!runPage) return;
  const active = ["queued", "warming", "running"].includes(runPage.dataset.runStatus);
  if (!active) return;

  const storageKey = `argus-scroll-${runPage.dataset.runId}`;
  const savedPosition = sessionStorage.getItem(storageKey);
  if (savedPosition !== null) {
    requestAnimationFrame(() => window.scrollTo(0, Number(savedPosition)));
    sessionStorage.removeItem(storageKey);
  }
  window.setTimeout(() => {
    if (document.visibilityState !== "visible") return;
    sessionStorage.setItem(storageKey, String(window.scrollY));
    window.location.reload();
  }, 2000);
})();
