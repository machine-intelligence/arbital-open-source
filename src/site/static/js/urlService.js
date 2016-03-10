"use strict";

// related to changing URLs
app.service("urlService", function($http, $location, $ngSilentLocation, $rootScope){
	var that = this;
	
	this.hasLoadedFirstPage = false;
	
	// Object reference that will change when a new URL handler is invoked
	this.currentPage = {};
	
	
	// The current page can register a pageUpdater function to handle certain URLs by modifying itself
	var pageUpdater = null;
	this.setPageUpdater = function(value){
		pageUpdater = value;
	};
	
	// Map of URL patterns to handlers
	this.urlRules = [];
	// Add a rule to handle URL changes
	this.addUrlHandler = function(urlPattern, rule) {
		var sections = urlPattern.split("/");
		// math path from beginning
		var builder = ["^"];
		var parameters = [];
		for (var n = 0; n < sections.length ; n++) {
			var section = sections[n];
			if (section == 0) {
				// ignore empty section
			}
			else if (section[0] == ":") {
				if (section.endsWith("?")) {
					// Optional parameter capture
					parameters.push(section.substring(1, section.length - 1));
					builder.push("(?:\\/([^\\/]+))?");
				}
				else {
					// Parameter capture
					parameters.push(section.substring(1));
					builder.push("\\/([^\\/]+)");
				}
			}
			else {
				// match name
				builder.push("\\/"+section);
			}
		}
		// optional trailing slash, optional query or fragment, match to end of path
		builder.push("\\/?(?:[\\?\\#].*)?$");
		var re = builder.join("");
		rule.urlPattern = new RegExp(re);
		rule.parameters = parameters;
		that.urlRules.push(rule);
	}
});
