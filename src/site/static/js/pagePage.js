// pageView controls various portions of the page like the navigation and RHS columns.
var pageView;
$(function() {
	pageView = new function() {
		var showingNavigation = true;
		var showingRhs = false;
		var $navigation = $(".navigation-column");
		var $questionDiv = $(".question-div");
		var $rhsButtonsDiv = $(".rhs-buttons-div");
		var $newInlineCommentDiv = $(".new-inline-comment-div");
		var $commentButton = $newInlineCommentDiv.find(".new-inline-comment-button");

		// Delete any expanded inline comments or inline comment editors.
		this.clearRhs = function() {
			$questionDiv.find("arb-edit-page").remove();
			$questionDiv.find("arb-comment").remove();
			$(".inline-comment-icon").removeClass("on");
		};
	
		// Show right hand side and call the callback after the animation has played.
		this.showRhs = function(callback) {
			if (showingRhs) {
				callback();
				return;
			}
			$questionDiv.find(".toggle-inline-comment-div").animate({"left": "-=30%"});
			$rhsButtonsDiv.hide();
			$questionDiv.animate({"width": "30%"}, {queue: false});
			$navigation.animate({"margin-left": "-30%"}, {queue: false, complete: function() {
				showingRhs = true;
				if (callback) callback();
			}});
		};

		// Hide RHS.
		this.hideRhs = function(callback) {
			if (!showingRhs) {
				callback();
				return;
			}
			$questionDiv.animate({"width": "14%"}, {queue: false});
			$questionDiv.find(".toggle-inline-comment-div").animate({"left": "+=30%"});
			$navigation.animate({"margin-left": "0%"}, {queue: false, complete: function() {
				showingRhs = false;
				$rhsButtonsDiv.show();
				$(".inline-comment-highlight").removeClass("inline-comment-highlight");
				if (callback) callback();
			}});
		};

		// Hide/show an inline comment.
		this.toggleInlineComment = function($toggleDiv, callback) {
			var $inlineComment = $toggleDiv.find(".inline-comment-icon");
			if ($inlineComment.hasClass("on")) {
				this.clearRhs();
				this.hideRhs();
			} else {
				this.clearRhs();
				$(".inline-comment-highlight").removeClass("inline-comment-highlight");
				this.showRhs(function() {
					var offset = {left: $questionDiv.offset().left + 32, top: $toggleDiv.offset().top + 40};
					$(".inline-comment-div").offset(offset);
					$inlineComment.addClass("on");
					callback();
				});
			}
		};

		// Show the edit inline comment box.
		this.showEditInlineComment = function($scope, selection) {
			this.clearRhs();
			$(".toggle-inline-comment-div").hide();
			this.showRhs(function() {
				var $newInlineCommentDiv = $(".new-inline-comment-div");
				var offset = {left: $questionDiv.offset().left + 30, top: $(".inline-comment-highlight").offset().top};
				$(".inline-comment-div").offset(offset);
				createEditCommentDiv($(".inline-comment-div"), $newInlineCommentDiv, $scope, {
					anchorContext: selection.context,
					anchorText: selection.text,
					anchorOffset: selection.offset,
					primaryPageId: newInlineCommentPrimaryPageId,
					callback: function() {
						pageView.clearRhs();
						pageView.hideRhs(function() {
							$(".toggle-inline-comment-div").hide();
						});
					},
				});
			});
		};

		// Store the primary page id used for creating a new inline comment.
		var newInlineCommentPrimaryPageId;
		this.setNewInlineCommentPrimaryPageId = function(id) {
			newInlineCommentPrimaryPageId = id;
		};
	}();
});

// MainCtrl is for the Page page.
app.controller("MainCtrl", function($scope, $compile, $location, pageService, userService) {
	$scope.pageService = pageService;
	$scope.$compile = $compile;
	$scope.relatedIds = gRelatedIds;
	$scope.answerIds = [];
	$scope.page = pageService.primaryPage;

	// Set up children pages and question ids.
	$scope.initialChildren = {};
	$scope.initialChildrenCount = 0;
	for (var n = 0; n < $scope.page.children.length; n++) {
		var id = $scope.page.children[n].childId;
		var page = pageService.pageMap[id];
		if (page.type === "question") {
			// Do nothing, process them in pageController.
		} else if (page.type === "answer") {
			$scope.answerIds.push(id);
		} else if (page.type === "comment") {
			// do nothing
		} else {
			$scope.initialChildren[id] = page;
			$scope.initialChildrenCount++;
		}
	}

	// Set up parents pages.
	$scope.initialParents = {};
	$scope.initialParentsCount = $scope.page.parents.length;
	for (var n = 0; n < $scope.initialParentsCount; n++) {
		var id = $scope.page.parents[n].parentId;
		$scope.initialParents[id] = pageService.pageMap[id];
	}

	// Question button stuff.
	keepDivFixed($(".rhs-buttons-div"));

	// Process question button click.
	$(".question-button").on("click", function(event) {
		if (userService.user.id === "0") {
			showSignupPopover($(event.currentTarget));
			return true;
		}
		$(document).trigger("new-page-modal-event", {
			modalKey: "newQuestion",
			parentPageId: pageService.primaryPage.pageId,
			callback: function(result) {
				if (result.abandon) {
					$scope.$apply(function() {
						pageService.primaryPage.childDraftId = 0;
					});
				} else if (result.hidden) {
					$scope.$apply(function() {
						pageService.primaryPage.childDraftId = result.alias;
					});
				} else {
					window.location.href = "/pages/" + result.alias;
				}
			},
		});
	});

	// Inline comment button stuff.
	var $newInlineCommentDiv = $(".new-inline-comment-div");
	var $commentButton = $newInlineCommentDiv.find(".new-inline-comment-button");
	// Process new inline comment button click.
	$commentButton.on("click", function(event) {
		$(".inline-comment-highlight").removeClass("inline-comment-highlight");
		var selection = getSelectedParagraphText();
		if (selection) {
			pageView.showEditInlineComment($scope, selection);
		}
		return false;
	});

	// Add answers pages.
	var $answersList = $(".answers-list");
	for (var n = 0; n < $scope.answerIds.length; n++){
		var el = $compile("<arb-page page-id='" + $scope.answerIds[n] + "'></arb-page><hr></hr>")($scope);
		$answersList.append(el);
	}

	// Add edit page for the answer.
	if ($scope.page.type === "question") {
		$scope.answerDoneFn = function(result) {
			if (result.abandon) {
				getNewAnswerId();
			} else if (result.alias) {
				window.location.assign($scope.page.url() + "#page-" + result.alias);
				window.location.reload();
			}
		};

		var createAnswerEditPage = function(page) {
			var el = $compile("<arb-edit-page page-id='" + page.pageId +
				"' primary-page-id='" + $scope.page.pageId +
				"' done-fn='answerDoneFn(result)'></arb-edit-page>")($scope);
			$(".new-answer").append(el);
		};
		var getNewAnswerId = function() {
			$(".new-answer").find("arb-edit-page").remove();
			pageService.loadPages([], {
				success: function(data, status) {
					var page = pageService.pageMap[Object.keys(data)[0]];
					page.group = $.extend({}, $scope.page.group);
					page.type = "answer";
					page.parents = [{parentId: $scope.page.pageId, childId: page.pageId}];
					createAnswerEditPage(page);
				},
			});
		};
		if ($scope.page.childDraftId > 0) {
			createAnswerEditPage(pageService.pageMap[$scope.page.childDraftId]);
		} else {
			getNewAnswerId();
		}
	}

	// Toggle between lenses.
	var performSwitchToLens = function(lensPage) {
		pageService.setPrimaryPage(lensPage);
		// Sigh. This generates an error, but it seems benign.
		var url = window.location.pathname + "?lens=" + lensPage.pageId + window.location.hash;
		history.pushState(null, lensPage.title + " - Arbital", url);
	};
	var switchToLens = function(lensId, callback) {
		var lensPage = pageService.pageMap[lensId];
		if (!lensPage) return;
		if (lensPage.text.length > 0) {
			performSwitchToLens(lensPage);
			if(callback) callback();
		} else {
			pageService.loadPages([lensId], {
				includeText: true,
				includeAuxData: true,
				loadComments: true,
				loadVotes: true, 
				loadChildren: true,
				loadChildDraft: true,
				overwrite: true,
				success: function(data, status) {
					var page = pageService.pageMap[lensId];
					var el = $compile("<arb-page page-id='" + page.pageId + "'></arb-page>")($scope);
					$("#lens-" + page.pageId).empty().append(el);
					performSwitchToLens(page);
					if(callback) callback();
				},
			});
		}
	};
	$(".lens-tabs").on("click", ".lens-tab", function(event) {
		var $tab = $(event.currentTarget);
		var lensId = $tab.attr("data-target");
		lensId = lensId.substring(lensId.indexOf("-") + 1);
		switchToLens(lensId);
		$scope.$apply();
		return true;
	});
	// Process url ?lens parameter.
	var searchLensId = $location.search().lens;
	if (searchLensId) {
		switchToLens(searchLensId, function() {
			$("[data-target='#lens-" + searchLensId + "']").tab("show");
		});
	}
});
