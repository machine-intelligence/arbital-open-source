'use strict';

import app from './angular.ts';

// Takes care of all path related functionality
app.service('pathService', function($http, $compile, $location, $mdToast, $rootScope, $interval, stateService, pageService, urlService) {
	var that = this;

	// This object is set when the user is learning / on a path.
	this.path = undefined;

	// Stop and forget the current path.
	this.abandonPath = function() {
		Cookies.remove('path');
		this.path = undefined;
		this.newPath = undefined;
	};

	// Update the path variables.
	$rootScope.$watch(function() {
		return $location.absUrl() + '|' + (stateService.primaryPage ? stateService.primaryPage.pageId : '');
	}, function() {
		that.path = undefined;
		that.path = Cookies.getJSON('path');
		if (!that.path || !stateService.primaryPage) return;

		// Check if the user is learning
		var currentPageId = pageService.getCurrentPageId();
		var pathPageIds = that.path.readIds;
		var currentIndex = pathPageIds.indexOf(currentPageId);
		if (currentIndex >= 0) {
			that.path.onPath = true;
			that.path.prevPageId = currentIndex > 0 ? pathPageIds[currentIndex - 1] : '';
			that.path.nextPageId = currentIndex < pathPageIds.length - 1 ? pathPageIds[currentIndex + 1] : '';
			that.path.currentPageId = currentPageId;
		} else {
			that.path.onPath = false;
			that.path.prevPageId = that.path.nextPageId = '';
		}
		Cookies.set('path', that.path);
	});

	// ============= NEW PATH STUFF =================

	// Start the path associated with the given guide for the current user
	this.startPath = function(guideId) {
		var params = {
			guideId: guideId,
		};
		stateService.postData('/json/startPath/', params, function(data) {
			stateService.path = data.result.path;
			that.goToPathPage();
		});
	};

	// Add/remove the given pageIds to the path at the current point
	this.extendPath = function(index, pageIds) {
		if (!that.isOnPath()) return;
		if (!stateService.path.pagesToInsert) {
			stateService.path.pagesToInsert = {};
		}
		var pages = [];
		for (var n = 0; n < pageIds.length; n++) {
			pages.push({
				pageId: pageIds[n],
				sourceId: pageService.getCurrentPageId(),
			});
		}
		stateService.path.pagesToInsert[index] = pages;
	};

	// Change the progress of the current path
	this.updateProgress = function(progress) {
		var path = stateService.path;
		var pagesToInsert = [];
		if (path.pagesToInsert) {
			for (var index in path.pagesToInsert) {
				pagesToInsert = pagesToInsert.concat(path.pagesToInsert[index]);
			}
		}

		// Compute pages
		var pages = path.pages.slice(0, path.progress + 1);
		pages = pages.concat(pagesToInsert);
		pages = pages.concat(path.pages.slice(path.progress + 1));

		var params = {
			id: path.id,
			progress: progress,
			pages: pages,
			isFinished: path.isFinished,
		};
		stateService.postData('/json/updatePath/', params, function(data) {
			stateService.path = data.result.path;
			that.goToPathPage();
		});
	};

	// Mark the current path as finished
	this.finishPath = function() {
		stateService.path.isFinished = true;
		that.updateProgress(stateService.path.progress);
		$location.replace().search("pathId", undefined);
	};

	// Go to the page that the path's progress says we should be on
	this.goToPathPage = function() {
		var path = stateService.path;
		if (path.progress >= 0) {
			var url = urlService.getPageUrl(path.pages[path.progress].pageId, {pathInstanceId: path.id});
		} else {
			var url = urlService.getPageUrl(path.guideId);
			stateService.path = undefined;
		}
		urlService.goToUrl(url);
	};

	// Return true iff the user is on the path.
	this.isOnPath = function() {
		var path = stateService.path;
		if (!path) return false;
		return path.pages[path.progress].pageId == pageService.getCurrentPageId();
	};

	// Return true iff the user is at the end of the path and there are no more pages left.
	this.showFinish = function() {
		return that.isOnPath() && that.pageExtensionLength() <= 0 &&
			stateService.path.progress >= stateService.path.pages.length - 1;
	};

	// Called when the primary page changes.
	this.primaryPageChanged = function() {
		if (!stateService.path) return;
		if (stateService.path.guideId == pageService.getCurrentPageId()) {
			stateService.path = undefined;
			return;
		}

		// Remove all pages that were added by the page we are currently on
		for (var n = 0; n < stateService.path.pages.length; n++) {
			stateService.path.pages = stateService.path.pages.filter(function(page) {
				return page.sourceId != pageService.getCurrentPageId();
			});
		}

		// Make sure progress value is right
		for (var n = 0; n < stateService.path.pages.length; n++) {
			if (stateService.path.pages[n].pageId == pageService.getCurrentPageId()) {
				stateService.path.progress = n;
				break;
			}
		}
	};

	// Return the number of pages that will be added to the path.
	this.pageExtensionLength = function() {
		if (!that.isOnPath()) return 0;
		if (!stateService.path.pagesToInsert) return 0;
		var count = 0;
		for (var index in stateService.path.pagesToInsert) {
			count += stateService.path.pagesToInsert[index].length;
		}
		return count;
	};

	// Return true if the user is on the path and will eventually read the given page.
	this.isBefore = function(pageId) {
		if (!that.isOnPath()) return false;
		var path = stateService.path;
		for (var n = path.progress + 1; n < path.pages.length; n++) {
			if (path.pages[n].pageId == pageId) return true;
		}
		return false;
	};

	// Return true if the user is on the path and already read the given page.
	this.isAfter = function(pageId) {
		if (!that.isOnPath()) return false;
		var path = stateService.path;
		for (var n = 0; n < path.progress; n++) {
			if (path.pages[n].pageId == pageId) return true;
		}
		return false;
	};
});
