import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import { join, relative, resolve, sep } from "node:path";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig, type Plugin, type ResolvedConfig } from "vite";

export default defineConfig({
  plugins: [svelte(), anibasOfflinePlugin()],
  server: {
    port: 5173
  }
});

function anibasOfflinePlugin(): Plugin {
  let config: ResolvedConfig;

  return {
    name: "anibas-offline-service-worker",
    apply: "build",
    configResolved(resolvedConfig) {
      config = resolvedConfig;
    },
    closeBundle() {
      const outDir = resolve(config.root, config.build.outDir);
      if (!existsSync(outDir)) {
        return;
      }

      const files = collectFiles(outDir).filter((file) => {
        const path = toBrowserPath(outDir, file);
        return path !== "./sw.js" && !path.endsWith(".map");
      });
      const hash = createHash("sha256");
      for (const file of files) {
        hash.update(toBrowserPath(outDir, file));
        hash.update(readFileSync(file));
      }

      const version = hash.digest("hex").slice(0, 16);
      const precacheUrls = Array.from(new Set(["./", ...files.map((file) => toBrowserPath(outDir, file))]));
      writeFileSync(join(outDir, "sw.js"), renderServiceWorker(version, precacheUrls));
    }
  };
}

function collectFiles(directory: string): string[] {
  const entries = readdirSync(directory);
  const files: string[] = [];

  for (const entry of entries) {
    const fullPath = join(directory, entry);
    const stats = statSync(fullPath);
    if (stats.isDirectory()) {
      files.push(...collectFiles(fullPath));
    } else if (stats.isFile()) {
      files.push(fullPath);
    }
  }

  return files.sort();
}

function toBrowserPath(outDir: string, file: string) {
  return `./${relative(outDir, file).split(sep).join("/")}`;
}

function renderServiceWorker(version: string, precacheUrls: string[]) {
  return `const CACHE_PREFIX = "anibas-feed-engine";
const CACHE_VERSION = ${JSON.stringify(version)};
const APP_CACHE = \`\${CACHE_PREFIX}-app-\${CACHE_VERSION}\`;
const RUNTIME_CACHE = \`\${CACHE_PREFIX}-runtime-\${CACHE_VERSION}\`;
const PRECACHE_URLS = ${JSON.stringify(precacheUrls, null, 2)};
const SCOPE_URL = self.registration.scope;
const PRECACHE_KEYS = new Set(PRECACHE_URLS.map((path) => normalizeUrl(new URL(path, SCOPE_URL).toString())));

self.addEventListener("install", (event) => {
  event.waitUntil(
    (async () => {
      const cache = await caches.open(APP_CACHE);
      await cache.addAll([...PRECACHE_KEYS]);
      await self.skipWaiting();
    })()
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      const names = await caches.keys();
      await Promise.all(
        names
          .filter((name) => name.startsWith(CACHE_PREFIX) && name !== APP_CACHE && name !== RUNTIME_CACHE)
          .map((name) => caches.delete(name))
      );
      await self.clients.claim();
    })()
  );
});

self.addEventListener("fetch", (event) => {
  const { request } = event;

  if (request.method !== "GET") {
    return;
  }

  const url = new URL(request.url);
  if (!isSameScope(url)) {
    return;
  }

  if (request.mode === "navigate") {
    event.respondWith(networkFirstNavigation(request));
    return;
  }

  if (PRECACHE_KEYS.has(normalizeUrl(request.url))) {
    event.respondWith(cacheFirst(request));
    return;
  }

  if (isStaticAsset(request, url)) {
    event.respondWith(staleWhileRevalidate(request));
  }
});

async function networkFirstNavigation(request) {
  const cache = await caches.open(APP_CACHE);

  try {
    const response = await fetch(request);
    if (response.ok) {
      await cache.put(normalizeUrl(request.url), response.clone());
    }
    return response;
  } catch {
    return (
      (await cache.match(normalizeUrl(request.url))) ||
      (await cache.match(normalizeUrl(new URL("./index.html", SCOPE_URL).toString()))) ||
      Response.error()
    );
  }
}

async function cacheFirst(request) {
  const cache = await caches.open(APP_CACHE);
  const cacheKey = normalizeUrl(request.url);
  const cached = await cache.match(cacheKey);

  if (cached) {
    return cached;
  }

  const response = await fetch(request);
  if (response.ok) {
    await cache.put(cacheKey, response.clone());
  }

  return response;
}

async function staleWhileRevalidate(request) {
  const cache = await caches.open(RUNTIME_CACHE);
  const cached = await cache.match(request, { ignoreSearch: true });
  const network = fetch(request)
    .then(async (response) => {
      if (response.ok) {
        await cache.put(request, response.clone());
      }
      return response;
    })
    .catch(() => undefined);

  return cached || (await network) || Response.error();
}

function isSameScope(url) {
  const scope = new URL(SCOPE_URL);
  return url.origin === scope.origin && url.pathname.startsWith(scope.pathname);
}

function isStaticAsset(request, url) {
  return (
    request.destination === "script" ||
    request.destination === "style" ||
    request.destination === "image" ||
    request.destination === "manifest" ||
    url.pathname.endsWith(".wasm") ||
    url.pathname.endsWith(".js") ||
    url.pathname.endsWith(".css") ||
    url.pathname.endsWith(".svg")
  );
}

function normalizeUrl(value) {
  const url = new URL(value);
  url.hash = "";
  url.search = "";
  return url.toString();
}
`;
}
