// from https://googlechrome.github.io/samples/service-worker/custom-offline-page/


const VS = 2; // version - only for forcing update

const CACHE_DEFAULT = 'offline';  // example 1


const CACHE_STATIC  = 'static-resources'; // example 2
const STATIC_RESS   = [
  '/css/progress-bar-2.css',
  '/css/styles-mobile.css',
  '/css/styles-quest-no-site-specified-a.css',
  '/css/styles-quest.css',
  '/css/styles.css',
  '/js/menu-and-form-control-keys.js',
  '/js/service-worker-register.js',
  '/js/validation.js',
  '/img/icon-072.png',
  '/img/icon-096.png',
  '/img/icon-128.png',
  '/img/icon-144.png',
  '/img/icon-192.png',
  '/img/icon-384.png',
  '/img/icon-512.png',
  '/img/mascot-squared.png',
  '/img/mascot.png',  
];

self.addEventListener('install', (event) => {
  console.log(`service worker ${VS} - installed`);


  event.waitUntil((async () => {
    // {cache: 'reload'} => new response pulled from network; not from HTTP cache
    const cch1 = await caches.open(CACHE_DEFAULT);
    await cch1.add(new Request('offline.html', {cache: 'reload'}));
    console.log(`service worker ${VS} - install - caching 1`);
  })());


  caches.open(CACHE_STATIC).then(cacheStatic => {
    cacheStatic.addAll(STATIC_RESS);
    console.log(`service worker ${VS} - install - caching 2`);

  });


});

self.addEventListener('activate', (event) => {
  console.log(`service worker ${VS} - activated`);

  event.waitUntil((async () => {
    // Enable navigation preload if it's supported.
    // See https://developers.google.com/web/updates/2017/02/navigation-preload
    if ('navigationPreload' in self.registration) {
      await self.registration.navigationPreload.enable();
    }
  })());

  // Tell the active service worker to take control of the page immediately.
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {

  // call event.respondWith() for HTML page navigation
  if (event.request.mode === 'navigate') {  
    console.log(`service worker ${VS} - fetch ${event.request.url}`);

    event.respondWith((async () => {
      try {
        // First, try using navigation preload response if it's supported.
        const preloadResponse = await event.preloadResponse;
        if (preloadResponse) {
          return preloadResponse;
        }

        const networkResponse = await fetch(event.request);
        return networkResponse;
      } catch (error) {
        // catch is only triggered if an exception is thrown,
        // which is likely due to a network error.
        // catch() will *not* be called for HTTP response codes 4xx or 5xx       
        console.log(`service worker ${VS} - fetch fail - serving offline; ${error}.`);

        const cache = await caches.open(CACHE_DEFAULT);
        const cachedResponse = await cache.match('offline.html');
        return cachedResponse;
      }
    })());
  }

  // If our if() condition is false, then this fetch handler won't intercept the
  // request. If there are any other fetch handlers registered, they will get a
  // chance to call event.respondWith(). If no fetch handlers call
  // event.respondWith(), the request will be handled by the browser as if there
  // were no service worker involvement.
});


