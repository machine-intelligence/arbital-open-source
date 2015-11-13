"use strict";

// Create new EditPage (for use with arb-edit-page directive).
// page - page object corresponding to the page being edited.
// pageService - pageService object which contains all loaded pages.
// options {
//   topParent - points to the arb-edit-page DOM element.
//   primaryPage - for an answer page, points to the question page; for a comment, point to the root page
//   isModal - set if the page is being edited inside a modal
//   doneFn - function to call when the user is done with editing.
// }
// Result returned from doneFn {
// 	abandon - if set to true, the page specified with 'alias' will be deleted
// 	alias - set to page id of the created page
// }
var EditPage = function(page, pageService, userService, autocompleteService, options) {
	var page = page;
	var pageId = page.pageId; // id of the page we are editing
	var options = options || {};
	var primaryPage = options.primaryPage;
	var $topParent = options.topParent;
	var isModal = options.isModal;
	var doneFn = options.doneFn;

	// Update all parent tags. In particular, update whether or not the group is
	// the same as for primary page.
	var updateParentElements = function() {
		$topParent.find(".tag[tag-id]").each(function() {
			var parentPageId = $(this).attr("tag-id");
			var parentPage = pageService.editMap[parentPageId];
			if (!parentPage) {
				parentPage = pageService.pageMap[parentPageId];
			}
			$(this).removeClass("label-danger").addClass("label-default").attr("title", parent.alias).tooltip();
		});
	}
	// Update all the parent tags when the group changes.
	$topParent.find(".group-select").change(function(event) {
		updateParentElements();
	});

	// Create a new tag for the page.
	var createNewParentElement = function(parentId) {
		var parentPage = pageService.editMap[parentId];
		if (!parentPage) {
			parentPage = pageService.pageMap[parentId];
			if (!parentPage) {
				console.error("parent is not in any map: " + parentId);
				return;
			}
		}

		// Prevent duplicates.
		if ($topParent.find(".tag[tag-id=" + parentId + "]").length > 0) return;

		// Create the tag.
		var $template = $topParent.find(".tag.template");
		var $newTag = $template.clone(true);
		$newTag.removeClass("template");
		$newTag.text(parentPage.title === "" ? "*Untitled*" : parentPage.title);
		$newTag.attr("tag-id", parentId);
		$newTag.insertBefore($template);
		updateParentElements();

		// Notify the BE of this potentially new parent connection.
		pageService.newPagePair({
			parentId: parentId,
			childId: pageId,
			type: "parent",
		});
	}
	var deleteParentElement = function($target) {
		var tagId = $target.attr("tag-id");
		$target.tooltip("destroy").remove();

		pageService.deletePagePair({
			parentId: tagId,
			childId: pageId,
			type: "parent",
		});
	};

	// Get similar pages
	var prevSimilarPageData = {};
	var $similarPages = $topParent.find(".similar-pages").find(".panel-body");
	var getComputeSimilarPagesFunc = function($compile, scope) {
		return createThrottledCallback(function() {
			var fullPageData = computeAutosaveData(false, false);
			var data = {
				title: fullPageData.title,
				text: fullPageData.text,
				clickbait: fullPageData.clickbait,
				pageType: fullPageData.type,
			};
			if (JSON.stringify(data) === JSON.stringify(prevSimilarPageData)) return false;
			prevSimilarPageData = data;
			autocompleteService.findSimilarPages(data, function(data){
				$similarPages.empty();
				for (var n = 0; n < data.length; n++) {
					var pageId = data[n].label;
					if (pageId === page.pageId) continue;
					var $el = $compile("<span class='admin' ng-show='userService.user.isAdmin'>" + data[n].score + "</span>" +
						"<div arb-likes-page-title page-id='" + pageId +
						"' show-clickbait='true'></div>")(scope);
					$similarPages.append($el);
				}
			});
			return true;
		}, 2000);
	};


	// Helper function for savePageInfo. Computes the data to submit via AJAX.
	var computePageInfoData = function() {
		var data = {
			pageId: pageId,
			editKarmaLock: +$topParent.find(".karma-lock-slider").bootstrapSlider("getValue"),
		};
		serializeFormData($topParent.find(".page-info-form"), data);
		return data;
	};
	var prevPageInfoData = undefined;
	// Save the page info.
	// callback is called when the server replies with success.
	var savePageInfo = function(){
		// Submit the form.
		var data = computePageInfoData();
		var $form = $topParent.find(".page-info-form");
		if (JSON.stringify(data) !== JSON.stringify(prevPageInfoData)) {
			submitForm($form, "/editPageInfo/", data, function(r) {
			});
			prevPageInfoData = data;
		}
	}

	// Process form submission for page options.
	$topParent.find(".page-info-form").on("submit", function(event) {
		var $target = $(event.target);
		var $body = $target.closest("body");
		savePageInfo();
		return false;
	});

	// Helper function for savePage. Computes the data to submit via AJAX.
	var computeAutosaveData = function(isAutosave, isSnapshot) {
		var data = {
			pageId: pageId,
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
	var prevEditPageData = undefined;
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
				if (isAutosave) {
					autosaving = false;
					// Refresh the lock
					page.lockedUntil = moment.utc().add(30, "m").format("YYYY-MM-DD HH:mm:ss");
				}
				if (isSnapshot) {
					// Prevent an autosave from triggering right after a successful snapshot
					prevEditPageData.isSnapshot = false;
					prevEditPageData.isAutosave = true;
					data.__invisibleSubmit = true; 
				}

				// Process warnings
				var warningsLength = r.result.aliasWarnings.length;
				var $aliasWarning = $topParent.find(".alias-warning");
				$aliasWarning.text("").toggle(warningsLength > 0);
				$topParent.find(".alias-form-group").toggleClass("has-warning", warningsLength > 0);
				for(var n = 0; n < warningsLength; n++){
					$aliasWarning.text($aliasWarning.text() + r.result.aliasWarnings[n] + "\n");
				}
				callback(true);
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
			callback(false);
			autosaving = false;
		}
	}

	// Process form submission.
	$topParent.find(".new-page-form").on("submit", function(event) {
		var $target = $(event.target);
		var $body = $target.closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		savePage(false, false, function(saved) {
			if (doneFn) {
				doneFn({alias: pageId});
			}
		});
		return false;
	});

	// Process safe snapshot button.
	$topParent.find(".save-snapshot-button").on("click", function(event) {
		var $body = $(event.target).closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		savePage(false, true, function(saved) {
			if (saved) {
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

	if (primaryPage && pageId === primaryPage.pageId) {
		// Resive textarea height to fit the screen.
		$topParent.find(".wmd-input").height($(window).height() - 140);
	}

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
	var addParentTags = function() {
		var parentsLen = page.parents.length;
		for(var n = 0; n < parentsLen; n++) {
			createNewParentElement(page.parents[n].parentId);
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

	// Setup autocomplete for parents field.
	autocompleteService.setupParentsAutocomplete($topParent.find(".tag-input"), function(event, ui) {
		createNewParentElement(ui.item.label);
		$(event.target).val("");
		return false;
	});

	// Setup autocomplete for user field.
	autocompleteService.setupUserAutocomplete($topParent.find(".tag-input"), function(event, ui) {
		createNewParentElement(ui.item.label);
		$(event.target).val("");
		return false;
	});

	// Add existing parent tags
	addParentTags();

	// Convert all links with pageIds to alias links.
	page.text = page.text.replace(complexLinkRegexp, function(whole, prefix, text, alias) {
		var page = pageService.pageMap[alias];
		if (page) {
			return prefix + "[" + text + "](" + page.alias + ")";
		}
		return whole;
	/*}).replace(voteEmbedRegexp, function (whole, prefix, alias) {
		var page = pageService.pageMap[alias];
		if (page) {
			return prefix + "[vote: " + page.alias + "]";
		}
		return whole;*/
	}).replace(forwardLinkRegexp, function (whole, prefix, alias, text) {
		var page = pageService.pageMap[alias];
		if (page) {
			return prefix + "[" + page.alias + " " + text + "]";
		}
		return whole;
	}).replace(simpleLinkRegexp, function (whole, prefix, alias) {
		var page = pageService.pageMap[alias];
		if (page) {
			return prefix + "[" + page.alias + "]";
		}
		return whole;
	}).replace(urlLinkRegexp, function(whole, prefix, text, url, alias) {
		var page = pageService.pageMap[alias];
		if (page) {
			return prefix + "[" + text + "](" + url + page.alias + ")";
		}
		return whole;
	});

	// Set up Markdown.
	arbMarkdown.init(true, pageId, "", undefined, pageService, userService, autocompleteService);

	// Setup karma lock slider.
	var $slider = $topParent.find(".karma-lock-slider");
	$slider.bootstrapSlider({
		value: page.editKarmaLock,
		min: 0,
		max: +$slider.attr("max"),
		step: 1,
		precision: 0,
		selection: "none",
		handle: "square",
		tooltip: "always",
	});

	// Process click event to revert the page to a certain edit
	$("body").on("click", ".edit-node-revert-to-edit", function(event) {
		var $target = $(event.target);
		var data = {
			pageId: pageId,
			editNum: +$target.attr("edit-num"),
		};
		$.ajax({
			type: "POST",
			url: "/revertPage/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
			window.location.href = page.url();
		});
	});

	// diffPage stores the edit we load for diffing.
	var diffPage;
	// Refresh the diff edit text.
	var refreshDiff = function() {
		var dmp = new diff_match_patch();
		var diffs = dmp.diff_main(diffPage.text, $("#wmd-input" + pageId).val());
		dmp.diff_cleanupSemantic(diffs);
		var html = dmp.diff_prettyHtml(diffs);
		$topParent.find(".edit-diff").html(html);
	}
	// Show/hide the diff edit.
	var showDiff = function(show) {
		var $diffHalf = $topParent.find(".diff-half");
		if (show) {
			$diffHalf.css("display", "inline-block");
		} else {
			$diffHalf.hide();
		}
		$topParent.find(".preview-half").toggle(!show);
	}
	// Process click event for diffing edits.
	$("body").on("click", ".edit-node-diff-edit", function(event) {
		// Load the edit from the server.
		pageService.loadEdit({
			pageAlias: pageId,
			specificEdit: $(event.target).attr("edit-num"),
			skipProcessDataStep: true,
			success: function(data, status) {
				diffPage = data[pageId];
				refreshDiff();
				showDiff(true);
				var $diffHalf = $topParent.find(".diff-half");
				$diffHalf.find(".edit-num-text").text("(#" + diffPage.edit + ")");
				$diffHalf.find(".page-title-text").text(diffPage.title);
			},
		});
	});
	// Process click event for diffing edits.
	$topParent.on("click", ".refresh-diff", function(event) {
		refreshDiff();
	});
	// Process click event for hiding diff.
	$topParent.on("click", ".hide-diff", function(event) {
		showDiff(false);
	});

	// Start initializes things that have to be killed when this editPage stops existing.
	this.autosaveInterval = null;
	this.similarPagesInterval = null;
	this.backdropInterval = null;
	this.start = function($compile, scope) {
		// Hide new page button if this is a modal.
		$topParent.find("#wmd-new-page-button" + pageId).toggle(!isModal);

		// Set the rendering for parents autocomplete
		autocompleteService.setAutocompleteRendering($topParent.find(".tag-input"), scope);

		// Set up link suggestions for the primary markdown textarea.
		$topParent.find(".wmd-input").textcomplete([

			{
				match: /\[([A-Za-z0-9.]+)$/,
				search: function (term, callback) {
					autocompleteService.parentsSource({term: term}, callback);
				},
				template: function (item) {
					return "<span class='search-result' arb-likes-page-title page-id='" + item.value +
						"' show-clickbait='true' is-search-result='true'></span>";
				},
				replace: function (value) {
					return ["[" + value.alias, "]"];
				},
				index: 1,
				cache: true,
			},
			{
				match: /\[(@[A-Za-z0-9.]+)$/,
				search: function (term, callback) {
					autocompleteService.userSource({term: term}, callback);
				},
				template: function (item) {
					return "<span class='search-result' arb-likes-page-title page-id='" + item.value +
						"' show-clickbait='true' is-search-result='true'></span>";
				},
				replace: function (value) {
					return ["[@" + value.alias, "]"];
				},
				index: 1,
				cache: true,
			},
		],
		{
			appendTo: $("body"),
			zIndex: 10001,
			header: function (data) {
				// HACK: we need to compile the angular template code, and header() is
				// called any time there is any kind of change, so we call compile here.
				setTimeout(function() {
					$compile($(".textcomplete-dropdown"))(scope);
				});
				return "";
	    },
		});

		// Create change logs if necessary
		if (page.changeLogs) {
			setTimeout(function() {
				$topParent.find(".change-logs").append($compile("<arb-change-logs page-id='" + pageId +
						"'></arb-change-logs>")(scope));
			});
		}

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
		var autosaveFunc = function(){
			$autosaveLabel.text("Autosave: Saving...").show();
			savePage(true, false, function(saved) {
				if (saved) {
					$autosaveLabel.text("Autosave: Saved!").show();
				} else {
					$autosaveLabel.hide();
				}
			});
		};
		this.autosaveInterval = window.setInterval(autosaveFunc, 5000);
		window.setTimeout(function() {
			// Compute prevEditPageData, so we don't fire off autosave when there were
			// no changes made.
			prevEditPageData = computeAutosaveData(true, false);
			prevPageInfoData = computePageInfoData();
		});

		// Set up finding similar pages
		if (page.type !== "comment") {
			var func = getComputeSimilarPagesFunc($compile, scope);
			scope.$watch("page.title", func);
			scope.$watch("page.clickbait", func);
			scope.$watch("page.text", func);
		}

		// Set up interval for updating meta-data
		var $metaTextInput = $topParent.find(".meta-text-input");
		var $metaTextError = $topParent.find(".meta-text-error");
		this.metaTextInterval = window.setInterval(function(){
			try {
				$metaTextError.hide();
				jsyaml.load($metaTextInput.val());
			} catch (err) {
				$metaTextError.text(err.message).show();
			} 
		}, 1300);

		// Workaround: Set up an interval to make sure modal backdrop is the right size.
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
		clearInterval(this.similarPagesInterval);
		clearInterval(this.backdropInterval);
		// Autosave just in case.
		savePage(true, false, function(saved) {});
	};
};



// Directive for the modal, where a user can create a new page, edit a page, 
// ask a question, etc...
app.directive("arbEditPageModal", function (pageService, userService) {
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
				var primaryPage = pageService.editMap[options.parentPageId];
				if (!primaryPage) {
					primaryPage = pageService.pageMap[options.parentPageId];
				}
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
					var newPage = pageService.editMap[pageId];
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
					}

					// Dynamically create arb-edit-page directive.
					var el = $compile("<arb-edit-page page-id='" + pageId +
							"' is-modal='true'" +
							"done-fn='doneFn(result)'></arb-edit-page>")($scope);
					$modalBody.empty().append(el);
					$modal.modal();

					// Handle modal's shown event to initialize editPage script.
					// This result will be returned if the user just hides the modal.
					var returnedResult = {hidden: true, alias: pageId}; 
					var editPage;
					$modal.on("shown.bs.modal", function (e) {
						editPage = new EditPage(newPage, pageService, userService, autocompleteService, {
							topParent: el,
							primaryPage: primaryPage,
							isModal: true,
							doneFn: function(result) {
								returnedResult = result;
								$modal.modal("hide");
								if (result.abandon) {
									pageService.abandonPage(result.alias);
								}
								if (result.abandon || result.alias) {
									resumePageId = undefined;
								}
							},
						});
						editPage.start($compile, $scope);
					});
					// Hande modal's close event and return the resulting alias.
					$modal.on("hidden.bs.modal", function (e) {
						pageIdCache[options.modalKey + pageService.primaryPage.pageId] = resumePageId
						if (options.callback) {
							// Make sure we got alias and not pageId
							var tempEditPage = pageService.editMap[returnedResult.alias];
							returnedResult.alias = tempEditPage.alias;
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
					if (resumePageId) {
						if (resumePageId === primaryPage.pageId) {
							console.error("trying to edit the same page in modal");
							return;
						}
						// Resume editing some page.
						pageService.loadEdit({
							pageAlias: resumePageId,
							success: function(data, status) {
								setupModal(resumePageId, true);
							},
							error: function(data, status) {
								// Let's try again, but without trying to resume editing.
								resumePageId = undefined;
								loadPages();
							},
						});
					} else {
						pageService.getNewPage({
							success: function(newPageId) {
								resumePageId = newPageId;
								if (resumePageId !== primaryPage.pageId) {
									setupModal(resumePageId, false);
								} else {
									console.error("trying to edit the same page in modal");
								}
							},
						});
					}
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
app.directive("arbEditPage", function($timeout, $compile, pageService, userService, autocompleteService) {
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
			scope.pageService = pageService;
			scope.page = pageService.editMap[scope.pageId];
			if (!scope.page && scope.pageId in pageService.pageMap) {
				// We are going to edit a page that's not in the editMap yet. Add it.
				scope.page = $.extend(true, {}, pageService.pageMap[scope.pageId]);
				pageService.addPageToEditMap(scope.page);
			}

			// Fix alias
			if (!scope.page.alias) {
				scope.page.alias = scope.page.pageId;
			}

			// If the page has "Group.Alias" alias, just change it to "Alias"
			var dotIndex = scope.page.alias.indexOf(".");
			if (dotIndex >= 0) {
				scope.page.alias = scope.page.alias.substring(dotIndex + 1);
			}

			// Set up some helper variables.
			scope.isQuestion = scope.page.type === "question";
			scope.isAnswer = scope.page.type === "answer";
			scope.isComment = scope.page.type === "comment";
			scope.isLens = scope.page.type === "lens";
			scope.isSecondary = scope.isQuestion || scope.isComment;
			scope.isGroup = scope.page.type === "group" || scope.page.type === "domain";
			scope.isFixedType = scope.isSecondary || scope.isGroup;
			scope.useVerticalView = scope.isModal;
			scope.lockExists = scope.page.lockedBy != "0" && moment.utc(scope.page.lockedUntil).isAfter(moment.utc());
			scope.lockedByAnother = scope.lockExists && scope.page.lockedBy !== userService.user.id;

			// Set up page types.
			if (scope.isQuestion) {
				scope.pageTypes = {question: "Question"};
			} else if(scope.isAnswer) {
				scope.pageTypes = {answer: "Answer"};
			} else if(scope.isComment) {
				scope.pageTypes = {comment: "Comment"};
			} else {
				scope.pageTypes = {wiki: "Wiki Page", lens: "Lens Page"};
				scope.page.type = scope.page.type in scope.pageTypes ? scope.page.type : "wiki";
			}

			// Set up sort types.
			scope.sortTypes = {
				likes: "By Likes",
				recentFirst: "Recent First",
				oldestFirst: "Oldest First",
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
				primaryPage = pageService.editMap[scope.primaryPageId];
				if (!primaryPage) {
					primaryPage = pageService.pageMap[scope.primaryPageId];
				}
			}
			if (scope.isAnswer && primaryPage) {
				// Set up answer page for when it appears on a question page.
				// TODO: shouldn't be setting Parents here
				scope.page.parents = [{parentId: primaryPage.pageId}];
				scope.useVerticalView = true;
			} else if ((scope.isComment || scope.isQuestion) && primaryPage) {
				scope.useVerticalView = true;
			}

			// Set up group names.
			var groupIds = userService.user.groupIds;
			scope.groupOptions = {"0": "-"};
			if (groupIds) {
				for (var i in groupIds) {
					var groupId = groupIds[i];
					var groupName = pageService.pageMap[groupId].title;
					scope.groupOptions[groupId] = groupName;
				}
			}
			// Also check if we are part of the necessary group.
			scope.groupPermissionsPassed = true;
			if (!(scope.page.editGroupId in scope.groupOptions)) {
				scope.groupPermissionsPassed = false;
				scope.groupOptions[scope.page.editGroupId] = pageService.pageMap[scope.page.editGroupId].title;
			}

			// if starting a new edit, clear the minor edit checkbox
			if (scope.page.isCurrentEdit) {
		    scope.page.isMinorEdit = false;
			}

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
					scope.editPage = new EditPage(scope.page, pageService, userService, autocompleteService, {
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
					scope.editPage.start($compile, scope);

					// Listen to destroy event to clean up.
					element.on("$destroy", function(event) {
						scope.editPage.stop();
					});
				});
			}
		},
	};
});

// Directive for showing page's change log.
app.directive("arbChangeLogs", function(pageService, userService) {
	return {
		templateUrl: "/static/html/changeLogs.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.pageService = pageService;
			scope.page = pageService.editMap[scope.pageId];
		},
	};
});
