export function registerOfflineSupport() {
  void requestPersistentStorage();

  if (!("serviceWorker" in navigator) || !import.meta.env.PROD || !window.isSecureContext) {
    return;
  }

  window.addEventListener("load", () => {
    void registerServiceWorker();
  });
}

async function registerServiceWorker() {
  try {
    const registration = await navigator.serviceWorker.register(`${import.meta.env.BASE_URL}sw.js`, {
      scope: import.meta.env.BASE_URL,
      updateViaCache: "none"
    });

    await registration.update();
  } catch (err) {
    console.warn("Offline cache registration failed", err);
  }
}

async function requestPersistentStorage() {
  if (!("storage" in navigator) || !navigator.storage.persist || !navigator.storage.persisted) {
    return;
  }

  try {
    if (!(await navigator.storage.persisted())) {
      await navigator.storage.persist();
    }
  } catch (err) {
    console.warn("Persistent browser storage request failed", err);
  }
}
