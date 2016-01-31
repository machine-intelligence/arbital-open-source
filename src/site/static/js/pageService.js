"use strict";

// pages stores all the loaded pages and provides multiple helper functions for
// working with pages.
app.service("pageService", function($http, $location, userService){
	var that = this;

	// All loaded pages.
	this.pageMap = {};
	
	// All loaded edits. (These are the pages we will be editing.)
	this.editMap = {};

	// All loaded masteries.
	this.masteryMap = {};
	// When the user answers questions or does other complex reversible actions, this map
	// allows us to store the new masteries the user acquired. That way we can allow the user
	// to change their answers, without messing up the masteries they learned through other means.
	// map "key" -> masteryMap
	this.tempMasteryMaps = {};

	// Update whether on not the user has certain masteries.
	// NOTE: removal will happen before adding
	this.updateMasteries = function(addMasteries, delMasteries, wantsMasteries, tempMap) {
		var masteryMap = tempMap || this.masteryMap;
		var affectedMasteryIds = [];
		for (var n = 0; n < delMasteries.length; n++) {
			var masteryId = delMasteries[n];
			var mastery = masteryMap[masteryId];
			if (!mastery) continue;
			mastery.has = false;
			mastery.wants = false;
			affectedMasteryIds.push(masteryId);
		}
		for (var n = 0; n < wantsMasteries.length; n++) {
			var masteryId = wantsMasteries[n];
			var mastery = masteryMap[masteryId];
			if (!mastery) {
				mastery = {pageId: masteryId};
				masteryMap[masteryId] = mastery;
			}
			mastery.has = false;
			mastery.wants = true;
			affectedMasteryIds.push(masteryId);
		}
		for (var n = 0; n < addMasteries.length; n++) {
			var masteryId = addMasteries[n];
			var mastery = masteryMap[masteryId];
			if (!mastery) {
				mastery = {pageId: masteryId};
				masteryMap[masteryId] = mastery;
			}
			mastery.has = true;
			mastery.wants = false;
			affectedMasteryIds.push(masteryId);
		}

		// Recompute mastery arrays to make sure the values reflect the curent, updated state
		addMasteries = [], delMasteries = [], wantsMasteries = [];
		for (var n = 0; n < affectedMasteryIds.length; n++) {
			var masteryId = affectedMasteryIds[n];
			if (this.hasMastery(masteryId)) {
				addMasteries.push(masteryId);
			} else if (this.wantsMastery(masteryId)) {
				wantsMasteries.push(masteryId);
			} else {
				delMasteries.push(masteryId);
			}
		}

		// Send POST request.
		var data = {
			removeMasteries: delMasteries,
			wantsMasteries: wantsMasteries,
			addMasteries: addMasteries,
		};
		$http({method: "POST", url: "/updateMasteries/", data: JSON.stringify(data)})
		.error(function(data, status){
			console.error("Failed to change masteries:"); console.log(data); console.log(status);
		});
	};

	this.addMasteryToMap = function(newMastery) {
		var oldMastery = this.masteryMap[newMastery.pageId];
		if (newMastery === oldMastery) return;
		if (oldMastery === undefined) {
			this.masteryMap[newMastery.pageId] = newMastery;
			return;
		}
		// Merge each variable.
		for (var k in oldMastery) {
			oldMastery[k] = smartMerge(oldMastery[k], newMastery[k]);
		}
	};

	// Id of the private group we are in. (Corresponds to the subdomain).
	this.privateGroupId = "0";

	// Primary page is the one with its id in the url
	this.primaryPage = undefined;
	
	// Call this to process data we received from the server.
	this.processServerData = function(data) {
		if (data.resetEverything) {
			this.pageMap = {};
			this.editMap = {};
			this.masteryMap = {};
			this.tempMasteryMaps = {};
		}

		var masteryData = data["masteries"];
		for (var id in masteryData) {
			this.addMasteryToMap(masteryData[id]);
		}

		var pageData = data["pages"];
		for (var id in pageData) {
			var page = pageData[id];
			if (page.isCurrentEdit) {
				this.addPageToMap(pageData[id]);
			} else {
				this.addPageToEditMap(pageData[id]);
			}
		}

		var editData = data["edits"];
		for (var id in editData) {
			this.addPageToEditMap(editData[id]);
		}
	}

	// Returns the url for the given page.
	this.getPageUrl = function(pageId){
		var page = that.pageMap[pageId];
		var url = "/pages/" + pageId;
		if (page) {
			// Check page's type to see if we need a special url
			if (page.isLens()) {
				for (var n = 0; n < page.parentIds.length; n++) {
					var parent = this.pageMap[page.parentIds[n]];
					if (parent) {
						url = "/pages/" + parent.pageId + "?lens=" + pageId + "#" + $location.hash();
						break;
					}
				}
			} else if (page.isComment() || page.isAnswer()) {
				for (var n = 0; n < page.parentIds.length; n++) {
					var parent = this.pageMap[page.parentIds[n]];
					if (parent && (
								(page.isComment() && (parent.isWiki() || parent.isLens())) ||
								(page.isAnswer() && parent.isQuestion()))) {
						url = "/pages/" + parent.pageId + "#subpage-" + pageId;
						break;
					}
				}
			}
			// Check if we should set the domain
			if (page.seeGroupId != that.privateGroupId) {
				if (page.seeGroupId !== "0") {
					url = that.getDomainUrl(that.pageMap[page.seeGroupId].alias) + url;
				} else {
					url = that.getDomainUrl() + url;
				}
			}
		}
		return url;
	};
	this.getEditPageUrl = function(pageId){
		return "/edit/" + pageId;
	};

	// Get a domain url (with optional subdomain)
	this.getDomainUrl = function(subdomain) {
		if (subdomain) {
			subdomain += ".";
		} else {
			subdomain = "";
		}
		if (/localhost/.exec($location.host())) {
			return "http://" + subdomain + "localhost:8012";
		} else {
			return "http://" + subdomain + "arbital.com"
		}
	};

	// These functions will be added to each page object.
	var pageFuncs = {
		likeScore: function() {
			return this.likeCount + this.myLikeValue;
		},
		// Check if the user has never visited this page before.
		isNewPage: function() {
			if (!userService.user || userService.user.id === "0") return false;
			return this.creatorId != userService.user.id &&
				(this.lastVisit === "" || this.originalCreatedAt >= this.lastVisit);
		},
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
			if (!userService.user || userService.user.id === "0") return false;
			return this.creatorId != userService.user.id &&
				this.lastVisit !== "" && this.createdAt >= this.lastVisit && this.lastVisit > this.originalCreatedAt;
		},
		isWiki: function() {
			return this.type === "wiki";
		},
		isLens: function() {
			return this.type === "lens";
		},
		isQuestion: function() {
			return this.type === "question";
		},
		isAnswer: function() {
			return this.type === "answer";
		},
		isComment: function() {
			return this.type === "comment";
		},
		isGroup: function() {
			return this.type === "group";
		},
		isDomain: function() {
			return this.type === "domain";
		},
		// Return empty string if the user can edit this page. Otherwise a reason for
		// why they can't.
		getEditLevel: function() {
			var karmaReq = 200; // TODO: fix this
			if (userService.user.karma < karmaReq) {
				if (userService.user.isAdmin) {
					// Can edit but only because user is an admin.
					return "admin";
				}
				return "" + karmaReq;
			}
			return "";
		},
		// Return empty string if the user can delete this page. Otherwise a reason
		// for why they can't.
		getDeleteLevel: function() {
			var karmaReq = 200; // TODO: fix this
			if (userService.user.karma < karmaReq) {
				if (userService.user.isAdmin) {
					return "admin";
				}
				return "" + karmaReq;
			}
			return "";
		},
		// Return true iff the page is deleted.
		isDeleted: function() {
			return this.type === "";
		},
		// Get page's url
		url: function() {
			return that.getPageUrl(this.pageId);
		},
		// Get url to edit the page
		editUrl: function() {
			return that.getEditPageUrl(this.pageId);
		},
	};
	
	// Massage page's variables to be easier to deal with.
	var setUpPage = function(page, pageMap) {
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		// Add page's alias to the map as well, both lowercase and uppercase
		if (pageMap && page.pageId !== page.alias) {
			pageMap[page.alias.substring(0,1).toLowerCase() + page.alias.substring(1)] = page;
			pageMap[page.alias.substring(0,1).toUpperCase() + page.alias.substring(1)] = page;
		}

		return page;
	};
	// Add the given page to the global pageMap. If the page with the same id
	// already exists, we do a clever merge.
	var isValueTruthy = function(v) {
		// "0" is falsy
		if (v === "0") {
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
	var smartMerge = function(oldV, newV) {
		if (!isValueTruthy(newV)) {
			return oldV;
		}
		return newV;
	};
	this.addPageToMap = function(newPage) {
		var oldPage = this.pageMap[newPage.pageId];
		if (newPage === oldPage) return;
		if (oldPage === undefined) {
			this.pageMap[newPage.pageId] = setUpPage(newPage, this.pageMap);
			return;
		}
		// Merge each variable.
		for (var k in oldPage) {
			oldPage[k] = smartMerge(oldPage[k], newPage[k]);
		}
	};

	// Remove page with the given pageId from the global pageMap.
	this.removePageFromMap = function(pageId) {
		delete this.pageMap[pageId];
	};

	// Add the given page to the global editMap.
	this.addPageToEditMap = function(page) {
		this.editMap[page.pageId] = setUpPage(page);
	}

	// Remove page with the given pageId from the global editMap;
	this.removePageFromEditMap = function(pageId) {
		delete this.editMap[pageId];
	};

	// Return function for sorting children ids.
	this.getChildSortFunc = function(sortChildrenBy) {
		var pageMap = this.pageMap;
		if(sortChildrenBy === "alphabetical") {
			return function(aId, bId) {
				var aTitle = pageMap[aId].title;
				var bTitle = pageMap[bId].title;
				// If title starts with a number, we want to compare those numbers directly,
				// otherwise "2" comes after "10".
				var aNum = parseInt(aTitle);
				if (aNum) {
					var bNum = parseInt(bTitle);
					if (bNum) {
						return aNum - bNum;
					}
				}
				return pageMap[aId].title.localeCompare(pageMap[bId].title);
			};
		} else if (sortChildrenBy === "recentFirst") {
			return function(aId, bId) {
				return pageMap[bId].originalCreatedAt.localeCompare(pageMap[aId].originalCreatedAt);
			};
		} else if (sortChildrenBy === "oldestFirst") {
			return function(aId, bId) {
				return pageMap[aId].originalCreatedAt.localeCompare(pageMap[bId].originalCreatedAt);
			};
		} else {
			if (sortChildrenBy !== "likes") {
				console.error("Unknown sort type: " + sortChildrenBy);
			}
			return function(aId, bId) {
				var diff = pageMap[bId].likeScore() - pageMap[aId].likeScore();
				if (diff === 0) {
					return pageMap[bId].createdAt > pageMap[aId].createdAt;
				}
				return diff;
			};
		}
	};
	// Sort the given page's children.
	this.sortChildren = function(page) {
		var sortFunc = this.getChildSortFunc(page.sortChildrenBy);
		page.childIds.sort(function(aChildId, bChildId) {
			return sortFunc(aChildId, bChildId);
		});
	};

	// Load the page with the given pageAlias.
	// options {
	//	 url: url to call
	//	 silentFail: don't print error if failed
	//   success: callback on success
	//   error: callback on error
	// }
	// Track which pages we are already loading. Map url+pageAlias -> true.
	var loadingPageAliases = {};
	var loadPage = function(pageAlias, options) {
		// Check if the page is already being loaded, and mark it as such if it's not.
		var loadKey = options.url + pageAlias;
		if (loadKey in loadingPageAliases) {
			return;
		}
		loadingPageAliases[loadKey] = true;

		console.log("Issuing a POST request to: " + options.url + "?pageAlias=" + pageAlias);
		$http({method: "POST", url: options.url, data: JSON.stringify({pageAlias: pageAlias})}).
			success(function(data, status){
				if (!options.silentFail) {
					console.log("JSON " + options.url + " data:"); console.dir(data);
				}
				userService.processServerData(data);
				that.processServerData(data);
				var pageData = data["pages"];
				for (var id in pageData) {
					delete loadingPageAliases[options.url + id];
					delete loadingPageAliases[options.url + pageData[id].alias];
				}
				if(options.success) options.success();
			}).error(function(data, status){
				if (!options.silentFail) {
					console.log("Error loading page:"); console.log(data); console.log(status);
				}
				if(options.error) options.error(data, status);
			}
		);
	};

	// Get data to display a popover for the page with the given alias.
	this.loadIntrasitePopover = function(pageAlias, options) {
		options = options || {};
		options.url = "/json/intrasitePopover/";
		loadPage(pageAlias, options);
	};

	// Get data to display a popover for the user with the given alias.
	// options {
	//	 url: url to call
	//   success: callback on success
	//   error: callback on error
	// }
	this.loadUserPopover = function(userId, options) {
		options = options || {};
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;

		console.log("Issuing POST request to /json/userPopover/?userId=" + userId);
		$http({method: "POST", url: "/json/userPopover/", data: JSON.stringify({userId: userId})})
		.success(function(data, status){
			userService.processServerData(data);
			that.processServerData(data);
			if (success) success(data, status);
		})
		.error(function(data, status){
			console.error("Error loading user popover:"); console.log(data); console.log(status);
			if (error) error(data, status);
		});
	};

	// Get data to display a lens.
	this.loadLens = function(pageAlias, options) {
		options = options || {};
		options.url = "/json/lens/";
		loadPage(pageAlias, options);
	};

	// Get data to display page's title
	this.loadTitle = function(pageAlias, options) {
		options = options || {};
		options.url = "/json/title/";
		loadPage(pageAlias, options);
	};
	
	// Load edit.
	// options {
	//   pageAlias: pageAlias to load
	//   specificEdit: load page with this edit number
	//	 editLimit: only load edits lower than this number
	//	 createdAtLimit: only load edits that were created before this date
	//	 skipProcessDataStep: if true, we don't process the data we get from the server
	//   success: callback on success
	//   error: callback on error
	// }
	this.loadEdit = function(options) {
		// Set up options.
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;
		var skipProcessDataStep = options.skipProcessDataStep; delete options.skipProcessDataStep;

		console.log("Issuing a POST request to: /json/edit/?pageAlias=" + options.pageAlias);
		$http({method: "POST", url: "/json/edit/", data: JSON.stringify(options)})
		.success(function(data, status){
			console.log("JSON /json/edit/ data:"); console.dir(data);
			if (!skipProcessDataStep) {
				userService.processServerData(data);
				that.processServerData(data);
			}
			if(success) success(data["edits"], status);
		})
		.error(function(data, status){
			console.log("Error loading page:"); console.log(data); console.log(status);
			if(error) error(data, status);
		});
	};

	// Get a new page from the server.
	// options {
	//  type: type of the page to create
	//  parentIds: optional array of parents to add to the new page
	//	success: callback on success
	//	error: callback on error
	//}
	this.getNewPage = function(options) {
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;

		$http({method: "POST", url: "/json/newPage/", data: JSON.stringify(options)})
		.success(function(data, status){
			console.log("JSON /json/newPage/ data:"); console.dir(data);
			userService.processServerData(data);
			that.processServerData(data);
			var pageId = Object.keys(data["edits"])[0];
			if(success) success(pageId);
		})
		.error(function(data, status){
			console.log("Error getting a new page:"); console.log(data); console.log(status);
			if(error) error(data, status);
		});
	};

	// Delete the page with the given pageId.
	this.deletePage = function(pageId, success, error) {
		var data = {
			pageId: pageId,
		};
		$http({method: "POST", url: "/deletePage/", data: JSON.stringify(data)})
		.success(function(data, status){
			console.log("Successfully deleted " + pageId);
			if(success) success(data, status);
		})
		.error(function(data, status){
			console.log("Error deleting " + pageId + ":"); console.log(data); console.log(status);
			if(error) error(data, status);
		}
		);
	};

	// Discard the page with the given id.
	this.discardPage = function(pageId, success, error) {
		var data = {
			pageId: pageId,
		};
		$http({method: "POST", url: "/discardPage/", data: JSON.stringify(data)})
		.success(function(data, status){
			console.log("Successfully discarded " + pageId);
			if(success) success(data, status);
		})
		.error(function(data, status){
			console.log("Error discarding " + pageId + ":"); console.log(data); console.log(status);
			if(error) error(data, status);
		}
		);
	};

	// (Un)subscribe a user to a page.
	this.subscribeTo = function($target) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			pageId: $target.attr("page-id"),
		};
		var isSubscribed = $target.hasClass("on");
		$.ajax({
			type: "POST",
			url: isSubscribed ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		});
		this.pageMap[data.pageId].isSubscribed = isSubscribed;
		$rootScope.$apply();
	}

	// Add a new relationship between pages using the given options.
	// options = {
	//	parentId: id of the parent page
	//	childId: id of the child page
	//	type: type of the relationships
	// }
	this.newPagePair = function(options, success) {
		$http({method: "POST", url: "/newPagePair/", data: JSON.stringify(options)})
		.success(function(data, status){
			if(success) success();
		})
		.error(function(data, status){
			console.log("Error creating new page pair:"); console.log(data); console.log(status);
		});
	};
	// Note: you also need to specify the type of the relationship here, sinc we
	// don't want to accidentally delete the wrong type.
	this.deletePagePair = function(options) {
		$http({method: "POST", url: "/deletePagePair/", data: JSON.stringify(options)})
		.error(function(data, status){
			console.log("Error deleting a page pair:"); console.log(data); console.log(status);
		});
	};

	// TODO: make these into page functions?
	// Return true iff we should show that this page is public.
	this.showPublic = function(pageId, useEditMap) {
		var page = (useEditMap ? this.editMap : this.pageMap)[pageId];
		if (!page) {
			console.error("Couldn't find pageId: " + pageId);
			return false;
		}
		return this.privateGroupId !== page.seeGroupId && page.seeGroupId === "0";
	};
	// Return true iff we should show that this page belongs to a group.
	this.showPrivate = function(pageId, useEditMap) {
		var page = (useEditMap ? this.editMap : this.pageMap)[pageId];
		if (!page) {
			console.error("Couldn't find pageId: " + pageId);
			return false;
		}
		return this.privateGroupId !== page.seeGroupId && page.seeGroupId !== "0";
	};

	// Create a new comment; optionally it's a reply to the given commentId
	// options: {
	//  parentPageId: id of the parent page
	//	replyToId: (optional) comment id this will be a reply to
	//	success: callback
	// }
	this.newComment = function(options) {
		var parentIds = [options.parentPageId];
		if (options.replyToId) {
			parentIds.push(options.replyToId);
		}
		// Create new comment
		this.getNewPage({
			type: "comment",
			parentIds: parentIds,
			success: function(newCommentId) {
				if (options.success) {
					options.success(newCommentId);
				}
			},
		});
	};

	// Called when the user created a new comment.
	this.newCommentCreated = function(commentId) {
		var comment = this.editMap[commentId];
		comment.originalCreatedAt = moment().format("YYYY-MM-DD HH:mm:ss");
		this.addPageToMap(comment);

		// Find the parent comment, or fall back on the parent page
		var parent = undefined;
		for (var n = 0; n < comment.parentIds.length; n++) {
			var p = this.pageMap[comment.parentIds[n]];
			if (!parent || p.isComment()) {
				parent = p;
			}
		}

		parent.commentIds.push(commentId);
		parent.subpageIds.push(commentId);
		$location.hash("subpage-" + commentId);
	};

	// Check if the user has a mastery
	this.hasMastery = function(masteryId) {
		// The user "has" the mastery if ANY of the temp maps says so
		for (var key in this.tempMasteryMaps) {
			var masteryMap = this.tempMasteryMaps[key];
			if (masteryId in masteryMap && masteryMap[masteryId].has) {
				return true;
			}
		}
		return (masteryId in this.masteryMap) && this.masteryMap[masteryId].has;
	};

	// Check if the user wants the given mastery
	this.wantsMastery = function(masteryId) {
		// The user "wants" the mastery if they don't know it already, and ANY of the
		// temp maps says so
		var has = false, wants = false;
		for (var key in this.tempMasteryMaps) {
			var masteryMap = this.tempMasteryMaps[key];
			if (masteryId in masteryMap) {
				has = has || masteryMap[masteryId].has;
				wants = wants || masteryMap[masteryId].wants;
			}
		}
		if (!has && wants) return true;
		return (masteryId in this.masteryMap) && this.masteryMap[masteryId].wants;
	};

	// =========== Questionnaire helpers ====================
	// Map questionIndex -> {answer: 'a', knows: [ids], wants: [ids]}
	this.answers = {};
	this.setQuestionAnswer = function(qIndex, answer, aKnows, aWants) {
		// Compute which masteries to remove
		var removeMasteries = [];
		if (qIndex in this.answers) {
			var answer = this.answers[qIndex];
			removeMasteries = answer.knows.concat(answer.wants);
		}
		this.answers[qIndex] = {answer: answer, knows: aKnows, wants: aWants};

		if (!(qIndex in this.tempMasteryMaps)) {
			this.tempMasteryMaps[qIndex] = {};
		}
		this.updateMasteries(aKnows, removeMasteries, aWants, this.tempMasteryMaps[qIndex]);
	};
});
