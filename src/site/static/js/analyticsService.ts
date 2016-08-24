'use strict';

import app from './angular.ts';
import {isLive} from './util.ts';

declare var ga: any;
declare var heap: any;
declare var FS: any;

// arb.analyticsService is a wrapper for Google Analytics
app.service('analyticsService', function($http, $location, stateService) {
	var that = this;

	// This is called to set the user id.
	this.setUserId = function(userId) {
		if (!!userId) {
			// heap
			heap.identify(userId);

			// full story
			FS.identify(userId);
		}

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

	// Called when a user autosaves a page
	this.reportAutosave = function(charCount) {
		heap.track('autosave', {characters: charCount});
	};

	// Called when a user edits a page
	this.reportEditPageAction = function(event, action) {
		if (!isLive()) return;
		ga('send', {
			hitType: 'event',
			eventCategory: 'Edit',
			eventAction: action,
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

	// Called when a user does something with the signup dialog
	this.reportSignupAction = function(action, attemptedAction) {
		if (!isLive()) return;
		ga('send', {
			hitType: 'event',
			eventCategory: 'Signup',
			eventAction: action,
			eventLabel: attemptedAction,
			eventValue: 1,
		});
	};

	// Called when a user publishes a page
	this.reportPublishAction = function(action, pageId, length) {
		if (!isLive()) return;
		ga('send', {
			hitType: 'event',
			eventCategory: 'Publish',
			eventAction: action,
			eventLabel: pageId,
			eventValue: length,
		});
	};
});
