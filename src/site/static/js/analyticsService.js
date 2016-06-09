'use strict';

// arb.analyticsService is a wrapper for Google Analytics
app.service('analyticsService', function($http, $location, stateService) {
	var that = this;

	// This is called to set the user id.
	this.setUserId = function(userId) {
		if (!isLive()) return;
		ga('set', 'userId', userId);
	};

	// This is calld when a user goes to a new page.
	this.reportPageView = function() {
		if (!isLive()) return;
		// Set the page, which which will be included with all future events.
		ga('set', 'page', $location.path());
		// Send "pageview" event, since we switched new a new view
		ga('send', 'pageview');
	};

	// Called when a user clicks on an edit link
	this.reportEditLinkClick = function(event) {
		if (!isLive()) return;
		ga('send', {
			hitType: 'event',
			eventCategory: 'Edit',
			eventAction: 'linkClick',
			eventLabel: event.target.href,
			eventValue: 1,
		});
	};

	// Called when a user submits a page to domain
	this.reportPageToDomainSubmission = function() {
		if (!isLive()) return;
		ga('send', {
			hitType: 'event',
			eventCategory: 'Page',
			eventAction: 'submitToDomain',
			eventLabel: '1lw',
			eventValue: 1,
		});
	};
});
