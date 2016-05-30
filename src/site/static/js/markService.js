'use strict';

// Manages marks
app.service('markService', function(stateService) {
	var that = this;

	// Map of all loaded marks: mark id -> mark object.
	this.markMap = {};

	// Call this to process data we received from the server.
	var postDataCallback = function(data) {
		if (data.resetEverything) {
			that.markMap = {};
		}

		// Populate marks map.
		var markData = data.marks;
		for (var id in markData) {
			stateService.smartAddToMap(that.markMap, markData[id], id);
		}
	};
	stateService.addPostDataCallback('markService', postDataCallback);

	// Create a new mark.
	this.newMark = function(params, successFn, errorFn) {
		stateService.postData('/newMark/', params, successFn, errorFn);
	};

	this.updateMark = function(params, successFn, errorFn) {
		stateService.postData('/updateMark/', params, successFn, errorFn);
	};

	this.resolveMark = function(params, successFn, errorFn) {
		stateService.postData('/resolveMark/', params, successFn, errorFn);
	};

	// Load all marks for a given page.
	this.loadMarks = function(params, successFn, errorFn) {
		stateService.postData('/json/marks/', params, successFn, errorFn);
	};
});
