'use strict';

// Contains all the services.
app.service('arb', function(autocompleteService, diffService, markService, masteryService, pageService, pathService, popoverService, stateService, userService, urlService) {
	var that = this;

	that.autocompleteService = arb.autocompleteService;
	that.diffService = arb.diffService;
	that.markService = arb.markService;
	that.markdownService = arb.markdownService;
	that.masteryService = arb.masteryService;
	that.pageService = arb.pageService;
	that.pathService = arb.pathService;
	that.popoverService = arb.popoverService;
	that.stateService = stateService.
	that.userService = arb.userService;
	that.urlService = arb.urlService;

	this.isTouchDevice = 'ontouchstart' in window || // works in most browsers
		(navigator.MaxTouchPoints > 0) ||
		(navigator.msMaxTouchPoints > 0);
};
