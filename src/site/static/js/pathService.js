'use strict';

// Takes care of all path related functionality
app.service('pathService', function($http, $compile, $location, $mdToast, $rootScope, $interval, userService, urlService) {
	var that = this;

	// This object is set when the user is learning / on a path.
	this.path = undefined;

	// Stop and forget the current path.
	this.abandonPath = function() {
		Cookies.remove('path');
		this.path = undefined;
	};

	// Update the path variables.
	$rootScope.$watch(function() {
		return $location.absUrl() + '|' + (that.primaryPage ? that.primaryPage.pageId : '');
	}, function() {
		that.path = undefined;
		that.path = Cookies.getJSON('path');
		if (!that.path || !that.primaryPage) return;

		// Check if the user is learning
		var currentPageId = that.getCurrentPageId();
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
};
