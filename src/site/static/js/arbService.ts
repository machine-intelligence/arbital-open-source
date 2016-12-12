'use strict';

import app from './angular.ts';
import {isTouchDevice,isIntIdValid} from './util.ts';

// Contains all the services.
app.service('arb', function(analyticsService, autocompleteService, diffService, markService, markdownService,
	masteryService, pageService, editService, pathService, popoverService, popupService, stateService, userService, urlService,
	signupService) {
	var that = this;

	that.analyticsService = analyticsService;
	that.autocompleteService = autocompleteService;
	that.diffService = diffService;
	that.markService = markService;
	that.markdownService = markdownService;
	that.masteryService = masteryService;
	that.pageService = pageService;
	that.editService = editService;
	that.pathService = pathService;
	that.popoverService = popoverService;
	that.popupService = popupService;
	that.signupService = signupService;
	that.stateService = stateService;
	that.userService = userService;
	that.urlService = urlService;

	this.isTouchDevice = isTouchDevice;
	this.isIntIdValid = isIntIdValid;
	this.versionUrl = versionUrl;
});
