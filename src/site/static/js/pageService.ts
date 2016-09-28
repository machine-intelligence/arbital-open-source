'use strict';

import app from './angular.ts';

import {
	complexLinkRegexp,
	forwardLinkRegexp,
	simpleLinkRegexp,
	atAliasRegexp,
	notEscaped,
	noParen,
} from './markdownService.ts';

// pages stores all the loaded pages and provides multiple helper functions for
// working with pages.
app.service('pageService', function($http, $compile, $location, $rootScope, $interval, analyticsService, stateService, userService, urlService) {
	var that = this;

	// Call this to process data we received from the server.
	var postDataCallback = function(data) {
		if (data.resetEverything) {
			stateService.pageMap = {};
			stateService.deletedPagesMap = {};
			stateService.editMap = {};
		}

		var pageData = data.pages;
		for (var id in pageData) {
			var page = pageData[id];
			if (page.isDeleted) {
				that.addPageToDeletedPagesMap(pageData[id]);
			} else if (page.isLiveEdit) {
				that.addPageToMap(pageData[id]);
			} else {
				that.addPageToEditMap(pageData[id]);
			}
		}

		var editData = data.edits;
		for (var id in editData) {
			that.addPageToEditMap(editData[id]);
		}
	};
	stateService.addPostDataCallback('pageService', postDataCallback);

	// Returns the id of the current page, if there is one.
	this.getCurrentPageId = function() {
		return $location.search().l ||
			(stateService.primaryPage ? stateService.primaryPage.pageId : '');
	};

	// Returns the current page
	this.getCurrentPage = function() {
		return stateService.pageMap[that.getCurrentPageId()];
	};

	// These functions will be added to each page object.
	var pageFuncs = {
		likeScore: function() {
			return this.likeCount + this.myLikeValue;
		},
		// Check if the user has never visited this page before.
		isNewPage: function() {
			if (!userService.user || userService.user.id === '') return false;
			return this.pageCreatorId != userService.user.id &&
				(this.lastVisit === '' || this.pageCreatedAt >= this.lastVisit);
		},
		// Check if the page has been updated since the last time the user saw it.
		isUpdatedPage: function() {
			if (!userService.user || userService.user.id === '') return false;
			// TODO: this is actually hard to compute correctly, because we don't want to show the change
			// if the user updated the page themselves, but there could have been a prior change
			// by someone else that they haven't seen.
			return this.editCreatorId != userService.user.id &&
				this.lastVisit !== '' && this.editCreatedAt >= this.lastVisit && this.lastVisit > this.pageCreatedAt;
		},
		isWiki: function() {
			return this.type === 'wiki';
		},
		isQuestion: function() {
			return this.type === 'question';
		},
		isComment: function() {
			return this.type === 'comment';
		},
		isGroup: function() {
			return this.type === 'group';
		},
		isDomain: function() {
			return this.type === 'domain';
		},
		isUser: function() {
			return this.pageId in userService.userMap;
		},
		// Return number of domainSubmissions values
		domainSubmissionsCount: function() {
			return Object.keys(this.domainSubmissions).length;
		},
		getCommentParentPage: function() {
			console.assert(this.isComment(), 'Calling getCommentParentPage on a non-comment');
			for (var n = 0; n < this.parentIds.length; n++) {
				var p = stateService.pageMap[this.parentIds[n]];
				if (!p.isComment()) {
					return p;
				}
			}
			return null;
		},
		// Return the top level comment for the thread this comment is in.
		getTopLevelComment: function() {
			console.assert(this.isComment(), 'Calling getTopLevelComment on a non-comment');
			for (var n = 0; n < this.parentIds.length; n++) {
				var p = stateService.pageMap[this.parentIds[n]];
				if (p.isComment()) {
					return p;
				}
			}
			return this;
		},
		// Get page's url
		url: function() {
			return urlService.getPageUrl(this.pageId);
		},
		// Get url to edit the page
		editUrl: function() {
			return urlService.getEditPageUrl(this.pageId);
		},
		// Return "pageInfo" component of the page.
		getPageInfo: function() {
			return {
				pageId: this.pageId,
				alias: this.alias,
				type: this.type,
				seeGroupId: this.seeGroupId,
				editGroupId: this.editGroupId,
				hasVote: this.hasVote,
				voteType: this.voteType,
				sortChildrenBy: this.sortChildrenBy,
				isRequisite: this.isRequisite,
				indirectTeacher: this.indirectTeacher,
				isEditorCommentIntention: this.isEditorCommentIntention,
			};
		},
		// Return true iff the page is in the given domain
		isInDomain: function(domainId) {
			return this.domainIds.indexOf(domainId) >= 0;
		},
		// Helper function for getBest...Id functions
		_getBestPageId: function(pageIds, excludePageId) {
			var pageId = undefined;
			for (var n = 0; n < pageIds.length; n++) {
				if (pageIds[n] != excludePageId) {
					pageId = pageIds[n];
					break;
				}
			}
			return pageId;
		},
		// Return pageId most suited for this user to boost understanding of this page
		getBestBoostPageId: function(currentLevel, excludePageId = undefined) {
			if (currentLevel <= 0) return undefined;
			var pageIds = this.hubContent.boostPageIds[currentLevel];
			return this._getBestPageId(pageIds, excludePageId);
		},
		// Return pageId most suited for this user to learn the next level of this page
		getBestLearnPageId: function(currentLevel, excludePageId) {
			if (currentLevel >= this.hubContent.learnPageIds.length - 1) return undefined;
			var pageIds = this.hubContent.learnPageIds[currentLevel + 1];
			return this._getBestPageId(pageIds, excludePageId);
		},
	};

	// Massage page's variables to be easier to deal with.
	var setUpPage = function(page) {
		for (var name in pageFuncs) {
			page[name] = pageFuncs[name];
		}
		return page;
	};
	this.addPageToMap = function(newPage) {
		var oldPage = stateService.pageMap[newPage.pageId];
		if (newPage === oldPage) return oldPage;
		if (oldPage === undefined) {
			stateService.pageMap[newPage.pageId] = setUpPage(newPage);
			// Add page's alias to the map as well, both with lowercase and uppercase first letter
			if (newPage.pageId !== newPage.alias) {
				stateService.pageMap[newPage.alias.substring(0, 1).toLowerCase() + newPage.alias.substring(1)] = newPage;
				stateService.pageMap[newPage.alias.substring(0, 1).toUpperCase() + newPage.alias.substring(1)] = newPage;
			}
			return newPage;
		}
		// Merge each variable.
		for (var k in oldPage) {
			oldPage[k] = stateService.smartMerge(oldPage[k], newPage[k]);
		}
		return oldPage;
	};

	// Remove page with the given pageId from the global pageMap.
	this.removePageFromMap = function(pageId) {
		delete stateService.pageMap[pageId];
	};

	// Add the given page to the global editMap.
	this.addPageToEditMap = function(page) {
		stateService.editMap[page.pageId] = setUpPage(page);
	};

	this.addPageToDeletedPagesMap = function(page) {
		stateService.deletedPagesMap[page.pageId] = setUpPage(page);
	};

	// Remove page with the given pageId from the global editMap;
	this.removePageFromEditMap = function(pageId) {
		delete stateService.editMap[pageId];
	};

	// Return function for sorting children ids.
	this.getChildSortFunc = function(sortChildrenBy) {
		var pageMap = stateService.pageMap;
		if (sortChildrenBy === 'alphabetical') {
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
		} else if (sortChildrenBy === 'recentFirst') {
			return function(aId, bId) {
				return pageMap[bId].pageCreatedAt.localeCompare(pageMap[aId].pageCreatedAt);
			};
		} else if (sortChildrenBy === 'oldestFirst') {
			return function(aId, bId) {
				return pageMap[aId].pageCreatedAt.localeCompare(pageMap[bId].pageCreatedAt);
			};
		} else {
			if (sortChildrenBy !== 'likes') {
				console.error('Unknown sort type: ' + sortChildrenBy);
			}
			return function(aId, bId) {
				var diff = pageMap[bId].likeScore() - pageMap[aId].likeScore();
				if (diff === 0) {
					return pageMap[bId].pageCreatedAt.localeCompare(pageMap[aId].pageCreatedAt);
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

		var successFn = function(data) {
			var pageData = data.pages;
			for (var id in pageData) {
				delete loadingPageAliases[options.url + id];
				delete loadingPageAliases[options.url + pageData[id].alias];
			}
			if (options.success) options.success();
		};
		var errorFn = function(data) {
			if (options.error) options.error(data);
			return options.silentFail;
		};
		stateService.postData(options.url, {pageAlias: pageAlias}, successFn, errorFn);
	};

	// Get data to display a popover for the page with the given alias.
	this.loadIntrasitePopover = function(pageAlias, options) {
		options = options || {};
		options.url = '/json/intrasitePopover/';
		loadPage(pageAlias, options);
		analyticsService.reportPopover(pageAlias);
	};

	// Get data to display a lens.
	this.loadLens = function(pageAlias, options) {
		options = options || {};
		options.url = '/json/lens/';
		loadPage(pageAlias, options);
	};

	// Get data to display page's title
	this.loadTitle = function(pageAlias, options) {
		options = options || {};
		options.url = '/json/title/';
		loadPage(pageAlias, options);
	};

	// Get data to display a comment thread.
	this.loadCommentThread = function(commentId, options) {
		options = options || {};
		options.url = '/commentThread/';
		loadPage(commentId, options);
	};

	// Load edit.
	// options {
	// 	pageAlias: pageAlias to load
	// 	specificEdit: load page with this edit number
	// 	editLimit: only load edits lower than this number
	// 	createdAtLimit: only load edits that were created before this date
	//	additionalPageIds: optional array of pages to load (e.g. for quick parent)
	// 	skipProcessDataStep: if true, we don't process the data we get from the server
	// 	convertPageIdsToAliases: false by default
	// 	success: callback on success
	// 	error: callback on error
	// }
	this.loadEdit = function(options) {
		// Set up options.
		var success = options.success; delete options.success;
		var error = options.error; delete options.error;
		var skipProcessDataStep = options.skipProcessDataStep; delete options.skipProcessDataStep;

		stateService.postDataWithOptions('/json/edit/',
				{
					pageAlias: options.pageAlias,
					specificEdit: options.specificEdit,
					editLimit: options.editLimit,
					createdAtLimit: options.createdAtLimit,
					additionalPageIds: options.additionalPageIds,
				},
				{
					callCallbacks: !options.skipProcessDataStep,
				},
				function(data) {
					if (options.convertPageIdsToAliases) {
						data.edits[options.pageAlias].text = that.convertPageIdsToAliases(data.edits[options.pageAlias].text);
					}
					success(data);
				},
				error);
	};

	this.convertPageIdsToAliases = function(textToConvert) {
		// Convert all links with pageIds to alias links.
		return textToConvert.replace(complexLinkRegexp, function(whole, prefix, text, alias) {
			var page = stateService.pageMap[alias];
			if (page) {
				return prefix + '[' + text + '](' + page.alias + ')';
			}
			return whole;
			/*}).replace(voteEmbedRegexp, function (whole, prefix, alias) {
				var page = stateService.pageMap[alias];
				if (page) {
				return prefix + '[vote: ' + page.alias + ']';
				}
				return whole;*/
		}).replace(forwardLinkRegexp, function(whole, prefix, alias, text) {
			var page = stateService.pageMap[alias];
			if (page) {
				return prefix + '[' + page.alias + ' ' + text + ']';
			}
			return whole;
		}).replace(simpleLinkRegexp, function(whole, prefix, alias) {
			if (alias.substring(0, 1) == '-') {
				var page = stateService.pageMap[alias.substring(1)];
				if (page) {
					return prefix + '[-' + page.alias + ']';
				}
			} else if (alias.substring(0, 1) == '+') {
				var page = stateService.pageMap[alias.substring(1)];
				if (page) {
					return prefix + '[+' + page.alias + ']';
				}
			} else {
				var page = stateService.pageMap[alias];
				if (page) {
					return prefix + '[' + page.alias + ']';
				}
			}
			return whole;
		}).replace(atAliasRegexp, function(whole, prefix, alias) {
			var page = stateService.pageMap[alias];
			if (page) {
				return prefix + '[@' + page.alias + ']';
			}
			return whole;
		});
	};

	// Get a new page from the server.
	// options {
	//  type: type of the page to create
	//  parentIds: optional array of parents to add to the new page
	//  isEditorComment: if true, will create an editor-only comment
	//	success: callback on success
	//	error: callback on error
	// }
	this.getNewPage = function(options) {
		var successFn = options.success; delete options.success;
		var errorFn = options.error; delete options.error;
		stateService.postData('/json/newPage/',
			options,
			function(data) {
				var pageId = Object.keys(data.edits)[0];
				successFn(pageId);
			},
			errorFn);
	};

	// Delete the page with the given pageId.
	this.deletePage = function(pageId, successFn, errorFn) {
		var data = {
			pageId: pageId,
		};
		stateService.postDataWithoutProcessing('/deletePage/', data, successFn, errorFn);
	};

	// Discard the page with the given id.
	this.discardPage = function(pageId, success, error) {
		var data = {
			pageId: pageId,
		};
		$http({method: 'POST', url: '/discardPage/', data: JSON.stringify(data)})
		.success(function(data, status) {
			if (success) success(data, status);
		})
		.error(function(data, status) {
			console.log('Error discarding ' + pageId + ':'); console.log(data); console.log(status);
			if (error) error(data, status);
		}
		);
	};

	// Save page's info.
	this.savePageInfo = function(page, callback) {
		$http({method: 'POST', url: '/editPageInfo/', data: JSON.stringify(page.getPageInfo())})
		.success(function(data) {
			if (callback) callback();
		})
		.error(function(data) {
			console.error('Error /editPageInfo/:'); console.error(data);
			if (callback) callback(data);
		});
	};

	// Add a new relationship between pages using the given params.
	this.newPagePair = function(params, successFn, errorFn) {
		stateService.postDataWithoutProcessing('/newPagePair/', params, successFn, errorFn);
	};
	// Note: you also need to specify the type of the relationship here, since we
	// don't want to accidentally delete the wrong type.
	this.deletePagePair = function(params, successFn, errorFn) {
		stateService.postDataWithoutProcessing('/deletePagePair/', params, successFn, errorFn);
	};
	// Update an existing page pair relationship
	this.updatePagePair = function(params, successFn, errorFn) {
		stateService.postDataWithoutProcessing('/updatePagePair/', params, successFn, errorFn);
	};

	// TODO: make these into page functions?
	// Return true iff we should show that this page is public.
	this.showPublic = function(pageId, useEditMap) {
		var page = stateService.getPageFromSomeMap(pageId, useEditMap);
		if (!page) {
			console.error('Couldn\'t find pageId: ' + pageId);
			return false;
		}
		return stateService.privateGroupId !== page.seeGroupId && page.seeGroupId === '';
	};
	// Return true iff we should show that this page belongs to a group.
	this.showPrivate = function(pageId, useEditMap) {
		var page = stateService.getPageFromSomeMap(pageId, useEditMap);
		if (!page) {
			console.error('Couldn\'t find pageId: ' + pageId);
			return false;
		}
		return stateService.privateGroupId !== page.seeGroupId && page.seeGroupId !== '';
	};

	// Create a new comment; optionally it's a reply to the given commentId
	// options: {
	//  parentPageId: id of the parent page
	//	replyToId: (optional) comment id this will be a reply to
	//	isEditorComment: if true, this will be created as an editor only comment
	//	success: callback
	// }
	this.newComment = function(options) {
		var parentIds = [options.parentPageId];
		if (options.replyToId) {
			parentIds.push(options.replyToId);
		}
		// Create new comment
		this.getNewPage({
			type: 'comment',
			parentIds: parentIds,
			// For now, all comments are editor-only
			isEditorComment: true,//!stateService.pageMap[options.parentPageId].permissions.comment.has,
			success: function(newCommentId) {
				if (options.success) {
					options.success(newCommentId);
				}
			},
		});
	};

	// Called when the user created a new comment.
	this.newCommentCreated = function(commentId) {
		var comment = stateService.editMap[commentId];
		if (comment.isEditorComment) {
			stateService.setShowEditorComments(true);
		}
		comment = this.addPageToMap(comment);
		// HACK: set the comment's data to make sure it's displayed correctly
		// TODO: actually fetch the newly created comment from the server
		comment.pageCreatedAt = moment().utc().format('YYYY-MM-DD HH:mm:ss');
		comment.permissions.comment.has = true;
		comment.isSubscribed = true;

		// If this comment is a reply, we add it to the parent comment. Otherwise we
		// add it to the lens its on.
		var parent;
		for (var n = 0; n < comment.parentIds.length; n++) {
			var p = stateService.pageMap[comment.parentIds[n]];
			if (!parent || p.isComment()) {
				parent = p;
			}
		}
		if (parent.subpageIds.indexOf(commentId) < 0) {
			parent.subpageIds.push(commentId);
		}
		// Only change the URL if we are on the actual lens page, since there are
		// ways to create new comments from other locations (e.g. discussion mode)
		if (stateService.primaryPage && stateService.primaryPage.pageId == comment.getCommentParentPage().pageId) {
			urlService.goToUrl(urlService.getPageUrl(commentId));
		}
	};

	// Make the given comment thread as (un)resolved.
	this.resolveThread = function(commentId, unresolve) {
		var data = {
			pageId: commentId,
			unresolve: unresolve,
		};
		stateService.postData('/json/resolveThread/', data, function(data) {
			stateService.pageMap[commentId].isResolved = !unresolve;
		});
	};

	// Approve the page that was submitted to the given domain.
	this.approvePageToDomain = function(pageId, domainId, successFn, errorFn) {
		var data = {
			pageId: pageId,
			domainId: domainId,
		};
		stateService.postData('/json/approvePageToDomain/', data, function(data) {
			var page = stateService.pageMap[pageId];
			if (page.domainIds.indexOf(data.domainId) < 0) {
				// The page is now part of the domain, even though it hasn't propagated yet
				page.domainIds.push(data.domainId);
			}
			if (successFn) successFn(data);
		}, errorFn);
	};

	// Approve the edit proposal associated with the given changelog
	this.approveEditProposal = function(changeLog, dismiss) {
		stateService.postDataWithoutProcessing('/json/approvePageEditProposal/', {
			changeLogId: changeLog.id,
			dismiss: dismiss,
		}, function(data) {
			changeLog.type = 'newEdit';
		});
	};

	this.getQualityTag = function(tagIds: string[]): string {
		if (tagIds.includes('72')) {
			return 'stub';
		}
		if (tagIds.includes('3rk')) {
			return 'start';
		}
		if (tagIds.includes('4y7')) {
			return 'c-class';
		}
		if (tagIds.includes('4yd')) {
			return 'b-class';
		}
		if (tagIds.includes('4yf')) {
			return 'a-class';
		}
		if (tagIds.includes('4yl')) {
			return 'featured';
		}

		return 'unassessed_meta_tag'
	}

	this.getQualityTagId = function(tagIds: string[]): string {
		if (tagIds.includes('72')) {
			return '72';
		}
		if (tagIds.includes('3rk')) {
			return '3rk';
		}
		if (tagIds.includes('4y7')) {
			return '4y7';
		}
		if (tagIds.includes('4yd')) {
			return '4yd';
		}
		if (tagIds.includes('4yf')) {
			return '4yf';
		}
		if (tagIds.includes('4yl')) {
			return '4yl';
		}

		return '4ym'
	};

	// Convert "alias_text" into "Alias text";
	this.getPrettyAlias = function(alias: string): string {
		let aliasWithSpaces = alias.replace(/_/g, ' ');
		return aliasWithSpaces.charAt(0).toUpperCase() + aliasWithSpaces.slice(1);
	};

	// Extract TODOs from the page's text
	this.todoBlockRegexpStr = '(%+)todo: ?([\\s\\S]+?)\\1 *(?=$|\Z|\n)';
	this.todoSpanRegexpStr = notEscaped + '\\[todo: ?([^\\]]+?)\\]' + noParen;
	this.computeTodos = function(page) {
		var todoBlockRegexp = new RegExp(that.todoBlockRegexpStr, 'gm');
		var todoSpanRegexp = new RegExp(that.todoSpanRegexpStr, 'g');
		if (page.todos.length > 0) return;
		page.todos = [];
		let match = todoBlockRegexp.exec(page.text);
		while (match != null) {
			page.todos.push(match[2]);
			match = todoBlockRegexp.exec(page.text);
		}

		match = todoSpanRegexp.exec(page.text);
		while (match != null) {
			page.todos.push(match[2]);
			match = todoSpanRegexp.exec(page.text);
		}
	};
});
