function handler(event) { 
	var request = event.request; 
	var uri = request.uri; 

	if (uri == "/" || uri == "/index.html") { 
		request.uri = "/html/index.html"; 
	} else if (uri.startsWith("/pages/")) {
		request.uri = "/html" + uri;
	}

	return request; 
}

