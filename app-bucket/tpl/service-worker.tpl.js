// from https://googlechrome.github.io/samples/service-worker/custom-offline-page/

// time of start of program
let tmSt = new Date().getTime();

const tmSince = () => {
  const tm = new Date().getTime();
  let diff = `${tm - tmSt}`;
  return diff;
}



const VS = "{{.Version}}"; // version - also forcing update



// const MY_C_old = `offline_${VS}`;
const MY_CACHE = `static-resources-${VS}`;

const STATIC_RESS   = [
  {{.ListOfFiles}}
];

self.addEventListener('install', (event) => {


  // {cache: 'reload'} => force fetching from network; not from html browser cache

  console.log(`service worker ${VS} - install - start ${tmSince()}ms`);



  const cOpts1 = {
    cache:  "reload",
    method: "GET",
    // headers: new Headers({ 'Content-Type': 'application/json' }),
  };


  const fc = async () => {
    const cch = await caches.open(MY_CACHE);

    let proms = [];
    STATIC_RESS.forEach(res => {
      proms.push(  cch.add(  new Request(res, cOpts1) ) );
    });
    const allPr = await Promise.all(proms);
    console.log(`service worker ${VS} - preloading status ${allPr}`);

    cch.put('/pets.json', new Response('{"tom": "cat", "jerry": "mouse"}') ); // no options possible

  };


  // event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );
  event.waitUntil( fc() );

  console.log(`service worker ${VS} - install - stop  ${tmSince()}ms`);




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

        const myCache2   = await caches.open(MY_CACHE);
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


