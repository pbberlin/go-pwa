// "app.js"
if ('serviceWorker' in navigator) {
	// must be in root dir
	navigator.serviceWorker.register('/service-worker.js')
		.then(  (reg) => console.log("service worker - registered", reg) )	
		.catch( (err) => console.log("service worker - NOT reg'ed", err) )
}
