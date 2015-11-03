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
			$questionDiv.find("arb-question").remove();
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

		// Hide/show an inline subpage.
		this.toggleInlineSubpage = function($toggleDiv, callback) {
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

		// Show the edit inline subpage box.
		this.showEditInlineSubpage = function($scope, selection, divType) {
			this.clearRhs();
			$(".toggle-inline-comment-div").hide();
			this.showRhs(function() {
				var $newInlineCommentDiv = $(".new-inline-comment-div");
				var offset = {left: $questionDiv.offset().left + 30, top: $(".inline-comment-highlight").offset().top};
				$(".inline-comment-div").offset(offset);
				createEditSubpageDiv($(".inline-comment-div"), $newInlineCommentDiv, $scope, {
					anchorContext: selection.context,
					anchorText: selection.text,
					anchorOffset: selection.offset,
					primaryPageId: newInlineCommentPrimaryPageId,
					divType: divType,
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

// Directive for the entire primary page.
app.directive("arbPrimaryPage", function($compile, $location, $timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/primaryPage.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.$compile = $compile;
			scope.answerIds = [];
			scope.page = pageService.primaryPage;
		
			// Set up children pages and question ids.
			scope.initialChildren = {};
			scope.initialChildrenCount = 0;
			for (var n = 0; n < scope.page.children.length; n++) {
				var id = scope.page.children[n].childId;
				var page = pageService.pageMap[id];
				if (!page) {
					continue;
				}
				if (page.type === "question") {
					// Do nothing, process them in pageController.
				} else if (page.type === "answer") {
					scope.answerIds.push(id);
				} else if (page.type === "comment" || page.type === "lens") {
					// do nothing
				} else {
					scope.initialChildren[id] = page;
					scope.initialChildrenCount++;
				}
			}
		
			// Set up parents pages.
			scope.initialParents = {};
			scope.initialParentsCount = scope.page.parents.length;
			for (var n = 0; n < scope.initialParentsCount; n++) {
				var id = scope.page.parents[n].parentId;
				scope.initialParents[id] = pageService.pageMap[id];
			}
		
			// Question button stuff.
			keepDivFixed($(".rhs-buttons-div"));
		
			// Process question button click.
			$(".question-button").on("click", function(event) {
				if (userService.user.id === "0") {
					showSignupPopover($(event.currentTarget));
					return true;
				}
				var selection = getSelectedParagraphText();
				if (selection) {
					pageView.showEditInlineSubpage(scope, selection, "question");
				} else {
					$(document).trigger("new-page-modal-event", {
						modalKey: "newQuestion",
						parentPageId: pageService.primaryPage.pageId,
						callback: function(result) {
							if (result.abandon) {
						 		scope.$apply(function() {
									pageService.primaryPage.childDraftId = 0;
						 		});
							} else if (result.hidden) {
						 		scope.$apply(function() {
									pageService.primaryPage.childDraftId = result.alias;
						 		});
					 	 	} else {
						 		window.location.href = "/pages/" + result.alias;
					 	 	}
						},
					});
				}
			});
		
			// Inline comment button stuff.
			var $newInlineCommentDiv = $(".new-inline-comment-div");
			var $commentButton = $newInlineCommentDiv.find(".new-inline-comment-button");
			// Process new inline comment button click.
			$commentButton.on("click", function(event) {
				$(".inline-comment-highlight").removeClass("inline-comment-highlight");
				var selection = getSelectedParagraphText();
				if (selection) {
					pageView.showEditInlineSubpage(scope, selection, "comment");
				}
				return false;
			});
		
			// Add answers pages.
			var $answersList = $(".answers-list");
			for (var n = 0; n < scope.answerIds.length; n++){
				var el = $compile("<arb-page page-id='" + scope.answerIds[n] + "'></arb-page><hr></hr>")(scope);
				$answersList.append(el);
			}
		
			// Set up finding existing answer for question pages.
			if (pageService.primaryPage.type === "question") {
				scope.findAnswerTerm = "";
				// Get similar pages
				var prevFindAnswerTerm = "";
				var $foundAnswers = $("#found-answers");
				var findAnswerTermChanged = createThrottledCallback(function() {
					if (scope.findAnswerTerm.length <= 2) return false;
					var options = {
						term: scope.findAnswerTerm,
						pageType: "answer",
					};
					if (options.term === prevFindAnswerTerm) return false;
					autocompleteService.performSearch(options, function(results){
						$foundAnswers.empty();
						for (var n = 0; n < results.length; n++) {
							var pageId = results[n].value;
							var $el = $compile("<button class='suggest-answer btn btn-info' answer-id='" + pageId + "'>Suggest</button><span arb-likes-page-title page-id='" + pageId +
								"' show-clickbait='true'></span>")(scope);
							$foundAnswers.append($el);
						}
					});
					return true;
				}, 300);
				scope.$watch("findAnswerTerm", findAnswerTermChanged);
		
				// User clicks to suggest an answer
				$("body").on("click", ".suggest-answer", function(event) {
					var answerId = $(event.target).attr("answer-id");
					pageService.newPagePair({
						parentId: pageService.primaryPage.pageId,
						childId: answerId,
						type: "parent",
					}, function() {
						location.reload();
					});
				});
			}
		
			// Add edit page for the answer.
			if (scope.page.type === "question") {
				// Callback for when we are done with editing the answer.
				scope.answerDoneFn = function(result) {
					if (result.abandon) {
						getNewAnswerId();
					} else if (result.alias) {
						window.location.assign(scope.page.url() + "#page-" + result.alias);
						window.location.reload();
					}
				};
		
				// Create the edit answer form
				var createAnswerEditPage = function(answerId) {
					var el = $compile("<arb-find-answer></arb-find-answer>")(scope);
					$(".new-answer").append(el);
		
					el = $compile("<arb-edit-page page-id='" + answerId +
						"' primary-page-id='" + scope.page.pageId +
						"' done-fn='answerDoneFn(result)'></arb-edit-page>")(scope);
					$(".new-answer").append(el);
				};
				// Request a new page id from the server, and then create a new edit answer form
				var getNewAnswerId = function() {
					$(".new-answer").find("arb-edit-page").remove();
					pageService.getNewPage({
						success: function(newPageId) {
							var page = pageService.editMap[newPageId];
							page.group = $.extend({}, scope.page.group);
							page.type = "answer";
							page.parents = [{parentId: scope.page.pageId, childId: page.pageId}];
							createAnswerEditPage(newPageId);
						},
					});
				};
				if (scope.page.childDraftId > 0) {
					// Load the edit to resume editing it
					pageService.loadEdit({
						pageAlias: scope.page.childDraftId,
						success: function(data, status) {
							createAnswerEditPage(scope.page.childDraftId);
						},
					});
				} else {
					getNewAnswerId();
				}
			}
		
			// Toggle between lenses.
			var performSwitchToLens = function(lensPage) {
				pageService.setPrimaryPage(lensPage);
				// Sigh. This generates an error, but it seems benign.
				console.log("==== Error might be generated, but it's not actually an error.... I think ====");
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
					pageService.loadLens(lensId, {
						success: function(data, status) {
							var page = pageService.pageMap[lensId];
							var el = $compile("<arb-page page-id='" + page.pageId + "'></arb-page>")(scope);
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
				scope.$apply();
				return true;
			});
			// Process url ?lens parameter.
			var searchLensId = $location.search().lens;
			if (searchLensId && searchLensId != pageService.primaryPage.pageId) {
				switchToLens(searchLensId, function() {
					$("[data-target='#lens-" + searchLensId + "']").tab("show");
				});
			}
		},
	};
});

// Directive for showing a the panel with tags.
app.directive("arbTagsPanel", function(pageService, userService, autocompleteService, $timeout, $http) {
	return {
		templateUrl: "/static/html/tagsPanel.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.primaryPage;
			if (!scope.page.taggedAsIds) {
				scope.page.taggedAsIds = [];
			}
			
			// Setup autocomplete for input field.
			autocompleteService.setupParentsAutocomplete(element.find(".tag-input"), function(event, ui) {
				var data = {
					parentId: ui.item.label,
					childId: scope.page.pageId,
					type: "tag",
				};
				$http({method: "POST", url: "/newPagePair/", data: JSON.stringify(data)})
					.error(function(data, status){
						console.error("Error creating tag:"); console.log(data); console.log(status);
					});

				scope.page.taggedAsIds.push(data.parentId);
				scope.$apply();
				$(event.target).val("");
				return false;
			});

			// Process deleting tags.
			element.on("click", ".delete-tag-link", function(event) {
				var $target = $(event.target);
				var options = {
					parentId: $target.attr("page-id"),
					childId: scope.page.pageId,
					type: "tag",
				};
				pageService.deletePagePair(options);

				scope.page.taggedAsIds.splice(scope.page.taggedAsIds.indexOf(options.parentId), 1);
				scope.$apply();
			});

			$timeout(function() {
				// Set the rendering for tags autocomplete
				autocompleteService.setAutocompleteRendering(element.find(".tag-input"), scope);
			});
		},
	};
});
