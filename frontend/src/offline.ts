let registrationStarted = false;

export function registerOfflineSupport() {
  if (!("serviceWorker" in navigator) || !import.meta.env.PROD || !window.isSecureContext) {
    return;
  }

  if (registrationStarted) {
    return;
  }

  registrationStarted = true;
  if (document.readyState === "complete") {
    void registerServiceWorker();
  } else {
    window.addEventListener("load", () => {
      void registerServiceWorker();
    });
  }
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
