"use strict";

// Create new EditPage (for use with znd-edit-page directive).
// page - page object corresponding to the page being edited.
// pageService - pageService object which contains all loaded pages.
// $topParent - points to the top DOM element of the znd-edit-page directive.
// primaryPage - set for modal edit page and creating a new answer on a question page; points to the primary page being edited / to the question.
var EditPage = function(page, pageService, autocompleteService, $topParent, primaryPage) {
	var page = page;
	var pageId = page.PageId; // id of the page we are editing
	var primaryPage = primaryPage;

	// Create a new tag for the page.
	var createNewParentElement = function(parentAlias) {
		parentAlias = autocompleteService.convertInputToAlias(parentAlias);
		// TODO: double check there isn't this parent already
		var $template = $topParent.find(".tag.template");
		var $newTag = $template.clone(true);
		if (parentAlias in autocompleteService.aliasMap) {
			var parentPageId = autocompleteService.aliasMap[parentAlias].PageId;
			var title = autocompleteService.aliasMap[parentAlias].PageTitle;
		} else if (parentAlias in pageService.pageMap) {
			var parentPageId = parentAlias;
			var title = pageService.pageMap[parentAlias].Title;
			parentAlias = pageService.pageMap[parentAlias].Alias;
		} else if (primaryPage !== undefined) {
			// The parent is the primaryPage.
			var parentPageId = primaryPage.PageId;
			var title = primaryPage.Title;
			if (title === "") {
				title = "*Untitled*";
			}
		} else {
			// The parent hasn't been published yet.
			var parentPageId = parentAlias;
			var title = "*Unpublished*";
		}
		$newTag.removeClass("template");
		$newTag.text(title);
		$newTag.attr("tag-id", parentPageId);
		$newTag.attr("title", parentAlias).tooltip();
		$newTag.insertBefore($template);
	}
	var deleteParentElement = function($target) {
		$target.tooltip("destroy").remove();
	};

	// Helper function for savePage. Computes the data to submit via AJAX.
	var computeAutosaveData = function(isAutosave, isSnapshot) {
		var parentIds = [];
		$topParent.find(".parent-container").children(".tag:not(.template)").each(function(index, element) {
			parentIds.push($(element).attr("tag-id"));
		});
		var privacyKey = $topParent.attr("privacy-key");
		var data = {
			pageId: pageId,
			parentIds: parentIds.join(),
			privacyKey: privacyKey,
			keepPrivacyKey: $topParent.find("input[name='private']").is(":checked"),
			karmaLock: +$topParent.find(".karma-lock-slider").bootstrapSlider("getValue"),
			isAutosave: isAutosave,
			isSnapshot: isSnapshot,
			__invisibleSubmit: isAutosave,
		};
		var $form = $topParent.find(".new-page-form");
		serializeFormData($form, data);
		return data;
	};
	var autosaving = false;
	var publishing = false;
	var prevEditPageData = {};
	// Save the current page.
	// callback is called when the server replies with success. If it's an autosave
	// and the same data has already been submitted, the callback is called with "".
	var savePage = function(isAutosave, isSnapshot, callback) {
		// Prevent stacking up saves without them returning.
		if (publishing) return;
		publishing = !isAutosave && !isSnapshot;
		if (isAutosave && autosaving) return;
		autosaving = isAutosave;

		// Submit the form.
		var data = computeAutosaveData(isAutosave, isSnapshot);
		var $form = $topParent.find(".new-page-form");
		if (!isAutosave || JSON.stringify(data) !== JSON.stringify(prevEditPageData)) {
			// TODO: if the call takes too long, we should show a warning.
			submitForm($form, "/editPage/", data, function(r) {
				if (isAutosave) autosaving = false;
				callback(r);
			});
			prevEditPageData = data;
		} else {
			callback(undefined);
			autosaving = false;
		}
	}

	// Process form submission.
	$topParent.find(".new-page-form").on("submit", function(event) {
		var $target = $(event.target);
		var $body = $target.closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		savePage(false, false, function(r) {
			if (page.Type === "answer" && primaryPage) {
				window.location.assign(primaryPage.Url + "#page-" + pageId);
				window.location.reload();
			} else if (primaryPage !== undefined) {
				$(document).trigger("new-page-modal-closed-event", {alias: r.substring(r.lastIndexOf("/") + 1)});
			} else {
				window.location.replace(r);
			}
		});
		return false;
	});

	// Process safe snapshot button.
	$topParent.find(".save-snapshot-button").on("click", function(event) {
		var $body = $(event.target).closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		savePage(false, true, function(r) {
			if (r !== undefined) {
				$body.attr("privacy-key", r);
				$loadingText.show().text("Saved!");
			}
		});
		return false;
	});

	// Process New Page button.
	$topParent.find(".new-page-button").on("click", function(event) {
		$(document).trigger("new-page-modal-closed-event", {reopen: true});
	});

	// Process Abandon button.
	$topParent.find(".abandon-page-button").on("click", function(event) {
		$(document).trigger("new-page-modal-closed-event", {alias: pageId, abandon: true});
	});

	// Deleting parents. (Only inside the parent container.)
	$topParent.find(".parent-container .tag").on("click", function(event) {
		var $target = $(event.target);
		deleteParentElement($target);
		return false;
	});

	// Resive textarea height to fit the screen.
	$("#wmd-input").height($(window).height() - 140);

	// Scroll wmd-panel so it's always inside the viewport.
	if (primaryPage === undefined) {
		var $wmdPanelContainer = $topParent.find(".wmd-panel-container");
		var $wmdPreview = $topParent.find(".wmd-preview");
		var $wmdPanel = $topParent.find(".wmd-panel");
		var wmdPanelY = $wmdPanel.offset().top;
		var wmdPanelHeight = $wmdPanel.outerHeight();
		var initialContainerHeight = $wmdPanelContainer.height();
		$(window).scroll(function(){
			var y = $(window).scrollTop() - wmdPanelY;
			y = Math.min($wmdPreview.outerHeight() - wmdPanelHeight, y);
			y = Math.max(0, y);
			$wmdPanel.stop(true).animate({top: y}, "fast");
		});
		// Automatically adjust height of wmd-panel-container.
		var adjustHeight = function(){
			$wmdPanelContainer.height(Math.max(initialContainerHeight, $wmdPreview.height() + $wmdPreview.offset().top - $wmdPanelContainer.offset().top));
		};
		window.setInterval(adjustHeight, 1000);
		adjustHeight();
	}

	// Keep title label in sync with the title input.
	var $titleLabel = $topParent.find(".page-title-text");
	$topParent.find("input[name='title']").on("keyup", function(event) {
		$titleLabel.text($(event.target).val());
	});

	// Add parent tags.
	// usePageIds - forces pageIds to be passed to createNewParentElement. Used
	//   to create initial parent elments.
	var addParentTags = function(usePageIds) {
		var parentsLen = page.Parents.length;
		for(var n = 0; n < parentsLen; n++) {
			var parentPage = pageService.pageMap[page.Parents[n].ParentId];
			if (usePageIds || parentPage.Alias === "") {
				var parentKey = parentPage.PageId;
			} else {
				var parentKey = parentPage.Alias;
			}
			createNewParentElement(parentKey);
		}
	};

	// Set up parent options buttons.
	if (primaryPage !== undefined) {
		var deleteAllParentTags = function() {
			$topParent.find(".parent-container").children(".tag:not(.template)").each(function(index, element) {
				deleteParentElement($(element));
			});
		};
		$topParent.find(".child-parent-option").on("click", function(event) {
			deleteAllParentTags();
			page.Parents = primaryPage.Parents.slice();
			addParentTags();
			$(event.target).hide();
		});
	}

	// === Trigger initial setup. ===

	// Setup autocomplete for tags.
	autocompleteService.loadAliasSource(function() {
		$topParent.find(".tag-input").autocomplete({
			source: autocompleteService.aliasSource,
			minLength: 2,
			select: function (event, ui) {
				createNewParentElement(ui.item.label);
				$(event.target).val("");
				return false;
			}
		});
		// Set up Markdown.
		zndMarkdown.init(true, pageId, "", undefined, autocompleteService);
	});
	addParentTags(true);

	// Setup karma lock slider.
	var $slider = $topParent.find(".karma-lock-slider");
	$slider.bootstrapSlider({
		value: +$slider.attr("value"),
		min: 0,
		max: +$slider.attr("max"),
		step: 1,
		precision: 0,
		selection: "none",
		handle: "square",
		tooltip: "always",
	});

	// Change all dates from UTC to local.
	$topParent.find(".date").each(function(index, element) {
		var date = new Date(element.innerHTML + " UTC");
		element.innerHTML = date.toLocaleString();
	});

	// Start initializes things that have to be killed when this editPage stops existing.
	this.autosaveInterval = null;
	this.backdropInterval = null;
	this.start = function() {
		// Hide new page button if this is a modal.
		$topParent.find("#wmd-new-page-button" + pageId).toggle(primaryPage === undefined);
		// Set up autosaving.
		var $autosaveLabel = $topParent.find(".autosave-label");
		this.autosaveInterval = window.setInterval(function(){
			$autosaveLabel.text("Autosave: Saving...").show();
			savePage(true, false, function(r) {
				if (r === undefined) {
					$autosaveLabel.hide();
				} else {
					$("body").attr("privacy-key", r);
					$autosaveLabel.text("Autosave: Saved!").show();
				}
			});
		}, 5000);
		// Compute prvEditPageData, so we don't fire off autosave when there were
		// no changes made.
		prevEditPageData = computeAutosaveData(true, false);
		console.log(prevEditPageData);
		// Set up an interval to make sure modal backdrop is the right size.
		if (primaryPage) {
			var $element = $topParent.closest(".modal-content");
			if ($element.length > 0) {
				var lastHeight = 0;
				this.backdropInterval = window.setInterval(function(){
					if ($element.css("height") != lastHeight) {
						lastHeight = $element.css("height"); 
						$("#new-page-modal").data("bs.modal").handleUpdate();
					}
				}, 1000);
			}
		}
	};

	// Called before this editPage is destroyed.
	this.stop = function() {
		clearInterval(this.autosaveInterval);
		clearInterval(this.backdropInterval);
		// Snapshot just in case.
		savePage(true, false, function(r) {});
	};
};



// Directive for the modal, where a user can create a new page, edit a page, 
// ask a question, etc...
app.directive("zndEditPageModal", function (pageService, userService) {
	return {
		templateUrl: "/static/html/editPageModal.html",
		controller: function ($scope, $compile, $timeout, pageService, autocompleteService) {
			// Process event to create the new-page-modal.
			$(document).on("new-page-modal-event", function(e, modalType, newPageCallback){
				var isQuestion = modalType === "newQuestion";
				var $modal = $('#new-page-modal');
				var $modalBody = $modal.find(".modal-body");
				if (isQuestion) {
					$modal.find(".modal-title").text("Ask a Question");
				} else {
					$modal.find(".modal-title").text("New Page");
				}
				$modalBody.empty().append("<img src='/static/images/loading.gif'/>");
				
				// Setup modal.
				// Set to true if we should immediately reopen the modal with a new empty page
				var reopenNew = false;
				var setupModal = function(pageId, isResumed) {
					var newPage = pageService.pageMap[pageId];
					if (!isResumed) {
						if (isQuestion) {
							newPage.Type = "question";
						} else {
							newPage.Type = $scope.parentPage.Type;
						}
						newPage.Parents = [{ParentId: $scope.parentPage.PageId}];
						newPage.CreatorId = userService.user.Id;
						newPage.Group = $scope.parentPage.Group;
					}

					// Dynamically create znd-edit-page directive.
					var el = $compile("<znd-edit-page page-id='\"" + pageId +
							"\"' is-modal='true'></znd-edit-page>")($scope);
					$modalBody.empty().append(el);
					$modal.modal();

					// Handle modal's shown event to initialize editPage script.
					var editPage;
					$modal.on("shown.bs.modal", function (e) {
						editPage = new EditPage(newPage, pageService, autocompleteService, el, $scope.parentPage);
						editPage.start();
						$modal.find("input[name='title']").focus();
					});
					// Hande modal's close event, with the possibility of returning the
					// new page's alias.
					var returnedNewPageAlias = null;
					$modal.on("hidden.bs.modal", function (e) {
						newPageCallback(returnedNewPageAlias);
						editPage.stop();
						editPage = undefined;
						$modal.off("shown.bs.modal");
						$modal.off("hidden.bs.modal");
						$(document).off("new-page-modal-closed-event");
						$modalBody.empty();
						if (reopenNew) {
							$(document).trigger("new-page-modal-event", [modalType, newPageCallback]);
						}
					});
					// Allow the modal to "return" the alias for the newly created page.
					// options.abandon - if set to true, the page specified with 'alias' will be deleted
					// options.alias - set to alias/page id of the created page
					// options.reopen - if true, the modal will close and immediately reopen
					$(document).on("new-page-modal-closed-event", function(e, options){
						$modal.modal("hide");
						if (options.abandon) {
							pageService.deletePage(options.alias, function(){
								$scope.parentPage.QuestionDraftId = 0;
							});
						} else if (options.alias) {
							returnedNewPageAlias = options.alias;
						}
						reopenNew = options.reopen;
						$scope.newPageId = undefined;
					});
				};
				
				var loadPages = function() {
					var loadPagesIds = [];
					if ($scope.newPageId !== undefined) {
						// Resume editing some page.
						loadPagesIds = [$scope.newPageId];
						pageService.removePageFromMap($scope.newPageId);
					}
					pageService.loadPages(loadPagesIds,
						function(data, status) {
							$scope.newPageId = Object.keys(data)[0];
							if ($scope.newPageId !== $scope.parentPage.PageId) {  // let's not try to edit the same page in two places
								setupModal($scope.newPageId, loadPagesIds.length > 0);
							}
						}, function(data, status) {
							console.log("Couldn't load pages: " + loadPagesIds);
							if (loadPagesIds.length > 0) {
								// Let's try again, but without trying to resume editing.
								$scope.newPageId = undefined;
								loadPages();
							}
						}
					);
				};
				loadPages();
			});
		},
		scope: {
			parentPageId: "=",
			// Optionally set if we want to resume editing a page.
			resumePageId: "=",
		},
		link: function(scope, element, attrs) {
			scope.parentPage = pageService.pageMap[scope.parentPageId];
			// NewPageId is the id of the page we are editing in the modal.
			if (scope.resumePageId && scope.resumePageId != "0") {
				scope.newPageId = scope.resumePageId;
			} else {
				scope.newPageId = undefined;
			}
		},
	};
});

// Directive for the actual DOM elements which allows the user to edit a page.
app.directive("zndEditPage", function(pageService, userService, autocompleteService, $timeout) {
	return {
		templateUrl: "/static/html/editPage.html",
		scope: {
			pageId: "=",
			isModal: "=",
			questionId: "=", // set if this edit page is for an answer to this question
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			scope.useVerticalView = scope.isModal;

			// Set up page types.
			if (scope.page.Type === "question") {
				scope.pageTypes = {question: "Question"};
			} else if(scope.page.Type === "answer") {
				scope.pageTypes = {answer: "Answer"};
			} else {
				scope.pageTypes = {wiki: "Wiki Page", blog: "Blog Page"};
				scope.page.Type = scope.page.Type in scope.pageTypes ? scope.page.Type : "wiki";
			}

			// Set up sort types.
			scope.sortTypes = {
				likes: "By Likes",
				choronological: "Chronologically",
				alphabetical: "Alphabetically",
			};
			scope.page.SortChildrenBy = scope.page.SortChildrenBy in scope.sortTypes ? scope.page.SortChildrenBy : "likes";

			// Set up vote types.
			scope.voteTypes = {
				"":"",
				probability: "Probability",
				approval: "Approval",
			};
			scope.page.VoteType = scope.page.VoteType in scope.voteTypes ? scope.page.VoteType : "";

			var primaryPage = undefined;
			if (scope.questionId) {
				primaryPage = pageService.pageMap[scope.questionId];
			}
			if (scope.page.Type === "answer" && primaryPage) {
				// Set up answer page for when it appears on a question page.
				scope.page.Parents = [{ParentId: primaryPage.PageId}];
				scope.useVerticalView = true;
			}

			if (!scope.isModal) {
				// Create Edit Page JS controller.
				$timeout(function(){
					scope.editPage = new EditPage(scope.page, pageService, autocompleteService, element, primaryPage);
					scope.editPage.start();
				});
			}

			// Set up group names.
			var groupNames = userService.user.GroupNames;
			scope.groupOptions = {};
			if (groupNames) {
				for (var i in groupNames) {
					var group = groupNames[i];
					scope.groupOptions[group] = group;
				}
				scope.groupOptionsLength = groupNames.length;
			} else {
				scope.groupOptionsLength = 0;
			}

			// Get the text that should appear on the primary submit button.
			scope.getSubmitButtonText = function() {
				if (scope.page.Type == "answer") {
					return "Submit Answer";
				} else if (scope.page.Type === "question") {
					if (scope.page.WasPublished) {
						return "Edit Question";
					} else {
						return "Ask Question";
					}
				} else if (scope.isModal) {
					return "Publish & Link";
				}
				return "Publish";
			};
			// Get the text for the placeholder in the title input.
			scope.getTitlePlaceholder = function() {
				if (scope.page.Type == "answer") {
					return "Answer summary";
				} else if (scope.page.Type === "question") {
					return "Complete sentence question";
				}
				return "Page title";
			}
		},
	};
});
