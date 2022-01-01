# HTTPS server prototype

<img src="./app-bucket/img/mascot.webp" style="float: left; width:8%; min-width: 160px; max-width: 22%; margin-right:5%; margin-bottom: 2%;"> 

Combining the most advanced `golang` techniques  
into a [Google Lighthouse](https://github.com/GoogleChrome/Lighthouse) compatible web app.

* HTTP/2

* HTTPS and HTTP in coexistence or redirecting

* Let's encrypt certification

* Localhost certificate based on [Filipo Valsordas tool](https://github.com/FiloSottile/mkcert)

* Content Security Policies (`CSP`)  
  against CSRF

* Precompress static content on app start  
  or on the fly

* Coherent new-versioning of HTML, JS, CSS  
  after server side changes

* HTML template for PWA

* PWA manifest,  
  PWA service worker
  * Register + Install
    * Prime the cache
  * Activate (`update on reload` or re-open browser)
  * Fetch
    * Updating cache

## Next steps

* Service worker should be a template,  
having the version compiled into
