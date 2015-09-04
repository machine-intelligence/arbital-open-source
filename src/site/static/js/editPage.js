"use strict";

// Create new EditPage (for use with znd-edit-page directive).
// page - page object corresponding to the page being edited.
// pageService - pageService object which contains all loaded pages.
// options {
//   topParent - points to the znd-edit-page DOM element.
//   primaryPage - for an answer page, points to the question page; for a comment, point to the root page
//   isModal - set if the page is being edited inside a modal
//   doneFn - function to call when the user is done with editing.
//			Function should return true iff the event was processed.
// }
// Result returned from doneFn {
// 	abandon - if set to true, the page specified with 'alias' will be deleted
// 	alias - set to alias/page id of the created page
// }
var EditPage = function(page, pageService, autocompleteService, options) {
	var page = page;
	var pageId = page.pageId; // id of the page we are editing
	var options = options || {};
	var primaryPage = options.primaryPage;
	var $topParent = options.topParent;
	var isModal = options.isModal;
	var doneFn = options.doneFn;

	// Create a new tag for the page.
	var createNewParentElement = function(parentAlias) {
		parentAlias = autocompleteService.convertInputToAlias(parentAlias);
		// TODO: double check there isn't this parent already
		var $template = $topParent.find(".tag.template");
		var $newTag = $template.clone(true);
		if (parentAlias in autocompleteService.aliasMap) {
			var parentPageId = autocompleteService.aliasMap[parentAlias].pageId;
			var title = autocompleteService.aliasMap[parentAlias].pageTitle;
		} else if (primaryPage !== undefined && parentAlias === primaryPage.alias) {
			// The parent is the primaryPage.
			var parentPageId = primaryPage.pageId;
			var title = primaryPage.title;
			if (title === "") {
				title = "*Untitled*";
			}
		} else if (parentAlias in pageService.pageMap) {
			var parentPageId = parentAlias;
			var title = pageService.pageMap[parentAlias].title;
			parentAlias = pageService.pageMap[parentAlias].alias;
		} else {
			// The parent hasn't been published yet.
			var parentPageId = parentAlias;
			var title = "*Not Yet Published*";
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
		serializeFormData($topParent.find(".new-page-form"), data);
		if (page.anchorContext) {
			data.anchorContext = page.anchorContext;
			data.anchorText = page.anchorText;
			data.anchorOffset = page.anchorOffset;
		}
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
			}, function() {
				if (isAutosave) autosaving = false;
				if (publishing) {
					publishing = false;
					// Pretend it was a failed autosave
					data.__invisibleSubmit = true; 
					data.isAutosave = true;
				}
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
			if (doneFn) {
				doneFn({alias: r.substring(r.lastIndexOf("/") + 1)});
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

	// Process Abandon button click.
	$topParent.find(".abandon-edit").on("click", function(event) {
		if (doneFn) {
			doneFn({alias: pageId, abandon: true});
		}
	});

	// Process Close button click.
	$topParent.find(".go-to-page-view").on("click", function(event) {
		if (doneFn) {
			doneFn({});
		}
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
	if (primaryPage === undefined && !isModal) {
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
		var parentsLen = page.parents.length;
		for(var n = 0; n < parentsLen; n++) {
			var parentPage = pageService.pageMap[page.parents[n].parentId];
			if (usePageIds || parentPage.alias === "") {
				var parentKey = parentPage.pageId;
			} else {
				var parentKey = parentPage.alias;
			}
			createNewParentElement(parentKey);
		}
	};

	// Set up parent options buttons.
	if (primaryPage !== undefined && isModal) {
		$topParent.find(".child-parent-option").on("click", function(event) {
			$topParent.find(".parent-container").children(".tag:not(.template)").each(function(index, element) {
				deleteParentElement($(element));
			});
			page.parents = primaryPage.parents.slice();
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
		zndMarkdown.init(true, pageId, "", undefined, pageService, autocompleteService);
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
		$topParent.find("#wmd-new-page-button" + pageId).toggle(!isModal);

		// Autofocus on some input.
		if (page.type !== "answer" || !primaryPage) {  
			window.setTimeout(function() {
				var $title = $topParent.find("input[name='title']");
				if ($title.is(":visible")) {
					$title.focus();
				} else {
					$topParent.find("textarea[name='text']").focus();
				}
			});
		}

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

		// Compute prevEditPageData, so we don't fire off autosave when there were
		// no changes made.
		prevEditPageData = computeAutosaveData(true, false);

		// Set up an interval to make sure modal backdrop is the right size.
		if (isModal) {
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
		// Autosave just in case.
		savePage(true, false, function(r) {});
	};
};



// Directive for the modal, where a user can create a new page, edit a page, 
// ask a question, etc...
app.directive("zndEditPageModal", function (pageService, userService) {
	return {
		templateUrl: "/static/html/editPageModal.html",
		scope: {
		},
		controller: function ($scope, $compile, $timeout, pageService, autocompleteService) {
			// Store which page was last edited. modalKey+primaryPageId -> pageId
			var pageIdCache = {};

			// Process event to create the new-page-modal.
			// options {
			//	modalKey: "newPage" or "newQuestion"
			//	parentPageId: the newly created page will have this page as a parent
			//	callback: function(result) to call when the user is done with the modal
			// }
			$(document).on("new-page-modal-event", function(e, options) {
				var primaryPage = pageService.pageMap[options.parentPageId];
				var resumePageId = pageIdCache[options.modalKey + primaryPage.pageId];
				var isQuestion = options.modalKey === "newQuestion";
				if (isQuestion && !resumePageId && primaryPage.childDraftId !== "0") {
					resumePageId = primaryPage.childDraftId;
				}
				var $modal = $("#new-page-modal");
				var $modalBody = $modal.find(".modal-body");
				$modal.find(".modal-title").text(isQuestion ? "Ask a Question" : "New Page");
				$modalBody.empty().append("<img src='/static/images/loading.gif'/>");
				
				// Setup modal.
				var setupModal = function(pageId, isResumed) {
					var newPage = pageService.pageMap[pageId];
					if (!isResumed) {
						if (isQuestion) {
							newPage.type = "question";
						}
						if (primaryPage.type !== "comment" &&
								primaryPage.type !== "answer" &&
								primaryPage.type !== "question") {
							newPage.parents = [{parentId: primaryPage.pageId}];
						}
						newPage.creatorId = userService.user.id;
						newPage.groupId = primaryPage.groupId;
					}

					// Dynamically create znd-edit-page directive.
					var el = $compile("<znd-edit-page page-id='" + pageId +
							"' is-modal='true'" +
							"done-fn='doneFn(result)'></znd-edit-page>")($scope);
					$modalBody.empty().append(el);
					$modal.modal();

					// Handle modal's shown event to initialize editPage script.
					// This result will be returned if the user just hides the modal.
					var returnedResult = {hidden: true, alias: pageId}; 
					var editPage;
					$modal.on("shown.bs.modal", function (e) {
						editPage = new EditPage(newPage, pageService, autocompleteService, {
							topParent: el,
							primaryPage: primaryPage,
							isModal: true,
							doneFn: function(result) {
								returnedResult = result;
								$modal.modal("hide");
								if (result.abandon) {
									pageService.deletePage(result.alias);
								}
								if (result.abandon || result.alias) {
									resumePageId = undefined;
								}
							},
						});
						editPage.start();
					});
					// Hande modal's close event and return the resulting alias.
					$modal.on("hidden.bs.modal", function (e) {
						pageIdCache[options.modalKey + pageService.primaryPage.pageId] = resumePageId
						if (options.callback) {
							options.callback(returnedResult);
						}
						editPage.stop();
						editPage = undefined;
						$modal.off("shown.bs.modal");
						$modal.off("hidden.bs.modal");
						$modalBody.empty();
					});
				};
				
				// Resume editing the page or get a new id from the server.
				var loadPages = function() {
					var loadPagesIds = [];
					if (resumePageId) {
						// Resume editing some page.
						loadPagesIds = [resumePageId];
						pageService.removePageFromMap(resumePageId);
					}
					pageService.loadPages(loadPagesIds, {
						includeText: true,
						allowDraft: true,
						success: function(data, status) {
							resumePageId = Object.keys(data)[0];
							// Let's not try to edit the same page in two places
							if (resumePageId !== primaryPage.pageId) {
								setupModal(resumePageId, loadPagesIds.length > 0);
							} else {
								console.log("Error: trying to edit the same page in modal");
							}
						},
						error: function(data, status) {
							console.log("Couldn't load pages: " + loadPagesIds);
							if (loadPagesIds.length > 0) {
								// Let's try again, but without trying to resume editing.
								resumePageId = undefined;
								loadPages();
							}
						}
					});
				};
				loadPages();
			});
		},
		link: function(scope, element, attrs) {
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
app.directive("zndEditPage", function($timeout, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/editPage.html",
		scope: {
			pageId: "@",
			isModal: "@",
			// Page this page will "belong" to (e.g. answer belongs to a question,
			// comment belongs to the page it's on)
			primaryPageId: "@",
			// Context, test, and offset are set for editing inline comments
			context: "@",
			text: "@",
			offset: "@",
			// Called when the user is done with the edit.
			doneFn: "&",
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];

			// Set up some helper variables.
			scope.isQuestion = scope.page.type === "question";
			scope.isAnswer = scope.page.type === "answer";
			scope.isComment = scope.page.type === "comment";
			scope.isLens = scope.page.type === "lens";
			scope.isSecondary = scope.isQuestion || scope.isComment;
			scope.useVerticalView = scope.isModal;
			scope.lockedByAnother = scope.page.lockedBy != '0' && scope.page.lockedBy !== userService.user.id;

			// Set up page types.
			if (scope.isQuestion) {
				scope.pageTypes = {question: "Question"};
			} else if(scope.isAnswer) {
				scope.pageTypes = {answer: "Answer"};
			} else if(scope.isComment) {
				scope.pageTypes = {comment: "Comment"};
			} else {
				scope.pageTypes = {wiki: "Wiki Page", blog: "Blog Page", lens: "Lens Page"};
				scope.page.type = scope.page.type in scope.pageTypes ? scope.page.type : "wiki";
			}

			// Set up sort types.
			scope.sortTypes = {
				likes: "By Likes",
				chronological: "Chronologically",
				alphabetical: "Alphabetically",
			};
			scope.page.sortChildrenBy = scope.page.sortChildrenBy in scope.sortTypes ? scope.page.sortChildrenBy : "likes";

			// Set up vote types.
			scope.voteTypes = {
				"": "",
				probability: "Probability",
				approval: "Approval",
			};
			scope.page.voteType = scope.page.voteType in scope.voteTypes ? scope.page.voteType : "";

			var primaryPage = undefined;
			if (scope.primaryPageId) {
				primaryPage = pageService.pageMap[scope.primaryPageId];
			}
			if (scope.isAnswer && primaryPage) {
				// Set up answer page for when it appears on a question page.
				// TODO: shouldn't be setting Parents here
				scope.page.parents = [{parentId: primaryPage.pageId}];
				scope.useVerticalView = true;
			} else if (scope.isComment && primaryPage) {
				scope.useVerticalView = true;
			}

			// Set up group names.
			var groupIds = userService.user.groupIds;
			scope.groupOptions = {"0": "-"};
			if (groupIds) {
				for (var i in groupIds) {
					var groupId = groupIds[i];
					var groupName = userService.groupMap[groupId].name;
					scope.groupOptions[groupId] = groupName;
				}
			}
			scope.groupOptionsLength = groupIds.length + 1;

			// Get the text that should appear on the primary submit button.
			scope.getSubmitButtonText = function() {
				if (scope.isAnswer) {
					return "Submit Answer";
				} else if (scope.isQuestion) {
					return "Submit Question";
				} else if (scope.isComment) {
					return "Comment";
				} else if (scope.isModal) {
					return "Publish & Link";
				}
				return "Publish";
			};
			// Get the text for the placeholder in the title input.
			scope.getTitlePlaceholder = function() {
				if (scope.isAnswer) {
					return "Answer summary";
				} else if (scope.isQuestion) {
					return "Complete sentence question";
				}
				return "Page title";
			}

			if (!scope.isModal) {
				// Create Edit Page JS controller.
				$timeout(function(){
					scope.editPage = new EditPage(scope.page, pageService, autocompleteService, {
						primaryPage: primaryPage,
						topParent: element,
						doneFn: function(result) {
							var continuation = function(data, status) {
								if (scope.doneFn) {
									scope.doneFn({result: result});
								}
							};
							if (result.abandon) {
								pageService.abandonPage(scope.pageId, continuation, continuation);
							} else {
								continuation();
							}
						}
					});
					scope.editPage.start();

					// Listen to destroy event to clean up.
					element.on("$destroy", function(event) {
						scope.editPage.stop();
					});
				});
			}
		},
	};
});
