/* 
from https://googlechrome.github.io/samples/service-worker/custom-offline-page/
    https://developers.google.com/web/ilt/pwa/caching-files-with-service-worker

fetch including cookies: 
  fetch(url, {credentials: 'include'})

non CORS fail by default; avoid by
  new Request(urlToPrefetch, { mode: 'no-cors' }

*/

// time of start of program
let tmSt = new Date().getTime();

const tmSince = () => {
  const tm = new Date().getTime();
  return `${tm - tmSt}`;
}

const tmReset = () => {
  tmSt = new Date().getTime();
}



const VS = "{{.Version}}"; // version - also forcing update
const CACHE_KEY = `static-resources-${VS}`;
const STATIC = {
  "image": true,
  "style": true,
  "script": true,
  // "font": true,
  // "video": true,
}


const cacheNaviResps = true; // cache navigational responses

const reqOpts = {
  cache: "reload",   // => force fetching from network; not from html browser cache
  method: "GET",
  // headers: new Headers({ 'Content-Type':  'application/json' }),
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

// on failure: go to chrome://serviceworker-internals and check "Open DevTools window and pause
self.addEventListener('install', (evt) => {
  console.log(`sw-${VS} - install  - start ${tmSince()}ms`);


  const fc = async () => {
    const cch = await caches.open(CACHE_KEY);

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

  evt.waitUntil( fc() );

  // event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );

  console.log(`sw-${VS} - install  - stop  ${tmSince()}ms`);
});

// cleanup previous service worker version caches
//   dont block - prevents page loads
//   https://www.youtube.com/watch?v=k1eoekN3nkA
self.addEventListener('activate', (evt) => {
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
      .filter(  key => key !== CACHE_KEY   ) // return true to remove this cache
      .map(     key => caches.delete(key) )
    );
  };

  evt.waitUntil( fc1() );
  evt.waitUntil( fc2() );

  // instantly taking control over page
  self.clients.claim();

  console.log(`sw-${VS} - activate - stop  ${tmSince()}ms`);
});

self.addEventListener('fetch', (evt) => {

  tmReset();

  const fcDoc = async () => {

    if (1>2) {
      console.log(evt.request.url, evt.request.method, evt.request.headers, evt.request.body);
      const cch = await caches.open(CACHE_KEY);
      const rsp = await cch.match('/pets.json');
      console.log(`    rsp pets is ${rsp}`);
    }

    try {

      // try navigation preload
      //  https://developers.google.com/web/updates/2017/02/navigation-preload
      const preRsp = await evt.preloadResponse; // preload response
      if (preRsp) {
        if (!preRsp.ok) throw Error("preRsp status code not 200-299");
        console.log(`sw-${VS} - fetch - prel  ${tmSince()}ms - preRsp ${preRsp}`);
        if (cacheNaviResps) {
          const cch = await caches.open(CACHE_KEY);
          cch.put(evt.request.url, preRsp.clone()); // response is a stream - browser and cache will consume the response
        }
        return preRsp;
      }

      // try network
      const netRsp = await fetch(evt.request);  // network response
      if (!netRsp.ok) throw Error("netRsp status code not 200-299");
      console.log(`sw-${VS} - fetch - net   ${tmSince()}ms - netRsp ${netRsp}`);
      if (cacheNaviResps) {
        const cch = await caches.open(CACHE_KEY);
        // cch.add(netRsp);
        cch.put(evt.request.url, netRsp.clone());
      }
      return netRsp;

    } catch (error) {
      // on network errors
      // not on resp codes 4xx or 5xx
      // codes 4xx or 5xx jump here via if (!rsp.ok) throw...
      console.log(`sw-${VS} - fetch - error ${tmSince()}ms - ${error}`);

      const cch = await caches.open(CACHE_KEY);
      const rsp = await cch.match(evt.request, matchOpts);
      if (rsp) {
        console.log(`sw-${VS} - fetch - cache ${tmSince()}ms - cachedResp ${rsp.url}`);
        return rsp;
      } else {
        const anotherRsp = new Response( '<p>Neither network nor cache available</p>',  { headers: { 'Content-Type': 'text/html' } });
        // todo respond offline.html
        return anotherRsp;
      }
    }
  };

  const fcReval = async () => {
    const cch = await caches.open(CACHE_KEY);
    const rsp = await fetch(evt.request);
    cch.put(evt.request.url, rsp); // no cloning necessary for revalidation
    console.log(`    static rvl - ${evt.request.url} - ${tmSince()}ms`); 
  }

  /* 
    serve from cache - and revalidate asynchroneously

  */
  const fcSttc = async () => {
    try {

      const rspCch = await caches.match(evt.request);
      if (rspCch) {
        Promise.resolve().then( fcReval() );
        console.log(`    static cch - ${evt.request.url} - ${tmSince()}ms`);
        return rspCch;
      } 
      const rspNet = await fetch(evt.request);

      {
        const cch = await caches.open(CACHE_KEY);
        cch.put(evt.request.url, rspNet.clone()); // response is a stream - browser and cache will consume the response
      }
      console.log(`    static net - ${evt.request.url} - ${tmSince()}ms`);
      return rspNet;


    } catch (error) {
      console.log(`sw-${VS} - fetch static - error ${tmSince()}ms - ${error}`);
    }

  };



  // https://medium.com/dev-channel/service-worker-caching-strategies-based-on-request-types-57411dd7652c
  const dest = evt.request.destination;
 

  if (evt.request.mode === 'navigate') { // only HTML pages
    // console.log(`sw-${VS} - fetch - navi start ${tmSince()}ms - ${evt.request.url}`);
    evt.respondWith( fcDoc() );
    console.log(`sw-${VS} - fetch - navi stop  ${tmSince()}ms - ${evt.request.url}`);
  } else if ( STATIC[dest] ) {      
    evt.respondWith( fcSttc() );      
    console.log(`sw-${VS} - fetch - sttc stop  - dest ${dest} - ${evt.request.url} - mode ${evt.request.mode}`);
  }

  // ...
  // other fetch handlers ... =>  event.respondWith()
  // ...
  // default browser fetch behaviour without service worker involvement


});


