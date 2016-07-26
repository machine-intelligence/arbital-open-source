/*
 * Mastery object: {
 *	pageId: id of the mastery,
 *	has: true if the user has this mastery
 * }
 * */


// ================================ CODE ===================================
function masteryService() {
	// Primary mastery map. Maps masteryId -> mastery object
	this.masteryMap = {};
	// Track all mastery maps
	this.masteryMapList = [this.masteryMap];

	// Return "has", "wants", or "" depending on the current status of the given mastery.
	this.getMasteryStatus = function (masteryId) {
		var foundStatus = null;
		for (var n = 0; n < this.masteryMapList.length; n++) {
			var masteryMap = this.masteryMapList[n];
			var status = masteryMap[masteryId];
			if (status != null) {
				if (status.has) {
					foundStatus = 'has';
				} else {
					foundStatus = 'wants';
				}
			}
		}
		if (foundStatus == null) {
			return ''
		} else {
			return foundStatus;
		}
	};

	// When the question specified by qIndex is checked as 'known',
	// then all masteries in the list knows will be marked as known.
	// Can add additional masteries to a question, but can't remove them
	this.setQuestionAnswer = function(qIndex, knows) { //knows is list of masteryIDs
		// Get existing map if possible, otherwise start a new one
		var affectedMasteryIds = (qIndex in this.masteryMapList) ? Object.keys(this.masteryMapList[qIndex]) : [];
		var masteryMap = {};
		for (var n = 0; n < knows.length; n++) {
			var masteryId = knows[n];
			masteryMap[masteryId] = {pageId: masteryId, has: true};
			// Only update new known requirements
			if (affectedMasteryIds.indexOf(masteryId) < 0) {
				affectedMasteryIds.push(masteryId);
			}
		}
		// Update model
		this.masteryMapList[qIndex] = masteryMap;
		// Sync model to server
		this.pushMasteriesToServer(affectedMasteryIds)
	};



	// Compute the status of the given masteries and update the server
	// TODO removed masteries?
	// TODO callback not used in set questions? When needed
	this.pushMasteriesToServer = function(affectedMasteryIds, callback) {
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

		// update server
		var data = {
			addMasteries: addMasteries,
			removedMasteries: removedMasteries,
			// Note: this is a bit hacky. We should probably pass computeUnlocked explicitly
			computeUnlocked: !!callback
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
}
