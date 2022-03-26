# Golang PWA prototype

<img src="./app-bucket/img/mascot.webp" style="float: left; width:20%; min-width: 140px; max-width: 20%; margin-right:5%; margin-bottom: 2%;"> 

[![coverage](https://github.com/pbberlin/go-pwa/actions/workflows/codecov.yml/badge.svg)](https://github.com/pbberlin/go-pwa/actions/workflows/codecov.yml)

Combining the most advanced `golang` techniques  
into a [Google Lighthouse](https://github.com/GoogleChrome/Lighthouse) approved web app

* HTTP/2

* Let's encrypt certification

* Localhost certificate based on [Filipo Valsordas tool](https://github.com/FiloSottile/mkcert)

* HTTP redirecting or co-existing

* Content Security Policies (`CSP`)  
  against CSRF

* Consistent versioning of HTML, JS, CSS, IMG,  
  _and_ service worker caching

* Adding a new version at any time by admin http request,  
  while older version files remain accessible

* Make changes to your app at any time,  
  without breaking service worker caching

* Server side gzip precompression of CSS and JS files;  
  integrated with version creation

* Fully developed PWA HTML template

* Fully developed PWA manifest

* PWA service worker with `cache-first` for static files

* Fallback to `/offline.html` for unprecedented user experience

  * PWA service worker register and install

  * PWA service worker pre-caching on install

  * PWA service worker fetch with `cache-first`

## Javascript docs

Finally some comprehensive JS docs

<https://javascript.info/indexeddb>

<https://javascript.info/microtask-queue>

## Package static

Preparing and serving static files.  

The package assumes a directory `./app-bucket`  
containing directories of files by mime-types.  
`/css`, `/js`, `/img`...

The package also supports typical special files:  
`robots.txt`, `service-worker.js`, `favicon.ico`  
being served under special URIs.

The package takes care of

* Execution of templates for special files
  * Service worker pre-caching
  * `manifest.json` icon files
  * Javascript database versioning

* Mime types by configuration
* HTTP caching by configuration

* Consistent versioning

* Gzipping by configuration
* Handler funcs for HTTP request serving
* Registering routes with a http.ServeMux

Template execution allows custom funcs for arbitrary dynamic preparations.

A few template execution funcs are provided,  
to prepare Google PWA config files dynamically  
from whatever is in the directories under `./app-bucket`.

All file preparation logic is put together in the HTTP handle func PrepareStatic(...).
Thus you whenever you changed any static file contents,
call PrepareStatic(), and you get a _consistent_ new version of all static files,
and you force your HTTP client (aka browser) to load

Todo:

* Make the config loadable via JSON
* Javascript templating is done in a highly inappropriate way; cannot get idiomatic way to work
* Markdown with some pre-processing is missing

## Gorm

* `save` first updates by primary key. Then selects whether the record exists. And if not, inserts.

* `create` inserts. On conflict adds DB specific jargon for upsert/merge. 

* Caveat: uniqueness indexes should include the deletion date column

### Relations / Associations

* 