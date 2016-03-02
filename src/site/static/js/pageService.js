"use strict";

// pages stores all the loaded pages and provides multiple helper functions for
// working with pages.
app.service("pageService", function($http, $location, $ngSilentLocation, $rootScope, userService){
	var that = this;

	// Id of the private group we are in. (Corresponds to the subdomain).
	this.privateGroupId = "";

	// Primary page is the one with its id in the url
	this.primaryPage = undefined;

	// All loaded pages.
	this.pageMap = {};

	// All loaded edits. (These are the pages we will be editing.)
	this.editMap = {};

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

	// This object is set when the user is learning / on a path.
	this.path = undefined;

	// Update the masteryMap. Execution happens in the order options are listed.
	// options = {
	//		delete: set these masteries to "doesn't know"
	//		wants: set these masteries to "wants"
	//		knows: set these masteries to "knows"
	//		skipPush: if set, don't push the changes to the server
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
		if (!options.skipPush) {
			this.pushMasteriesToServer(affectedMasteryIds, options.callback);
		}
	};

	// Compute the status of the given masteries and update the server
	this.pushMasteriesToServer = function(affectedMasteryIds, callback) {
		var addMasteries = [], delMasteries = [], wantsMasteries = [];
		for (var n = 0; n < affectedMasteryIds.length; n++) {
			var masteryId = affectedMasteryIds[n];
			var masteryStatus = this.getMasteryStatus(masteryId);
			if (masteryStatus === "has") {
				addMasteries.push(masteryId);
			} else if (masteryStatus === "wants") {
				wantsMasteries.push(masteryId);
			} else {
				delMasteries.push(masteryId);
			}
		}

		if (userService.user.id !== "") {
			var data = {
				removeMasteries: delMasteries,
				wantsMasteries: wantsMasteries,
				addMasteries: addMasteries,
				computeUnlocked: !!callback, // hacky
			};
			$http({method: "POST", url: "/updateMasteries/", data: JSON.stringify(data)})
			.success(function(data) {
				if (callback) {
					userService.processServerData(data);
					that.processServerData(data);
					callback(data);
				}
			})
			.error(function(data, status){
				console.error("Failed to change masteries:"); console.log(data); console.log(status);
			});
		} else {
			this.updateMasteryMap({
				delete: delMasteries,
				wants: wantsMasteries,
				knows: addMasteries,
				skipPush: true,
			});
			Cookies.set("masteryMap", this.masteryMap, {expires: 365});
		}
	};

	// Compute the status of the given masteries and update the server
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

		if (userService.user.id !== "") {
			$http({method: "POST", url: "/updatePageObject/", data: JSON.stringify(options)})
			.error(function(data, status){
				console.error("Failed to update page object:"); console.log(data); console.log(status);
			});
		} else {
			Cookies.set("pageObjectMap", this.pageObjectMap, {expires: 365});
		}
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
			oldObject[k] = smartMerge(oldObject[k], newObject[k]);
		}
	};

	// Call this to process data we received from the server.
	this.processServerData = function(data) {
		if (data.resetEverything) {
			this.pageMap = {};
			this.editMap = {};
			this.masteryMap = {};
			this.masteryMapList = [this.masteryMap];
			this.pageObjectMap = {};
		}

		// Populate page object map.
		var pageObjectData = data["pageObjects"];
		for (var id in pageObjectData) {
			this.smartAddToMap(this.pageObjectMap, pageObjectData[id], id);
		}

		// Load page objects from cookie
		if (data.resetEverything && !userService.user.id) {
			var cookiePageObjectMap = Cookies.getJSON("pageObjectMap") || {};
			for (var id in cookiePageObjectMap) {
				this.smartAddToMap(this.pageObjectMap, cookiePageObjectMap[id], id);
			}
		} else if (data.resetEverything && userService.user.id) {
			Cookies.remove("pageObjectMap");
		}

		// Populate materies map.
		var masteryData = data["masteries"];
		for (var id in masteryData) {
			this.smartAddToMap(this.masteryMap, masteryData[id], id);
		}

		// Load masteries from cookie
		if (data.resetEverything && !userService.user.id) {
			var cookieMasteryMap = Cookies.getJSON("masteryMap") || {};
			for (var id in cookieMasteryMap) {
				this.smartAddToMap(this.masteryMap, cookieMasteryMap[id], id);
			}
		} else if (data.resetEverything && userService.user.id) {
			Cookies.remove("masteryMap");
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
	};


	// Construct a part of the URL with id and alias if id!=alias, otherwise just id
	var getBaseUrl = function(base, id, alias) {
		return "/" + base + "/" + id + (alias === id ? "" : "/" + alias) + "/";
	};

	// Returns the url for the given page.
	// options {
	//	 permalink: if true, we'll include page's id, otherwise, we'll use alias
	//	 includeHost: if true, include "http://" + host in the url
	// }
	// Track which pages we are already loading. Map url+pageAlias -> true.
	this.getPageUrl = function(pageId, options){
		var options = options || {};
		var host = window.location.host;
		var url = "/p/" + pageId + "/";
		var alreadyIncludedHost = false;
		var page = that.pageMap[pageId];

		if (page) {
			var pageId = page.pageId;
			var pageAlias = page.alias;
			url = getBaseUrl("p", options.permalink ? pageId : pageAlias, pageAlias);
			// Check page's type to see if we need a special url
			if (page.isLens()) {
				for (var n = 0; n < page.parentIds.length; n++) {
					var parent = this.pageMap[page.parentIds[n]];
					if (parent) {
						url = getBaseUrl("p", options.permalink ? parent.pageId : parent.alias, parent.alias) + "?l=" + pageId;
						if ($location.hash()) {
							url += "#" + $location.hash();
						}
						break;
					}
				}
			} else if (page.isComment() || page.isAnswer()) {
				for (var n = 0; n < page.parentIds.length; n++) {
					var parent = this.pageMap[page.parentIds[n]];
					if (parent && (
								(page.isComment() && (parent.isWiki() || parent.isLens())) ||
								(page.isAnswer() && parent.isQuestion()))) {
						url = getBaseUrl("p", options.permalink ? parent.pageId : parent.alias, parent.alias) + "#subpage-" + pageId;
						break;
					}
				}
			}
			// Check if we should set the domain
			if (page.seeGroupId != that.privateGroupId) {
				if (page.seeGroupId !== "") {
					url = that.getDomainUrl(that.pageMap[page.seeGroupId].alias) + url;
				} else {
					url = that.getDomainUrl() + url;
				}
				alreadyIncludedHost = true;
			}
		}
		if (options.includeHost && !alreadyIncludedHost) {
			url = "http://" + host + url;
		}
		return url;
	};

	this.getEditPageUrl = function(pageId){
		if (pageId in this.pageMap) {
			return getBaseUrl("edit", pageId, this.pageMap[pageId].alias);
		}
		return "/edit/" + pageId + "/";
	};

	// Return url to the user page.
	this.getUserUrl = function(userId) {
		if (userId in this.pageMap) {
			return getBaseUrl("user", userId, this.pageMap[userId].alias);
		}
		return "/user/" + userId;
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

	// Make sure the URL path is in the given canonical form, otherwise silently change
	// the URL, preserving the search() params.
	this.ensureCanonUrl = function(canonPath) {
		var pathname = location.pathname;
		if (pathname != canonPath) {
			var hash = $location.hash();
			var search = $location.search();
			$ngSilentLocation.silent(canonPath, true);
			$location.hash(hash);
			for (var k in search) {
				$location.search(k, search[k]);
			}
		}
	};

	// Return the corresponding page object, or undefined if not found.
	this.getPageObject = function(pageId, objectAlias) {
		var objectMap = this.pageObjectMap[pageId];
		if (!objectMap) return undefined;
		return objectMap[objectAlias];
	};

	// These functions will be added to each page object.
	var pageFuncs = {
		likeScore: function() {
			return this.likeCount + this.myLikeValue;
		},
		// Check if the user has never visited this page before.
		isNewPage: function() {
			if (!userService.user || userService.user.id === "") return false;
			return this.creatorId != userService.user.id &&
				(this.lastVisit === "" || this.originalCreatedAt >= this.lastVisit);
		},
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
			if (!userService.user || userService.user.id === "") return false;
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
		// Return just the title to display for a lens.
		lensTitle: function() {
			var parts = this.title.split(":");
			return parts[parts.length - 1].trim();
		},
	};

	// Massage page's variables to be easier to deal with.
	var setUpPage = function(page, pageMap) {
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		// Add page's alias to the map as well, both with lowercase and uppercase first letter
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
				if (options.success) options.success();
			}).error(function(data, status){
				if (!options.silentFail) {
					console.log("Error loading page:"); console.log(data); console.log(status);
				}
				if (options.error) options.error(data, status);
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
	//   success: callback on success
	//   error: callback on error
	// }
	var loadingUserPopovers = {};
	this.loadUserPopover = function(userId, options) {
		if (userId in loadingUserPopovers) {
			return;
		}
		loadingUserPopovers[userId] = true;
		options = options || {};
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;

		console.log("Issuing POST request to /json/userPopover/?userId=" + userId);
		$http({method: "POST", url: "/json/userPopover/", data: JSON.stringify({userId: userId})})
		.success(function(data, status){
			delete loadingUserPopovers[userId];
			userService.processServerData(data);
			that.processServerData(data);
			if (success) success(data, status);
		})
		.error(function(data, status){
			delete loadingUserPopovers[userId];
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
		return this.privateGroupId !== page.seeGroupId && page.seeGroupId === "";
	};
	// Return true iff we should show that this page belongs to a group.
	this.showPrivate = function(pageId, useEditMap) {
		var page = (useEditMap ? this.editMap : this.pageMap)[pageId];
		if (!page) {
			console.error("Couldn't find pageId: " + pageId);
			return false;
		}
		return this.privateGroupId !== page.seeGroupId && page.seeGroupId !== "";
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
		if (comment.isEditorComment) {
			userService.showEditorComments = true;
		}
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

		parent.subpageIds.push(commentId);
		$ngSilentLocation.silent(this.getPageUrl(commentId));
	};

	// Return "has", "wants", or "" depending on the current status of the given mastery.
	this.getMasteryStatus = function(masteryId) {
		var has = false, wants = false;
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
		if (has) return "has";
		if (wants) return "wants";
		return "";
	};

	// Check if the user has the mastery
	this.hasMastery = function(masteryId) {
		return this.getMasteryStatus(masteryId) === "has";
	};

	// Check if the user wants the mastery
	this.wantsMastery = function(masteryId) {
		return this.getMasteryStatus(masteryId) === "wants";
	};

	// Check if the user doesn't have or want the mastery
	this.nullMastery = function(masteryId) {
		return this.getMasteryStatus(masteryId) === "";
	};

	// =========== Questionnaire helpers ====================
	// Map questionIndex -> {knows: [ids], wants: [ids], forgets: [ids]}
	this.setQuestionAnswer = function(qIndex, knows, wants, delKnows, delWants, updatePageObjectOptions) {
		if (qIndex <= 0) {
			return console.error("qIndex has to be > 0");
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

	// Update the path variables.
	$rootScope.$watch(function() {
		return $location.absUrl() + "|" + (that.primaryPage ? that.primaryPage.pageId : "");
	}, function() {
		that.path = undefined;
		that.path = Cookies.getJSON("path");
		if (!that.path || !that.primaryPage) return;

		// Check if the user is learning
		var currentPageId = $location.search().l || that.primaryPage.pageId;
		var pathPageIds = that.path.readIds;
		var currentIndex = pathPageIds.indexOf(currentPageId);
		if (currentIndex >= 0) {
			that.path.onPath = true;
			that.path.prevPageId = currentIndex > 0 ? pathPageIds[currentIndex - 1] : "";
			that.path.nextPageId = currentIndex < pathPageIds.length - 1 ? pathPageIds[currentIndex + 1] : "";
			that.path.currentPageId = currentPageId;
		} else {
			that.path.onPath = false;
			that.path.prevPageId = that.path.nextPageId = "";
		}
		Cookies.set("path", that.path);
	});
});
