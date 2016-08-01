'use strict';

import app from './angular.ts';
import {arraysSortFn} from './util.ts';

// Manages masteries
app.service('masteryService', function($http, $compile, $location, $mdToast, $rootScope, $interval, stateService, pageService) {
	var that = this;

	// All loaded masteries.
	this.masteryMap = {};
	// When the user answers questions or does other complex reversible actions, this map
	// allows us to store the new masteries the user acquired/lost. That way we can allow the user
	// to change their answers, without messing up the masteries they learned through other means.
	// sorted array: [map "key" -> masteryMap]
	// Array is sorted by the order in which the questions appear in the text.
	this.masteryMapList = [this.masteryMap];

	// All page objects currently loaded
	// pageId -> {object -> {object data}}
	this.pageObjectMap = {};

	// This is called when we get data from the server.
	var postDataCallback = function(data) {
		if (data.resetEverything) {
			that.masteryMap = {};
			that.masteryMapList = [that.masteryMap];
			that.pageObjectMap = {};
		}

		// Populate materies map.
		var masteryData = data.masteries;
		for (var id in masteryData) {
			stateService.smartAddToMap(that.masteryMap, masteryData[id], id);
		}

		// Populate page object map.
		var pageObjectData = data.pageObjects;
		for (var id in pageObjectData) {
			stateService.smartAddToMap(that.pageObjectMap, pageObjectData[id], id);
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
				var masteryId = options.delete[n].parentId;
				var mastery = this.masteryMap[masteryId];
				if (!mastery) continue;
				mastery.has = false;
				mastery.wants = false;
				affectedMasteryIds.push(masteryId);
			}
		}
		if (options.wants) {
			for (var n = 0; n < options.wants.length; n++) {
				var masteryId = options.wants[n].parentId;
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
				var masteryId = options.knows[n].parentId;
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
		if (affectedMasteryIds.length <= 0) {
			return;
		}
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
			taughtBy: pageService.getCurrentPageId(),
		};
		if (callback) {
			stateService.postData('/updateMasteriesOld/', data, callback);
		} else {
			stateService.postDataWithoutProcessing('/updateMasteriesOld/', data);
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

	// ===== Hub content helpers ======

	// Returns true if the current user can learn something new by reading this page
	this.doesPageTeachUnknownReqs = function(pageId) {
		return stateService.pageMap[pageId].subjects.some(function(subject) {
			if (!subject.isStrong) return false;
			var subjectId = subject.parentId;
			if (!(subjectId in that.masteryMap)) return false;
			var currentUserLevel = 0;
			if (that.masteryMap[subjectId].has) {
				currentUserLevel = that.masteryMap[subjectId].level;
			}
			return subject.level > currentUserLevel;
		});
	};

	// Sort pageIds in the given page's hubContent
	this.sortHubContent = function(page) {
		for (var n = 0; n < page.hubContent.boostPageIds.length; n++) {
			page.hubContent.boostPageIds[n].sort(arraysSortFn(function(pageId) {
				var page = stateService.pageMap[pageId];
				return [
					-page.pathPages.length,
					that.doesPageTeachUnknownReqs(pageId) ? 1 : 0,
				];
			}));
			page.hubContent.learnPageIds[n].sort(arraysSortFn(function(pageId) {
				var page = stateService.pageMap[pageId];
				return [
					-page.pathPages.length,
					that.doesPageTeachUnknownReqs(pageId) ? 1 : 0,
				];
			}));
		}
	};

	// Return true if there is at least one boost page where the user will learn something new
	this.hasUnreadBoostPages = function(page, level, ignorePageId) {
		return page.hubContent.boostPageIds[level].some(function(pageId) {
			if (pageId == ignorePageId) return false;
			if (that.doesPageTeachUnknownReqs(pageId)) {
				return true;
			}
		});
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

	// Propate object state to the server
	// options = {
	//	pageId: id of the page
	//	edit: current edit of the page
	//	object: page object's alias
	//	value: page object's value
	// }
	this.updatePageObject = function(options) {
		if (!(options.pageId in this.pageObjectMap)) {
			this.pageObjectMap[options.pageId] = {};
		}
		this.pageObjectMap[options.pageId][options.object] = options;

		$http({method: 'POST', url: '/updatePageObject/', data: JSON.stringify(options)})
		.error(function(data, status) {
			console.error('Failed to update page object:'); console.log(data); console.log(status);
		});
	};

	// Return the corresponding page object, or undefined if not found.
	this.getPageObject = function(pageId, objectAlias) {
		var objectMap = this.pageObjectMap[pageId];
		if (!objectMap) return undefined;
		return objectMap[objectAlias];
	};
});
