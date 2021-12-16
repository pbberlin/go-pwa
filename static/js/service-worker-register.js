// "app.js"
if ('serviceWorker' in navigator) {
	// service-worker.js must be in root dir
	navigator.serviceWorker.register('/service-worker.js')
		.then( () => console.log("service worker registered") 
		);	
}
