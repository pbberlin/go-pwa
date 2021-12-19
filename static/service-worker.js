// from https://googlechrome.github.io/samples/service-worker/custom-offline-page/


const VS = 8; // version - only for forcing update

const MY_CACHE_1 = 'offline';  // example 1


const MY_CACHE_2  = 'static-resources'; // example 2
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

  // example for external res
  // 'https://fonts.google.com/icon?family=Material+Icons',

];

self.addEventListener('install', (event) => {
  console.log(`service worker ${VS} - installed`);

  event.waitUntil(  
    (  async()  =>  { 
      caches.open(MY_CACHE_2)
        .then(myCache2 => {
          myCache2.addAll(STATIC_RESS); // why no await?
          console.log(`service worker ${VS} - preloading myCache2 finish`);
        });
    })()  
  );


  event.waitUntil((async () => {    
    const myCache1 = await caches.open(MY_CACHE_1);
    await myCache1.add(new Request('offline.html', { cache: 'reload' }));  // {cache: 'reload'} => force fetching from network; not from cache
    console.log(`service worker ${VS} - preloading myCache1 finish`);
  })());


  event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );


  

});

self.addEventListener('activate', (event) => {
  console.log(`service worker ${VS} - activated`);

  event.waitUntil((async () => {
    // https://developers.google.com/web/updates/2017/02/navigation-preload
    if ('navigationPreload' in self.registration) {
      await self.registration.navigationPreload.enable();
    }
  })());

  // Tell the active service worker to take control of the page immediately.
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {

  if (event.request.mode === 'navigate') { // only HTML pages
    console.log(`service worker ${VS} - fetch ${event.request.url}`);

    event.respondWith((async () => {
      try {
        const preloadResponse = await event.preloadResponse; // try navigation preload first
        if (preloadResponse) {
          return preloadResponse;
        }
        const networkResponse = await fetch(event.request);
        return networkResponse;

      } catch (error) {
        // triggered on exceptions; likely due to network error.
        // *not* triggered for HTTP response codes 4xx or 5xx
        console.log(`service worker ${VS} - network fetch fail - ${error} - trying offline.`);

        const myCache1 = await caches.open(MY_CACHE_1);
        const cachedResponse = await myCache1.match('offline.html');
        return cachedResponse;
      }
    })());
  }

  // ... 
  // other fetch handlers ... =>  event.respondWith()
  // ...
  // default browser fetch behaviour without service worker involvement
});


