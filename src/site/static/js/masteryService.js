'use strict';

// Manages masteries
app.service('masteryService', function($http, $compile, $location, $mdToast, $rootScope, $interval, stateService) {
	var that = this;

	// All loaded masteries.
	this.masteryMap = {};
	// When the user answers questions or does other complex reversible actions, this map
	// allows us to store the new masteries the user acquired/lost. That way we can allow the user
	// to change their answers, without messing up the masteries they learned through other means.
	// sorted array: [map "key" -> masteryMap]
	// Array is sorted by the order in which the questions appear in the text.
	this.masteryMapList = [this.masteryMap];

	// This is called when we get data from the server.
	var postDataCallback = function(data) {
		if (data.resetEverything) {
			this.masteryMap = {};
			this.masteryMapList = [this.masteryMap];
		}

		// Populate materies map.
		var masteryData = data.masteries;
		for (var id in masteryData) {
			stateService.smartAddToMap(this.masteryMap, masteryData[id], id);
		}
	};
	stateService.addPostDataCallback('masteryService', postDataCallback);

	// Update the masteryMap. Execution happens in the order options are listed.
	// options = {
	//		delete: set these masteries to "doesn't know"
	//		wants: set these masteries to "wants"
	//		knows: set these masteries to "knows"
	//		callback: optional callback function
	// }
	this.updateMasteryMap = function(options) {
		var affectedMasteryIds = [];
		if (options.delete) {
			for (var n = 0; n < options.delete.length; n++) {
				var masteryId = options.delete[n];
				var mastery = this.masteryMap[masteryId];
				if (!mastery) continue;
				mastery.has = false;
				mastery.wants = false;
				affectedMasteryIds.push(masteryId);
			}
		}
		if (options.wants) {
			for (var n = 0; n < options.wants.length; n++) {
				var masteryId = options.wants[n];
				var mastery = this.masteryMap[masteryId];
				if (!mastery) {
					mastery = {pageId: masteryId};
					this.masteryMap[masteryId] = mastery;
				}
				mastery.has = false;
				mastery.wants = true;
				affectedMasteryIds.push(masteryId);
			}
		}
		if (options.knows) {
			for (var n = 0; n < options.knows.length; n++) {
				var masteryId = options.knows[n];
				var mastery = this.masteryMap[masteryId];
				if (!mastery) {
					mastery = {pageId: masteryId};
					this.masteryMap[masteryId] = mastery;
				}
				mastery.has = true;
				mastery.wants = false;
				affectedMasteryIds.push(masteryId);
			}
		}
		this.pushMasteriesToServer(affectedMasteryIds, options.callback);
	};

	// Compute the status of the given masteries and update the server
	this.pushMasteriesToServer = function(affectedMasteryIds, callback) {
		var addMasteries = [];
		var delMasteries = [];
		var wantsMasteries = [];
		for (var n = 0; n < affectedMasteryIds.length; n++) {
			var masteryId = affectedMasteryIds[n];
			var masteryStatus = this.getMasteryStatus(masteryId);
			if (masteryStatus === 'has') {
				addMasteries.push(masteryId);
			} else if (masteryStatus === 'wants') {
				wantsMasteries.push(masteryId);
			} else {
				delMasteries.push(masteryId);
			}
		}

		var data = {
			removeMasteries: delMasteries,
			wantsMasteries: wantsMasteries,
			addMasteries: addMasteries,
			// Note: this is a bit hacky. We should probably pass computeUnlocked explicitly
			computeUnlocked: !!callback,
			taughtBy: that.getCurrentPageId(),
		};
		if (callback) {
			stateService.postData('/updateMasteries/', data, callback);
		} else {
			stateService.postDataWithoutProcessing('/updateMasteries/', data);
		}
	};

	// Return "has", "wants", or "" depending on the current status of the given mastery.
	this.getMasteryStatus = function(masteryId) {
		var has = false;
		var wants = false;
		for (var n = 0; n < this.masteryMapList.length; n++) {
			var masteryMap = this.masteryMapList[n];
			if (masteryMap && masteryId in masteryMap) {
				var mastery = masteryMap[masteryId];
				if (!mastery.has && !mastery.wants) {
					if (mastery.delHas) has = false;
					if (mastery.delWants) wants = false;
				} else if (mastery.wants) {
					wants = true;
				} else if (mastery.has) {
					has = true;
				}
			}
		}
		if (has) return 'has';
		if (wants) return 'wants';
		return '';
	};

	// Check if the user has the mastery
	this.hasMastery = function(masteryId) {
		return this.getMasteryStatus(masteryId) === 'has';
	};

	// Check if the user wants the mastery
	this.wantsMastery = function(masteryId) {
		return this.getMasteryStatus(masteryId) === 'wants';
	};

	// Check if the user doesn't have or want the mastery
	this.nullMastery = function(masteryId) {
		return this.getMasteryStatus(masteryId) === '';
	};

	// =========== Questionnaire helpers ====================
	// Map questionIndex -> {knows: [ids], wants: [ids], forgets: [ids]}
	this.setQuestionAnswer = function(qIndex, knows, wants, delKnows, delWants, updatePageObjectOptions) {
		if (qIndex <= 0) {
			return console.error('qIndex has to be > 0');
		}
		// Compute which masteries are affected
		var affectedMasteryIds = (qIndex in this.masteryMapList) ? Object.keys(this.masteryMapList[qIndex]) : [];
		// Compute new mastery map
		var masteryMap = {};
		for (var n = 0; n < delWants.length; n++) {
			var masteryId = delWants[n];
			masteryMap[masteryId] = {pageId: masteryId, has: false, wants: false, delWants: true};
			if (affectedMasteryIds.indexOf(masteryId) < 0) {
				affectedMasteryIds.push(masteryId);
			}
		}
		for (var n = 0; n < delKnows.length; n++) {
			var masteryId = delKnows[n];
			if (masteryId in masteryMap) {
				masteryMap[masteryId].delHas = true;
			} else {
				masteryMap[masteryId] = {pageId: masteryId, has: false, wants: false, delHas: true};
			}
			if (affectedMasteryIds.indexOf(masteryId) < 0) {
				affectedMasteryIds.push(masteryId);
			}
		}
		for (var n = 0; n < wants.length; n++) {
			var masteryId = wants[n];
			if (masteryId in masteryMap) {
				masteryMap[masteryId].wants = true;
			} else {
				masteryMap[masteryId] = {pageId: masteryId, has: false, wants: true};
			}
			if (affectedMasteryIds.indexOf(masteryId) < 0) {
				affectedMasteryIds.push(masteryId);
			}
		}
		for (var n = 0; n < knows.length; n++) {
			var masteryId = knows[n];
			if (masteryId in masteryMap) {
				masteryMap[masteryId].has = true;
			} else {
				masteryMap[masteryId] = {pageId: masteryId, has: true, wants: false};
			}
			if (affectedMasteryIds.indexOf(masteryId) < 0) {
				affectedMasteryIds.push(masteryId);
			}
		}
		this.masteryMapList[qIndex] = masteryMap;
		this.pushMasteriesToServer(affectedMasteryIds);
		this.updatePageObject(updatePageObjectOptions);
	};
});
