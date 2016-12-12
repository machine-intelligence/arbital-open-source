'use strict';

import app from './angular.ts';

import {isIntIdValid} from './util.ts';

// editService provides functions for working with edits.
app.service('editService', function($http, $compile, $location, $rootScope, $interval, analyticsService, stateService, pageService, userService, urlService) {
	var that = this;

	this.maxQuestionTextLength = 1000;

	// Set up vote types.
	this.voteTypes = {
		probability: 'Probability',
		approval: 'Approval',
	};
	this.nullableVoteTypes = angular.extend({'': '-'}, this.voteTypes);

	// Set up sort types.
	this.sortTypes = {
		likes: 'By likes',
		recentFirst: 'Recent first',
		oldestFirst: 'Oldest first',
		alphabetical: 'Alphabetically',
	};

	// Helper function for /editPage/. Computes the data to submit via AJAX.
	this.computeSavePageData = function(page) {
		var data: any = {
			pageId: page.pageId,
			prevEdit: page.currentEdit,
			currentEdit: page.currentEdit,
			title: page.title,
			clickbait: page.clickbait,
			text: page.text,
			snapshotText: page.snapshotText,
			editSummary: page.newEditSummary,
		};
		if (page.isQuestion()) {
			data.text = data.text.length > that.maxQuestionTextLength ? data.text.slice(-that.maxQuestionTextLength) : data.text;
		}
		if (page.anchorContext) {
			data.anchorContext = page.anchorContext;
			data.anchorText = page.anchorText;
			data.anchorOffset = page.anchorOffset;
		}
		return data;
	};
});
