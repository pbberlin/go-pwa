# HTTPS server prototype

<img src="./appbucket/img/mascot.webp" style="float: left; width:8%; min-width: 160px; max-width: 22%; margin-right:5%; margin-bottom: 2%;"> 

Keeping up with

* HTTP/2

* HTTPS and HTTP in coexistence or redirecting

* Let's encrypt certification

* Localhost certificate based on [Filipo Valsordas tool](https://github.com/FiloSottile/mkcert)

* Content Security Policies - CSP  
  against CSRF

* Compress static content  
  on the fly - or precompress on app start

* HTML template for PWA

* PWA manifest

* PWA service worker
  * Register + Install
    * Prime the cache
  * Activate (`update on reload` or re-open browser)
  * Fetch
    * Updating cache

Trying to combine it into a slim `go` program,  
which is testable using [Google Lighthouse](https://github.com/GoogleChrome/Lighthouse).

## Next steps

* Service worker should be a template,  
having the version compiled into

* CSS and JS files should have a "version" directory,  
such as  
`/js/32168/service-worker-registration.js`
