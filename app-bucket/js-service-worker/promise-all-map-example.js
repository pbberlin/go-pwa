const url = 1;

Promise.all(
    ['/styles.css', '/script.js'].map( 
        async url => {
            return fetch(url)
            .then(
                (response) => {
                    if (!response.ok) throw Error('Not ok');
                    return cache.put(url, response);
                }
            );
       }
    ) // map
); // all



//  my fcSttc and fcReval amount to the same thing: 
//    https://web.dev/offline-cookbook/#stale-while-revalidate
const fc1 = async (evt,cch) => {
    return cch.match(evt.request)
        .then(  
             (rspCch) => {
                var promNet = fetch(evt.request)
                 .then(
                    function (netRsp) {
                        cch.put(evt.request, netRsp.clone());
                        return netRsp;
                    }
                );
                return rspCch || promNet;
            }
        );
};

let event = null;
self.addEventListener('fetch', (event) => {
    event.respondWith(
        caches.open('mysite-dynamic').then( fc1(event, cache) )
    );
});