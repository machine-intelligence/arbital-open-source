"use strict";

// pages stores all the loaded pages and provides multiple helper functions for
// working with pages.
app.service("pageService", function(userService, $http){
	var that = this;

	// All loaded pages.
	this.pageMap = {};
	
	// All loaded edits. (These are the pages we will be editing.)
	this.editMap = {};

	// All loaded masteries.
	this.masteryMap = {};

	// Update whether on not the user has a mastery.
	this.updateMastery = function(scope, masteryId, has) {
		var mastery = that.masteryMap[masteryId];
		if (!mastery) {
			mastery = {pageId: masteryId};
			that.masteryMap[masteryId] = mastery;
		}
		mastery.has = has;
		mastery.isManuallySet = true;
		scope.$apply();

		// Send POST request.
		var data = {
			masteryId: masteryId,
			has: has,
		};
		$.ajax({
			type: "POST",
			url: "/updateMastery/",
			data: JSON.stringify(data),
		}).fail(function(r) {
			console.log("Failed to claim mastery:"); console.log(r);
		});
	};

	// Primary page is the one that's displayed front and center.
	this.primaryPage = undefined;
	// List of callbacks to notify when primary page changes.
	this.primaryPageCallbacks = [];
	// Set the primary page, triggering the callbacks.
	this.setPrimaryPage = function(newPrimaryPage) {
		var oldPrimaryPage = this.primaryPage;
		this.primaryPage = newPrimaryPage;
		for (var n = 0; n < this.primaryPageCallbacks.length; n++) {
			this.primaryPageCallbacks[n](oldPrimaryPage);
		}
		$("body").attr("last-visit", moment.utc(this.primaryPage.lastVisit).format("YYYY-MM-DD HH:mm:ss"));
	};
	
	// Call this to process data we received from the server.
	this.processServerData = function(data) {
		if (data.resetEverything) {
			this.pageMap = {};
			this.editMap = {};
			this.masteryMap = {};
		}
		$.extend(this.masteryMap, data["masteries"]);

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

	this.getPageUrl = function(pageId){
		return "/pages/" + pageId;
	};
	this.getEditPageUrl = function(pageId){
		return "/edit/" + pageId;
	};

	// These functions will be added to each page object.
	var pageFuncs = {
		// Check if the user has never visited this page before.
		isNewPage: function() {
			if (userService.user.id === "0") return false;
			return this.creatorId != userService.user.id &&
				(this.lastVisit === "" || this.originalCreatedAt >= this.lastVisit);
		},
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
			if (userService.user.id === "0") return false;
			return this.creatorId != userService.user.id &&
				this.lastVisit !== "" && this.createdAt >= this.lastVisit && this.lastVisit > this.originalCreatedAt;
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
			return this.type === "deleted";
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
		if (page.children == null) page.children = [];
		if (page.parents == null) page.parents = [];
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		// Add page's alias to the map as well
		if (pageMap && page.pageId !== page.alias) {
			pageMap[page.alias] = page;
		}
		return page;
	};
	// Add the given page to the global pageMap. If the page with the same id
	// already exists, we do a clever merge.
	var isPageValueTruthy = function(v) {
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
	this.addPageToMap = function(newPage) {
		var oldPage = this.pageMap[newPage.pageId];
		if (newPage === oldPage) return;
		if (oldPage === undefined) {
			this.pageMap[newPage.pageId] = setUpPage(newPage, this.pageMap);
			return;
		}
		// Merge each variable.
		for (var k in oldPage) {
			var oldV = isPageValueTruthy(oldPage[k]);
			var newV = isPageValueTruthy(newPage[k]);
			if (!newV) {
				// No new value.
				continue;
			}
			if (!oldV) {
				// No old value, so use the new one.
				oldPage[k] = newPage[k];
			}
			// Both new and old values are legit. Overwrite with new.
			oldPage[k] = newPage[k];
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

	// Load children for the given page. Success/error callbacks are called only
	// if request was actually made to the server.
	this.loadChildren = function(parent, success, error) {
		var that = this;
		if (parent.hasLoadedChildren) {
			success(parent.loadChildrenData, 200);
			return;
		} else if (parent.isLoadingChildren) {
			return;
		}
		parent.isLoadingChildren = true;
		console.log("Issuing POST request to /json/children/?parentId=" + parent.pageId);
		$http({method: "POST", url: "/json/children/", data: JSON.stringify({parentId: parent.pageId})}).
			success(function(data, status){
				parent.isLoadingChildren = false;
				parent.hasLoadedChildren = true;
				userService.processServerData(data);
				that.processServerData(data);
				parent.loadChildrenData = data["pages"];
				success(data["pages"], status);
			}).error(function(data, status){
				parent.isLoadingChildren = false;
				console.log("Error loading children:"); console.log(data); console.log(status);
				error(data, status);
			});
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
				console.log(page);
			}
			return function(aId, bId) {
				var diff = pageMap[bId].likeCount - pageMap[aId].likeCount;
				if (diff === 0) {
					return pageMap[aId].title.localeCompare(pageMap[bId].title);
				}
				return diff;
			};
		}
	};
	// Sort the given page's children.
	this.sortChildren = function(page) {
		var sortFunc = this.getChildSortFunc(page.sortChildrenBy);
		page.children.sort(function(aChild, bChild) {
			return sortFunc(aChild.childId, bChild.childId);
		});
	};

	// Load parents for the given page. Success/error callbacks are called only
	// if request was actually made to the server.
	this.loadParents = function(child, success, error) {
		var that = this;
		if (child.hasLoadedParents) {
			success(child.loadParentsData, 200);
			return;
		} else if (child.isLoadingParents) {
			return;
		}
		child.isLoadingParents = true;
		console.log("Issuing POST request to /json/parents/?childId=" + child.pageId);
		$http({method: "POST", url: "/json/parents/", data: JSON.stringify({childId: child.pageId})}).
			success(function(data, status){
				child.isLoadingParents = false;
				child.hasLoadedParents = true;
				userService.processServerData(data);
				that.processServerData(data);
				child.loadParentsData = data["pages"];
				success(data["pages"], status);
			}).error(function(data, status){
				child.isLoadingParents = false;
				console.log("Error loading parents:"); console.log(data); console.log(status);
				error(data, status);
			});
	};

	// Load the page with the given pageAlias.
	// options {
	//	 url: url to call
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
				console.log("JSON " + options.url + " data:"); console.dir(data);
				userService.processServerData(data);
				that.processServerData(data);
				var pageData = data["pages"];
				for (var id in pageData) {
					delete loadingPageAliases[options.url + id];
					delete loadingPageAliases[options.url + pageData[id].alias];
				}
				if(options.success) options.success();
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(options.error) options.error(data, status);
			}
		);
	};

	// Get data to display a popover for the page with the given alias.
	this.loadIntrasitePopover = function(pageAlias, options) {
		options.url = "/json/intrasitePopover/";
		loadPage(pageAlias, options);
	};

	// Get data to display a lens.
	this.loadLens = function(pageAlias, options) {
		options.url = "/json/lens/";
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
		$http({method: "POST", url: "/json/edit/", data: JSON.stringify(options)}).
			success(function(data, status){
				console.log("JSON /json/edit/ data:"); console.dir(data);
				if (!skipProcessDataStep) {
					userService.processServerData(data);
					that.processServerData(data);
				}
				if(success) success(data["edits"], status);
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

	// Get a new page from the server.
	// options {
	//	success: callback on success
	//}
	this.getNewPage = function(options) {
		$http({method: "POST", url: "/json/newPage/"}).
			success(function(data, status){
				console.log("JSON /json/newPage/ data:"); console.dir(data);
				userService.processServerData(data);
				that.processServerData(data);
				var pageId = Object.keys(data["pages"])[0];
				if(options.success) options.success(pageId);
			}).error(function(data, status){
				console.log("Error loading page:"); console.log(data); console.log(status);
				if(error) error(data, status);
			});
	}

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

	// Abandon the page with the given id.
	this.abandonPage = function(pageId, success, error) {
		var data = {
			pageId: pageId,
		};
		$http({method: "POST", url: "/abandonPage/", data: JSON.stringify(data)}).
			success(function(data, status){
				console.log("Successfully abandoned " + pageId);
				if(success) success(data, status);
			})
			.error(function(data, status){
				console.log("Error abandoning " + pageId + ":"); console.log(data); console.log(status);
				if(error) error(data, status);
			}
		);
	};

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
	this.showPublic = function(pageId) {
		/*if (this.privateGroupId !== undefined) {
			return this.privateGroupId !== this.pageMap[pageId].seeGroupId;
		}*/
		var page = this.pageMap[pageId];
		if (!this.primaryPage) return false;
		return this.primaryPage.seeGroupId !== page.seeGroupId && page.seeGroupId === "0";
	};
	// Return true iff we should show that this page belongs to a group.
	this.showLockedGroup = function(pageId) {
		var page = this.pageMap[pageId];
		if (!this.primaryPage) return page.seeGroupId !== "0";
		return this.primaryPage.seeGroupId !== page.seeGroupId && page.seeGroupId !== "0";
	};
});
