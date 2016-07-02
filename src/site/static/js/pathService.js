'use strict';

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
		if (!stateService.path.pageIdsToInsert) {
			stateService.path.pageIdsToInsert = {};
		}
		stateService.path.pageIdsToInsert[index] = pageIds;
	};

	// Change the progress of the current path
	this.updateProgress = function(progress) {
		var pageIdsToInsert = [];
		if (stateService.path.pageIdsToInsert) {
			for (var index in stateService.path.pageIdsToInsert) {
				pageIdsToInsert = pageIdsToInsert.concat(stateService.path.pageIdsToInsert[index]);
			}
		}
		var params = {
			id: stateService.path.id,
			progress: progress,
			pageIdsToInsert: pageIdsToInsert,
			isFinished: stateService.path.isFinished,
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
	};

	// Go to the page that the path's progress says we should be on
	this.goToPathPage = function() {
		var path = stateService.path;
		if (path.progress >= 0) {
			var url = urlService.getPageUrl(path.pageIds[path.progress], {pathInstanceId: path.id});
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
		return path.pageIds[path.progress] == pageService.getCurrentPageId();
	};

	// Return true iff the user is at the end of the path and there are no more pages left.
	this.showFinish = function() {
		return that.isOnPath() && that.pageExtensionLength() <= 0 &&
			stateService.path.progress >= stateService.path.pageIds.length - 1;
	};

	// Return the number of pages that will be added to the path.
	this.pageExtensionLength = function() {
		if (!that.isOnPath()) return 0;
		if (!stateService.path.pageIdsToInsert) return 0;
		var count = 0;
		for (var index in stateService.path.pageIdsToInsert) {
			count += stateService.path.pageIdsToInsert[index].length;
		}
		return count;
	};
});
