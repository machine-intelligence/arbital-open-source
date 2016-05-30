'use strict';

// Manages common state for services
// NOTE: stateService should be includable by any service that relies on any kind of app state
app.service('stateService', function($http, $compile, $location, $mdToast, $rootScope, $interval, popupService) {
	var that = this;

	// Id of the private group we are in. (Corresponds to the subdomain).
	this.privateGroupId = '';

	// Set to true when user has some text selected.
	this.lensTextSelected = false;

	// Should we show all marks for the currently selected lens. By default we just
	// show the current user's.
	this.showAllMarks = false;

	// Should we show editor comments for the currently selected lens.
	var showEditorComments = false;
	this.getShowEditorComments = function() {
		return showEditorComments || !this.primaryPage;
	};
	this.setShowEditorComments = function(value) {
		showEditorComments = value;
	};

	// Add the given page to the global pageMap. If the page with the same id
	// already exists, we do a clever merge.
	var isValueTruthy = function(v) {
		// "0" is falsy
		if (v === '0') {
			return false;
		}
		// Empty array is falsy.
		if ($.isArray(v) && v.length == 0) {
			return false;
		}
		// Empty object is falsy.
		if ($.isPlainObject(v) && $.isEmptyObject(v)) {
			return false;
		}
		return !!v;
	};
	this.smartMerge = function(oldV, newV) {
		if (!isValueTruthy(newV)) {
			return oldV;
		}
		return newV;
	};
	// Use our smart merge technique to add a new object to existing object map.
	this.smartAddToMap = function(map, newObject, newObjectId) {
		var oldObject = map[newObjectId];
		if (newObject === oldObject) return;
		if (oldObject === undefined) {
			map[newObjectId] = newObject;
			return;
		}
		// Merge each variable.
		for (var k in oldObject) {
			oldObject[k] = that.smartMerge(oldObject[k], newObject[k]);
		}
	};

	// ================== Standard POSTing to server =====================
	// Functions to call when we get data from the server.
	var postDataCallbacks = {};
	this.addPostDataCallback = function(key, fn) {
		postDataCallbacks[key] = fn;
	};
	this.processServerData = function(data) {
		for (var key in postDataCallbacks) {
			postDataCallbacks[key](data);
		}
	};

	// Load data from the server and process it.
	// options = {
	//	callCallbacks: if true, call postDataCallbacks
	// }
	this.postDataWithOptions = function(url, params, options, successFn, errorFn) {
		$http({method: 'POST', url: url, data: JSON.stringify(params)})
			.success(function(data) {
				console.log(url + ' data:'); console.dir(data);
				if (options.callCallbacks) {
					that.processServerData(data);
				}
				if (successFn) {
					successFn(data);
				}
			})
			.error(function(data) {
				console.error('Error getting data from ' + url); console.dir(data);
				var silentError = false;
				if (errorFn) {
					silentError = errorFn(data);
				}
				if (!silentError) {
					popupService.showToast({text: 'Error getting data from ' + url + ': ' + data, isError: true});
				}
			});
	};
	this.postData = function(url, params, successFn, errorFn) {
		that.postDataWithOptions(url, params, {callCallbacks: true}, successFn, errorFn);
	};
	this.postDataWithoutProcessing = function(url, params, successFn, errorFn) {
		that.postDataWithOptions(url, params, {callCallbacks: false}, successFn, errorFn);
	};

	// ========================== Mathjax caching ===============================
	// Mathjax text -> rendered html string cache
	// mathjax expression -> {
	//   html: rendered html
	//   style: computed style for the div / span
	// }
	var mathjaxCache = {};
	var mathjaxRecency = {}; // key -> timestamp

	// Update the timestamp on a cached mathjax.
	this.getMathjaxCacheValue = function(text) {
		if (!(text in mathjaxCache)) return null;
		mathjaxRecency[text] = new Date().getTime();
		return mathjaxCache[text];
	};

	// Add the given {text: value} pair to mathjax cache.
	this.cacheMathjax = function(text, value) {
		mathjaxCache[text] = value;
		mathjaxRecency[text] = new Date().getTime();

		var cacheSize = 0;
		var minTime = new Date().getTime();
		var minKey;
		for (var key in mathjaxRecency) {
			var time = mathjaxRecency[key];
			if (minTime > time) {
				minTime = time;
				minKey = key;
			}
			cacheSize++;
		}
		if (cacheSize > 100) {
			delete mathjaxRecency[minKey];
			delete mathjaxCache[minKey];
		}
	};
});
