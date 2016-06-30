'use strict';

var isTouchDevice = 'ontouchstart' in window || // works in most browsers
	(navigator.MaxTouchPoints > 0) ||
	(navigator.msMaxTouchPoints > 0);

// Contains all the services.
app.service('arb', function(analyticsService, autocompleteService, diffService, markService, markdownService,
	masteryService, pageService, pathService, popoverService, popupService, stateService, userService, urlService,
	signupService, likeService) {
	var that = this;

	that.analyticsService = analyticsService;
	that.autocompleteService = autocompleteService;
	that.diffService = diffService;
	that.likeService = likeService;
	that.markService = markService;
	that.markdownService = markdownService;
	that.masteryService = masteryService;
	that.pageService = pageService;
	that.pathService = pathService;
	that.popoverService = popoverService;
	that.popupService = popupService;
	that.signupService = signupService;
	that.stateService = stateService;
	that.userService = userService;
	that.urlService = urlService;

	this.isTouchDevice = isTouchDevice;
	this.versionUrl = versionUrl;
});
