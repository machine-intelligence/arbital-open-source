/*
 *	Mastery object: {
 *		pageId: id of the mastery,
 *		has: true if the user has this mastery
 *	}
 * */


// ================================ CODE ===================================
function masteryService() {
	// Primary mastery map. Maps masteryId -> mastery object
	this.masteryMap = {};
	// Track all mastery maps
	this.masteryMapList = [this.masteryMap];

	// Return "has", "wants", or "" depending on the current status of the given mastery.
	this.getMasteryStatus = function(masteryId) {
		var has = false;
		var wants = false;
		for (var n = 0; n < this.masteryMapList; n++) {
			var masteryMap = this.masteryMapList[n];
			if (masteryMap && masteryId in masteryMap) {
				var mastery = masteryMap[masteryId]
				if (mastery.has) {
					return 'has';
				}
			}
		}
		return '';
	};

	this.setQuestionAnswer = function(qIndex, knows) {
		var affectedMasteryIds = (qIndex in this.masteryMapList) ? Object.keys(this.masteryMapList[qIndex]) : [];
		var masteryMap = {};
		for (var n = 0; n < knows.length; n++) {
			var masteryId = knows[n];
			masteryMap[masteryId] = {pageId: masteryId, has: true};
			if (affectedMasteryIds.indexOf(masteryId) < 0) {
				affectedMasteryIds.push(masteryId);
			}
		}
		this.masteryMapList[qIndex] = masteryMap;
		this.pushMasteriesToServer(affectedMasteryIds)
	};



	// Compute the status of the given masteries and update the server
	var pushMasteriesToServer = function(affectedMasteryIds, callback) {
		if (affectedMasteryIds.length <= 0) {
			return;
		}
		var addMasteries = [];
		var removedMasteries = [];
		for (var n = 0; n < affectedMasteryIds.length; n++)
		{
			var masteryId = affectedMasteryIds[n];
			var masteryStatus = this.getMasteryStatus(masteryId);
			if (masteryStatus === 'has') {
				addMasteries.push(masteryId);
			}
		}

		var data = {
			addMasteries: addMasteries,
			removedMasteries: removedMasteries,
			// Note: this is a bit hacky. We should probably pass computeUnlocked explicitly
			computeUnlocked: !!callback,
		};
		if (callback) {
			console.log('/updateMasteries/ with callback');
			//stateService.postData('/updateMasteries/', data, callback);
		} else
		{
			console.log('/updateMasteries/ without callback');
			//stateService.postDataWithoutProcessing('/updateMasteries/', data);
		}
	};
	this.pushMasteriesToServer = pushMasteriesToServer;
};
