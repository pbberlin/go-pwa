// from https://googlechrome.github.io/samples/service-worker/custom-offline-page/

// time of start of program
let tmSt = new Date().getTime();

const tmSince = () => {
  const tm = new Date().getTime();
  let diff = `${tm - tmSt}`;
  return diff;
}



const VS = "{{.Version}}"; // version - also forcing update
const MY_CACHE = `static-resources-${VS}`;

const STATIC_RESS   = [
  {{.ListOfFiles}}
];

self.addEventListener('install', (event) => {
  console.log(`sw-${VS} - install - start ${tmSince()}ms`);

  const cOpts1 = {
    cache:  "reload",   // => force fetching from network; not from html browser cache
    method: "GET",
    ignoreVary: true,   // ignore differences in Headers
    ignoreMethod: true, // ignore differences in HTTP methods
    ignoreSearch: true  // ignore differences in query strings
    // headers: new Headers({ 'Content-Type': 'application/json' }),
  };

  const fc = async () => {
    const cch = await caches.open(MY_CACHE);

    let proms = [];
    STATIC_RESS.forEach(res => {
      proms.push(  cch.add(  new Request(res, cOpts1) ) );
    });
    const allPr = await Promise.all(proms);
    console.log(`sw-${VS} - install - preloading status ${allPr}`);

    cch.put('/pets.json', new Response('{"tom": "cat", "jerry": "mouse"}') ); // no options possible

  };

  event.waitUntil( fc() );

  // event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );

  console.log(`sw-${VS} - install - stop  ${tmSince()}ms`);
});

// cleanup previous service worker version caches
//   dont block - prevents page loads
//   https://www.youtube.com/watch?v=k1eoekN3nkA
self.addEventListener('activate', (event) => {
  console.log(`sw-${VS} - activate - start ${tmSince()}ms`);

  const fc = async () => {
    // https://developers.google.com/web/updates/2017/02/navigation-preload
    if ('navigationPreload' in self.registration) {
      await self.registration.navigationPreload.enable();
    }
  };

  event.waitUntil(fc());

  // instantly taking control over page
  self.clients.claim();

  console.log(`sw-${VS} - activate - stop  ${tmSince()}ms`);
});

self.addEventListener('fetch', (event) => {
  console.log(`sw-${VS} - fetch - start ${tmSince()}ms - ${event.request.url}`);


  const fc = async () => {

    if (1>2) {
      const cch = await caches.open(MY_CACHE);
      const rsp = await cch.match('/pets.json');
      console.log(`    rsp pets is ${rsp}`);
    }

    try {
      const preloadResponse = await event.preloadResponse; // try navigation preload first
      if (preloadResponse) {
        return preloadResponse;
      }
      const networkResponse = await fetch(event.request);
      return networkResponse;

    } catch (error) {
      // on network errors
      // *not* triggered for HTTP resp codes 4xx or 5xx
      console.log(`sw-${VS} - fetch - network fail - ${error} - trying offline`);

      const cch = await caches.open(MY_CACHE);
      const rsp = await cch.match(event.request);
      console.log(`sw-${VS} - fetch - cache rsp ${tmSince()}ms - cachedResp ${rsp}`);
      return rsp;
    }
  };



  if (event.request.mode === 'navigate') { // only HTML pages
    console.log(`sw-${VS} - fetch - navi  ${tmSince()}ms - ${event.request.url}`);
    event.respondWith( fc() );
  }

  // ...
  // other fetch handlers ... =>  event.respondWith()
  // ...
  // default browser fetch behaviour without service worker involvement

  console.log(`sw-${VS} - fetch - stop  ${tmSince()}ms - ${event.request.url}`);

});


