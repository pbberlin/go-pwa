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
