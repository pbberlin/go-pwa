// from https://googlechrome.github.io/samples/service-worker/custom-offline-page/


const VS = "{{.Version}}"; // version - only for forcing update

const MY_CACHE_1 = `offline_${VS}`;
const MY_CACHE_2 = `static-resources-v${VS}`;

const STATIC_RESS   = [
  {{.ListOfFiles}}
];

self.addEventListener('install', (event) => {
  console.log(`service worker ${VS} - installed`);

  event.waitUntil(
    (  async()  =>  {
      caches.open(MY_CACHE_2)
        .then(myCache2 => {
          if (true) {
            myCache2.addAll(STATIC_RESS); // no return - error does not prevent entire install
          }
          if (true) {
            // return myCache2.addAll("must.css"); // addAll returns promise - error prevents install
          }
          console.log(`service worker ${VS} - preloading myCache2 finish`);
        });
    })()
  );

  /* 
  event.waitUntil((async () => {
    const myCache1 = await caches.open(MY_CACHE_1);
    await myCache1.add(new Request('offline.html', { cache: 'reload' }));  // {cache: 'reload'} => force fetching from network; not from cache
    console.log(`service worker ${VS} - preloading myCache1 finish`);
  })());


  event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );

  */


});

// cleanup previous service worker version caches
//   dont block - prevents page loads
//   https://www.youtube.com/watch?v=k1eoekN3nkA
self.addEventListener('activate', (event) => {
  console.log(`service worker ${VS} - activated`);

  event.waitUntil((async () => {
    // https://developers.google.com/web/updates/2017/02/navigation-preload
    if ('navigationPreload' in self.registration) {
      await self.registration.navigationPreload.enable();
    }
  })());

  // activated service worker to take control of the page immediately.
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
        // triggered on exceptions; mostly on network error
        // *not* triggered for HTTP response codes 4xx or 5xx
        console.log(`service worker ${VS} - network fetch fail - ${error} - trying offline.`);

        const myCache2   = await caches.open(MY_CACHE_2);
        const cachedResp = await myCache2.match(event.request);
        return cachedResp;

        /* 
        caches.open(MY_CACHE_2)
          .then(myCache2 => {
            myCache2.match(event.request).then(cachedResp => {
              console.log(`service worker ${VS} - cachedResp ${cachedResp}`);
              return cachedResp;
            })
          });

         */
      }
    })());
  }

  // ...
  // other fetch handlers ... =>  event.respondWith()
  // ...
  // default browser fetch behaviour without service worker involvement
});


