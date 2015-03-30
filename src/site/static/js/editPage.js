"use strict";

// Create a new tag for the page.
function createNewParentElement(value) {
	var $template = $("#tag-template");
	var $newTag = $template.clone(true);
	var pageId = pageAliases[value].pageId;
	var title = pageAliases[value].title;
	$newTag.removeClass("template");
	$newTag.text(title);
	$newTag.attr("id", pageId).attr("alias", value);
	$newTag.insertBefore($template);
	$newTag.attr("title", value).tooltip();
	availableParents.splice(availableParents.indexOf(value), 1);
}

// Set up Markdown.
$(function() {
	setUpMarkdown(true);
});

// Helper function for calling the pageHandler.
// callback is called when the server replies with success. If it's an autosave
// and the same data has already been submitted, the callback is called with "".
var prevEditPageData = {};
var callPageHandler = function(isAutosave, isSnapshot, callback) {
	var $body = $("body");
	var parentIds = [];
	$body.find("#tag-container").children(".tag:not(.template)").each(function(index, element) {
		parentIds.push($(element).attr("id"));
	});
	var privacyKey = $body.attr("privacy-key");
	var data = {
		pageId: $("body").attr("page-id"),
		parentIds: parentIds.join(),
		privacyKey: privacyKey,
		keepPrivacyKey: $("input[name='private']").is(":checked"),
		karmaLock: +$(".karma-lock-slider").bootstrapSlider("getValue"),
		isAutosave: isAutosave,
		isSnapshot: isSnapshot,
		__invisibleSubmit: isAutosave,
	};
	var $form = $body.find(".new-page-form");
	serializeFormData($form, data);
	// TODO: when we start using Angular, use angular.equals instead
	if (!isAutosave || JSON.stringify(data) !== JSON.stringify(prevEditPageData)) {
		submitForm($form, "/editPage/", data, callback);
		prevEditPageData = data;
	} else {
		callback(undefined);
	}
}

// Set up triggers.
$(function() {
	// Process form submission.
	$(".new-page-form").on("submit", function(event) {
		var $target = $(event.target);
		var $body = $target.closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		callPageHandler(false, false, function(r) {
			window.location.replace(r);
		});
		return false;
	});

	// Process safe snapshot button.
	$(".save-snapshot-button").on("click", function(event) {
		var $body = $(event.target).closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		callPageHandler(false, true, function(r) {
			if (r !== undefined) {
				$body.attr("privacy-key", r);
				$loadingText.show().text("Saved!");
			}
		});
		return false;
	});

	// Show correct options when the type of the page changes.
	$(".type-select").on("change", function(event) {
		$(".type-help").children().hide();
		$(".type-help-" + this.value).show();
	});

	// Setup autocomplete for tags.
	$(".tag-input").autocomplete({
		source: availableParents,
		minLength: 2,
		select: function (event, ui) {
			createNewParentElement(ui.item.label);
			$(event.target).val("");
			return false;
		}
	/*}).on("keyup", function(event) {
		console.log("en");
		if (event.keyCode == 13) {
			var $target = $(event.target);
			createNewParentElement($target.val());
			$target.val("");
			return false;
		}*/
	});

	// Deleting tags. (Only inside the tag container.)
	$("#tag-container .tag").on("click", function(event) {
		var $target = $(event.target);
		var alias = $target.attr("alias");
		if (alias in pageAliases) {
			availableParents.push(alias);
		}
		$target.tooltip("destroy").remove();
		return false;
	});

	// Scroll wmd-panel so it's always inside the viewport.
	var $wmdPreview = $(".wmd-preview");
	var $wmdPanel = $(".wmd-panel");
	var wmdPanelY = $wmdPanel.offset().top;
	var wmdPanelHeight = $wmdPanel.outerHeight();
	$(window).scroll(function(){
		var y = $(window).scrollTop() - wmdPanelY;
		y = Math.min($wmdPreview.outerHeight() - wmdPanelHeight, y);
		y = Math.max(0, y);
		$wmdPanel.stop(true).animate({top: y}, "fast");
	});

	// Keep title label in sync with the title input.
	var $titleLabel = $(".page-title-text");
	$("input[name='title']").on("keyup", function(event) {
		$titleLabel.text($(event.target).val());
	});

	// Set up new tag modal.
	/*var newTagModalSetup = false;
	$("#new-tag-modal").on("shown.bs.modal", function (event) {
		if (newTagModalSetup) return;
		newTagModalSetup = true;

		var $modal = $(event.target);
		var $tagInput = $modal.find(".new-tag-input");
		var $parentTagInput = $modal.find(".parent-tag-input");
		var $parentLink = $parentTagInput.next(".tag");
		$parentTagInput.focus();

		// Set up autocomplete on the parent tag input.
		$parentTagInput.autocomplete({
			source: allTags,
			minLength: 3,
			focus: function (event, ui) {
				$parentTagInput.val(ui.item.label);
				return false;
			},
			select: function (event, ui) {
				event.preventDefault();
				$parentTagInput.val(ui.item.value).toggle();
				$parentLink.text(ui.item.label);
				$tagInput.focus();
				return false;
			}
		});

		// Set up canceling parent tag by clicking on it.
		$parentLink.on("click", function (event) {
			$parentTagInput.val("").toggle().focus();
			$parentLink.text("");
			return false;
		});

		// Process modal buttons to determine if we should close the modal after form submission.
		var closeModalAfterSubmit = false;
		$modal.find(".add-tag-button").on("click", function (event) {
			closeModalAfterSubmit = false;
		});
		$modal.find(".add-tag-close-button").on("click", function (event) {
			closeModalAfterSubmit = true;
		});

		// Process new tag form submit.
		$modal.find(".new-tag-form").on("submit", function (event) {
			var data = {
				parentId: $parentTagInput.val() === "" ? "0" : $parentTagInput.val(),
			};
			submitForm($(event.target), "/newTag/", data, function(r){
				var parts = r.split(",");
				var fullName = parts[0];
				var id = parts[1];
				// Update all tag collections with this new tag.
				availableTags.push(fullName);
				allTags.push({label: fullName, value: id});
				tagMap[fullName] = id;

				$parentTagInput.val("").show();
				$parentLink.text("");
				$tagInput.val("");
				if (closeModalAfterSubmit) {
					$modal.modal("hide");
				} else {
					$modal.find(".alert-success").text("Added: " + fullName).show();
				}
			});
			return false;
		});
	});*/
});

// Trigger initial setup.
$(function() {
	// Update help for the type menu.
	$(".type-select").trigger("change");

	// Process tags that are already being used.
	var parentsLength = parents.length;
	for(var i = 0; i < parentsLength; i++) {
		createNewParentElement(parents[i]);
	}

	// Setup karma lock slider.
	var $slider = $(".karma-lock-slider");
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
});

// Autosave.
var canAutosave = false;
$(function() {
	canAutosave = true;
})
window.setInterval(function(){
	if (!canAutosave) return;
	//$("#autosave-label").text("Saving...").show();
	callPageHandler(true, false, function(r) {
		if (r === undefined) {
			$("#autosave-label").hide();
		} else {
			$("body").attr("privacy-key", r);
			//$("#autosave-label").text("Saved!").show();
		}
	});
}, 5000);
