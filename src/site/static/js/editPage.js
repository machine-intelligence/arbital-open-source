"use strict";

// Create new EditPage (for use with znd-edit-page directive).
// page - page object corresponding to the page being edited.
// pageService - pageService object which contains all loaded pages.
// $topParet - points to the top DOM element of the znd-edit-page directive.
// primaryPage - set only for modal edit page; points to the primary page being edited.
var EditPage = function(page, pageService, $topParent, primaryPage) {
	var page = page;
	var pageId = page.PageId; // id of the page we are editing
	var primaryPage = primaryPage;

	// This array is used for "new parent" autocompletion.
	var availableParents = [];
	for (var fullName in pageAliases) {
		availableParents.push(fullName);
	}

	// Create a new tag for the page.
	var createNewParentElement = function(parentAlias) {
		var $template = $topParent.find(".tag.template");
		var $newTag = $template.clone(true);
		if (parentAlias in pageAliases) {
			var parentPageId = pageAliases[parentAlias].pageId;
			var title = pageAliases[parentAlias].title;
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
		$newTag.attr("id", parentPageId).attr("alias", parentAlias);
		$newTag.insertBefore($template);
		$newTag.attr("title", parentAlias).tooltip();
		availableParents.splice(availableParents.indexOf(parentAlias), 1);
	}

	// Helper function for savePage. Computes the data to submit via AJAX.
	var computeAutosaveData = function(isAutosave, isSnapshot) {
		var parentIds = [];
		$topParent.find(".parent-container").children(".tag:not(.template)").each(function(index, element) {
			parentIds.push($(element).attr("id"));
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
		if (page.WasPublished) {
			// Gah! Since we only display one of the inputs for hasVoteStr and
			// voteType, we have to manually make sure they are synced up, so
			// we can unabmiguously parse it on the server.
			if ($("input[name='hasVoteStr']").is(":visible")) {
				if (!$("input[name='hasVoteStr']").is(":checked")) {
					data["voteType"] = "";
				}
			} else {
				$("input[name='hasVoteStr']").prop("checked", $("input[name='voteType']").val() != "");
			}
		}
		//if (!("hasVoteStr" in data)$("input[name='hasVoteStr']").is(":visible");
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

	// Set up Markdown.
	zndMarkdown.init(true, pageId, "");

	// Process form submission.
	$topParent.find(".new-page-form").on("submit", function(event) {
		var $target = $(event.target);
		var $body = $target.closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		savePage(false, false, function(r) {
			if (primaryPage !== undefined) {
				$(document).trigger("new-page-modal-closed-event", r.substring(r.lastIndexOf("/") + 1));
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

	// Process close button.
	$topParent.find(".new-page-button").on("click", function(event) {
		$(document).trigger("new-page-modal-closed-event", [null, true]);
	});

	// Setup autocomplete for tags.
	$topParent.find(".tag-input").autocomplete({
		source: availableParents,
		minLength: 2,
		select: function (event, ui) {
			createNewParentElement(ui.item.label);
			$(event.target).val("");
			return false;
		}
	});

	// Deleting parents. (Only inside the parent container.)
	var deleteParentTag = function($target) {
		var alias = $target.attr("alias");
		if (alias in pageAliases) {
			availableParents.push(alias);
		}
		$target.tooltip("destroy").remove();
	};
	$topParent.find(".parent-container .tag").on("click", function(event) {
		var $target = $(event.target);
		deleteParentTag($target);
		return false;
	});

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

	// Set up parent options buttons.
	if (primaryPage !== undefined) {
		var deleteAllParentTags = function() {
			$topParent.find(".parent-container").children(".tag:not(.template)").each(function(index, element) {
				deleteParentTag($(element));
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

	// Add parent tags.
	var addParentTags = function() {
		var parentsLen = page.Parents.length;
		for(var n = 0; n < parentsLen; n++) {
			var parentPage = pageService.pageMap[page.Parents[n].ParentId];
			createNewParentElement(parentPage.Alias == "" ? parentPage.PageId : parentPage.Alias);
		}
	};
	addParentTags();

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
	};

	// Called before this editPage is destroyed.
	this.stop = function() {
		clearInterval(this.autosaveInterval);
		// Snapshot just in case.
		savePage(true, false, function(r) {});
	};
};

// EditPageCtrl is the controller for the Edit Page panel.
app.controller("EditPageCtrl", function ($scope, pageService) {
});

app.directive("zndEditPage", function(pageService, userService, $timeout) {
	return {
		templateUrl: "/static/html/editPage.html",
		controller: "EditPageCtrl", // can be a function!
		scope: {
			pageId: "=",
			isModal: "=",
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];

			scope.pageTypes = {
				wiki:"Wiki Page",
				blog:"Blog Page",
			};
			scope.page.Type = scope.page.Type in scope.pageTypes ? scope.page.Type : "wiki";

			scope.sortTypes = {
				likes:"By Likes",
				choronological:"Chronologically",
				alphabetical:"Alphabetically",
			};
			scope.page.SortChildrenBy = scope.page.SortChildrenBy in scope.sortTypes ? scope.page.SortChildrenBy : "likes";

			scope.voteTypes = {
				"":"",
				probability:"Probability",
				approval:"Approval",
			};
			scope.page.VoteType = scope.page.VoteType in scope.voteTypes ? scope.page.VoteType : "";
			scope.showVoteCheckbox = scope.page.WasPublished && scope.page.VoteType != "";

			if (!scope.isModal) {
				$timeout(function(){
					scope.editPage = new EditPage(scope.page, pageService, element, undefined);
					scope.editPage.start();
				});
			}
			scope.groupOptions = {};
			for (var i in userService.user.GroupNames) {
				var group = userService.user.GroupNames[i];
				scope.groupOptions[group] = group;
			}
			scope.groupOptionsLength = userService.user.GroupNames.length;
		},
	};
});
