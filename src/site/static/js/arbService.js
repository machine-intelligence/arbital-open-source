'use strict';

// Contains all the services.
app.service('arb', function(autocompleteService, diffService, markService, markdownService, masteryService, pageService, pathService, popoverService, stateService, userService, urlService) {
	var that = this;

	that.autocompleteService = autocompleteService;
	that.diffService = diffService;
	that.markService = markService;
	that.markdownService = markdownService;
	that.masteryService = masteryService;
	that.pageService = pageService;
	that.pathService = pathService;
	that.popoverService = popoverService;
	that.stateService = stateService;
	that.userService = userService;
	that.urlService = urlService;

	this.isTouchDevice = 'ontouchstart' in window || // works in most browsers
		(navigator.MaxTouchPoints > 0) ||
		(navigator.msMaxTouchPoints > 0);
});
