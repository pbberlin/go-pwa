// from https://googlechrome.github.io/samples/service-worker/custom-offline-page/
//      https://developers.google.com/web/ilt/pwa/caching-files-with-service-worker

// time of start of program
let tmSt = new Date().getTime();

const tmSince = () => {
  const tm = new Date().getTime();
  let diff = `${tm - tmSt}`;
  return diff;
}



const VS = "{{.Version}}"; // version - also forcing update
const MY_CACHE = `static-resources-${VS}`;

const reqOpts = {
  cache: "reload",   // => force fetching from network; not from html browser cache
  method: "GET",
  // headers: new Headers({ 'Content-Type': 'application/json' }),
  // headers: new Headers({ 'Cache-Control': 'max-age=31536000' }),
  // headers: new Headers({ 'Cache-Control': 'no-cache' }),

};

const matchOpts = {
  ignoreVary:   true, // ignore differences in Headers
  ignoreMethod: true, // ignore differences in HTTP methods
  ignoreSearch: true  // ignore differences in query strings
};


const STATIC_RESS   = [
  {{.ListOfFiles}}
];

self.addEventListener('install', (event) => {
  console.log(`sw-${VS} - install  - start ${tmSince()}ms`);


  const fc = async () => {
    const cch = await caches.open(MY_CACHE);

    let proms = [];
    STATIC_RESS.forEach( res => {
      // if (!rsp.ok) throw Error('Not ok');
      // return cch.put(url, rsp);
      proms.push(  cch.add(  new Request(res, reqOpts) ) );
    });
    const allPr = await Promise.all(proms);
    console.log(`sw-${VS} - install  - preld ${tmSince()}ms ${allPr}`);

    cch.put('/pets.json', new Response('{"tom": "cat", "jerry": "mouse"}') );

  };

  event.waitUntil( fc() );

  // event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );

  console.log(`sw-${VS} - install  - stop  ${tmSince()}ms`);
});

// cleanup previous service worker version caches
//   dont block - prevents page loads
//   https://www.youtube.com/watch?v=k1eoekN3nkA
self.addEventListener('activate', (event) => {
  console.log(`sw-${VS} - activate - start ${tmSince()}ms`);

  const fc1 = async () => {
    // https://developers.google.com/web/updates/2017/02/navigation-preload
    if ('navigationPreload' in self.registration) {
      await self.registration.navigationPreload.enable();
    }
  };

  const fc2 = async () => {
    const keys = await caches.keys();
    return await Promise.all(
      keys
      .filter(  key => key !== MY_CACHE   ) // return true if you want to remove this cache
      .map(     key => caches.delete(key) )
    );

  };

  event.waitUntil( fc1() );
  event.waitUntil( fc2() );

  // instantly taking control over page
  self.clients.claim();

  console.log(`sw-${VS} - activate - stop  ${tmSince()}ms`);
});

self.addEventListener('fetch', (event) => {

  const fc = async () => {

    if (1>2) {
      const cch = await caches.open(MY_CACHE);
      const rsp = await cch.match('/pets.json');
      console.log(`    rsp pets is ${rsp}`);
    }

    try {

      // try navigation preload
      const preRsp = await event.preloadResponse; // preload response
      if (preRsp) {
        if (!preRsp.ok) throw Error("preRsp status code not 200-299");
        console.log(`sw-${VS} - fetch - prel  ${tmSince()}ms - preRsp ${preRsp}`);
        return preRsp;
      }

      // try network
      const netRsp = await fetch(event.request);  // network response
      if (!netRsp.ok) throw Error("netRsp status code not 200-299");
      console.log(`sw-${VS} - fetch - net   ${tmSince()}ms - netRsp ${netRsp}`);
      return netRsp;

    } catch (error) {
      // on network errors
      // *not* on resp codes 4xx or 5xx
      console.log(`sw-${VS} - fetch - error ${tmSince()}ms - ${error}`);

      const cch = await caches.open(MY_CACHE);
      const rsp = await cch.match(event.request, matchOpts);
      console.log(`sw-${VS} - fetch - cache ${tmSince()}ms - cachedResp ${rsp}`);
      return rsp;
    }
  };



  if (event.request.mode === 'navigate') { // only HTML pages
    console.log(`sw-${VS} - fetch - navi start ${tmSince()}ms - ${event.request.url}`);
    event.respondWith( fc() );
    console.log(`sw-${VS} - fetch - navi stop  ${tmSince()}ms - ${event.request.url}`);
  }

  // ...
  // other fetch handlers ... =>  event.respondWith()
  // ...
  // default browser fetch behaviour without service worker involvement


});


