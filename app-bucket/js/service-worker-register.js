// "app.js"
if ('serviceWorker' in navigator) {

	// window load event to keep the page load performant
	window.addEventListener('load', () => {
		// must be in root dir
		navigator.serviceWorker.register('/service-worker.js')
			.then((reg)  => console.log("service worker - registered", {reg}))  // {} leads to shortened dump
			.catch((err) => console.log("service worker - NOT reg'ed",  err ));
	});
}