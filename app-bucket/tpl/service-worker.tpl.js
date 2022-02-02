/*
from
  googlechrome.github.io/samples/service-worker/custom-offline-page/
  developers.google.com/web/ilt/pwa/caching-files-with-service-worker
  web.dev/offline-cookbook

fetch including cookies:
  fetch(url, {credentials: 'include'})

non CORS fail by default; avoid by
  new Request(urlToPrefetch, { mode: 'no-cors' }

*/

const VS = "{{.Version}}"; // version - also forcing update

const STATIC_TYPES = { // request.destination
  "image": true,
  "style": true,
  "script": true,
  // "font": true,
  // "video": true,
  "manifest": true,  // special for PWA manifest.json
}



const oldCons = console;  // window.console undefined in service worker
const newCons = function(oldCons) {

  var prefixes = {};

  return {
    log: function (text) {
      // 'sw-vs-476442 - fetch - sttc stop'
      // '    static cch -'

      const prefPref = `sw-${VS} -`;
      const pref1 = text.substring(0, prefPref.length);
      if (pref1 === prefPref) {
        text = text.substring(prefPref.length);
        // oldCons.log("trimmed pref1");
      }


      const pref2 = text.substring(0, 16);
      if (pref2 in prefixes) {
        prefixes[pref2]++;
      } else {
        prefixes[pref2] = 1;
      }

      // avoid escaping of < to &lt; by golang template engine
      if ( 4>prefixes[pref2] ) {
        oldCons.log( prefixes[pref2], text);
      } else {
        // oldCons.log("swallowed", prefixes[prefix],text);
      }
    },
    info: function (text) {
      oldCons.info(text);
    },
    warn: function (text) {
      oldCons.warn(text);
    },
    error: function (text) {
      oldCons.error(text);
    },

  }

}

// console = newCons(oldCons);




// time of start of program
let tmSt = new Date().getTime();

const tmSince = () => {
  const tm = new Date().getTime();
  return `${tm - tmSt}`;
}

const tmReset = () => {
  tmSt = new Date().getTime();
}

const chopHost = (url) => {
  url = url.replace(/^(AB)/, '');          // replaces "AB" only if it is at the beginning
  url = url.replace(/^(https:\/\/localhost)/, '');
  return url;
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

  async function requestBackgroundSync(tag) {
    try {
      await self.registration.sync.register(tag);
      console.log("sync - supported (from service worker)");
    } catch (err) {
      console.log(`sw-${VS} - self.registration.sync failed ${err}`);
    }
  }
  requestBackgroundSync('tag-sync-install');

  // event.waitUntil(  (  async()  =>  { console.log(`payload`); })()  );
  console.log(`sw-${VS} - install  - stop  ${tmSince()}ms`);
});

// cleanup previous service worker version caches
//   dont block - prevents page loads
//   www.youtube.com/watch?v=k1eoekN3nkA
self.addEventListener('activate', (evt) => {
  console.log(`sw-${VS} - activate - start ${tmSince()}ms`);

  const fc1 = async () => {
    // developers.google.com/web/updates/2017/02/navigation-preload
    if ('navigationPreload' in self.registration) {
      await self.registration.navigationPreload.enable();
    }
  };

  // No way for cache TTL: stackoverflow.com/questions/55729284
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


  // respond documents from net
  //   caching
  //   falling back to cache
  //   falling back offline
  const fcDoc = async () => {

    if (1>2) {
      const evtr = evt.request;
      console.log(evtr.url, evtr.method, evtr.headers, evtr.body);
      console.log(evtr.url.hostname, evtr.url.origin, evtr.url.pathname);
      const cch = await caches.open(CACHE_KEY);
      const rsp = await cch.match('/pets.json');
      console.log(`    rsp pets is ${rsp}`);
    }

    try {

      // try navigation preload
      //  developers.google.com/web/updates/2017/02/navigation-preload
      const preRsp = await evt.preloadResponse; // preload response
      if (preRsp) {
        if (!preRsp.ok) throw Error("preRsp status code not 200-299");
        console.log(`sw-${VS} - fetch - prel  ${tmSince()}ms - preRsp ${preRsp.url}`);
        if (cacheNaviResps) {
          const cch = await caches.open(CACHE_KEY);
          cch.put(evt.request.url, preRsp.clone()); // response is a stream - browser and cache will consume the response
        }
        return preRsp;
      }

      // try network
      const netRsp = await fetch(evt.request);  // network response
      if (!netRsp.ok) throw Error("netRsp status code not 200-299");
      console.log(`sw-${VS} - fetch - net   ${tmSince()}ms - netRsp ${netRsp.url}`);
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
        if (1>2) {
          const anotherRsp = new Response('<p>Neither network nor cache available</p>', { headers: { 'Content-Type': 'text/html' } });
          return anotherRsp;
        }
        return caches.match('/offline.html');

      }
    }
  };

  // revalidate
  const fcReval = async () => {
    try {
      if (!navigator.onLine) {
        return;
      }
      const cch = await caches.open(CACHE_KEY);
      const rsp = await fetch(evt.request);
      cch.put(evt.request.url, rsp); // no cloning necessary for revalidation
      console.log(`    static rvl - ${chopHost(evt.request.url)} - ${tmSince()}ms`);
    } catch (error) {
      console.log(`sw-${VS} - reval fetch - error ${tmSince()}ms - ${error} - ${chopHost(evt.request.url)}`);
    }
  }


  // serve from cache - and revalidate asynchroneously
  //   or serve from net and put into synchroneously
  //   so called "Stale-while-revalidate" - web.dev/offline-cookbook/#stale-while-revalidate
  //
  // to see the revalidated response within the same request, we need to call this from the html page
  const fcSttc = async () => {
    try {

      //
      const rspCch = await caches.match(evt.request);
      if (rspCch) {
        // Promise.resolve().then( fcReval() );  // rewritten on the next two lines
        const dummy = await Promise.resolve();
        fcReval();
        console.log(`    static cch - ${chopHost(evt.request.url)} - ${tmSince()}ms`);
        return rspCch;
      }

      // this results in chained promises fetch => cache open => cache put => return fetch
      //   we could async the cache open, cache put ops, but it does not save much
      const rspNet = await fetch(evt.request);
      const cch = await caches.open(CACHE_KEY);
      cch.put(evt.request.url, rspNet.clone()); // response is a stream - browser and cache will consume the response
      console.log(`    static net - ${chopHost(evt.request.url)} - ${tmSince()}ms`);
      return rspNet;


    } catch (error) {
      console.log(`sw-${VS} - fetch static - error ${tmSince()}ms - ${error} - ${chopHost(evt.request.url)}`);
    }

  };



  // medium.com/dev-channel/service-worker-caching-strategies-based-on-request-types-57411dd7652c
  const dest = evt.request.destination;


  if (evt.request.mode === 'navigate') { // only HTML pages
    // console.log(`sw-${VS} - fetch - navi start ${tmSince()}ms - ${chopHost(evt.request.url)}`);
    evt.respondWith( fcDoc() );
    console.log(`sw-${VS} - fetch - navi stop  ${tmSince()}ms - ${chopHost(evt.request.url)}`);
  } else if ( STATIC_TYPES[dest] ) {
    evt.respondWith( fcSttc() );
    console.log(`sw-${VS} - fetch - sttc stop  - dest ${dest} - ${chopHost(evt.request.url)} - mode ${evt.request.mode}`);
  } else {
    console.log(`sw-${VS} - fetch - unhandled  - dest ${dest} - ${chopHost(evt.request.url)} - mode ${evt.request.mode}`);
  }

  // ...default browser fetch behaviour without service worker involvement


});


importScripts(`/js/${VS}/idb.js`);
importScripts(`/js/${VS}/idb-init.js`);

// not triggered by request.mode navigate
//   https://davidwalsh.name/background-sync
self.addEventListener('sync', (evt) => {

  tmReset();

  console.log(`sw-${VS} - sync tag ${evt.tag} - start `);
  // console.log(evt);


  const pref = "tag-sync-";
  const cmp = evt.tag.substring(0, pref.length);
  if (cmp === pref) {

    console.log(`sw-${VS} - sync tag ${evt.tag} - before waitUntil `);

    evt.waitUntil( async () => {

      console.log(`sw-${VS} - sync tag ${evt.tag} - after  waitUntil `);

      const cch = await caches.open(CACHE_KEY);
      cch.add(new Request('/home-sync.html', reqOpts));

      console.log(`sw-${VS} - sync tag ${evt.tag} - db `);

      try {
        const outbx = await store.outbox('readonly');
        const msgs = await outbx.getAll();
        // send messages
        const rawResponse = await fetch('https://localhost/save-json', {
          method: 'POST',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          },
          // body: JSON.stringify({ a: 1, b: 'Textual content' }),
          body: JSON.stringify(msgs),
        });
        const rsp = await rawResponse.json();
        console.log(`save-json response: `,{rsp});

      } catch (err) {
        console.error(err);
      }

    });
  }

  console.log(`sw-${VS} - sync tag ${evt.tag} - stop `);


});


const CACHE_KEY = `static-resources-${VS}`;
