"use strict";

function createNewTagElement(value) {
	var $template = $("#tag-template");
	var $newTag = $template.clone(true);
	$newTag.removeClass("template");
	$newTag.text(value);
	$newTag.attr("id", tagMap[value]);
	$newTag.insertBefore($template);
	availableTags.splice(availableTags.indexOf(value), 1);
}

// Setup Markdown.
$(function() {
	var converter = Markdown.getSanitizingConverter();
	var editor = new Markdown.Editor(converter, "", {handler: function(){
		window.open("http://math.stackexchange.com/editing-help", "_blank");
	}});
	// Convert <embed> tags into a link.
	converter.hooks.chain("preBlockGamut", function (text, rbg) {
		return text.replace(/ {0,3}<embed> *(.+) *<\/embed> */g, function (whole, inner) {
			var s = "";
			s = "[EMBEDDED PAGE](" + inner + ")";
			return rbg(s);
		});
	});
	InitMathjax(converter, editor, "");
	/*converter.hooks.chain("postNormalization", function (text, runSpanGamut) {
		return text.replace(/(.+?)( {0,2}\n)(.[^]*?\n)?([\n]{1,})/g, "$1[[[[1]]]]$2$3$4");
		//return text;
		//return text + "[[[[" + Math.floor(Math.random() * 1000000000) + "]]]]";
		/*return text.replace(/^ {0,3}""" *\n((?:.*?\n)+?) {0,3}""" *$/gm, function (whole, inner) {
			return "<blockquote>" + runBlockGamut(inner) + "</blockquote>\n";
		});
	});*/
	editor.run();
});

// Setup triggers.
$(function() {
	// Helper function for calling the pageHandler
	var callPageHandler = function(isDraft, $body, callback) {
		var tagIds = [];
		$body.find("#tag-container").children(".tag:not(.template)").each(function(index, element) {
			tagIds.push(+$(element).attr("id"));
		});
		var privacyKey = $body.attr("privacy-key");
		var data = {
			pageId: $body.attr("page-id"),
			isDraft: isDraft,
			tagIds: tagIds,
			privacyKey: $("input[name='private']").is(":checked") ? privacyKey : "-1",
			karmaLock: $(".karma-lock-slider").slider("value"),
		};
		submitForm($body.find(".new-page-form"), "/editPage/", data, callback);
	}

	// Process form submission.
	$(".new-page-form").on("submit", function(event) {
		var $target = $(event.target);
		var $body = $target.closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		callPageHandler(false, $body, function(r) {
			window.location.replace(r);
		});
		return false;
	});

	// Process safe draft button.
	$(".save-draft-button").on("click", function(event) {
		var $body = $(event.target).closest("body");
		var $loadingText = $body.find(".loading-text");
		$loadingText.hide();
		callPageHandler(true, $body, function(r) {
			if ($body.attr("page-id") === "0") {
				window.location.replace(r);
			} else {
				var id = (/^\/pages\/edit\/([0-9]+).*$/g).exec(r)[1];
				$loadingText.show().text("Saved!");
				$body.attr("page-id", id);
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
		source: availableTags,
		select: function (event, ui) {
			createNewTagElement(ui.item.label);
			$(event.target).val("");
			return false;
		}
	});

	// Deleting tags. (Only inside the tag container.)
	$(".tag-container .tag").on("click", function(event) {
		var $target = $(event.target);
		availableTags.push($target.text());
		$target.remove();
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
	var newTagModalSetup = false;
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
			focus: function (event, ui) {
				$parentTagInput.val(ui.item.label);
				return false;
			},
			select: function (event, ui) {
				$parentTagInput.val(ui.item.value).toggle();
				$parentLink.text(ui.item.label);
				$tagInput.focus();
				return false;
			}
		});

		// Set up canceling parent tag by clicking on it.
		$parentLink.on("click", function (event) {
			$parentTagInput.val("").toggle();
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
	});
});

// Trigger initial setup.
$(function() {
	// Update help for the type menu.
	$(".type-select").trigger("change");

	// Process tags that are already being used.
	var $tagInput = $(".tag-input");
	var usedTagsLength = usedTags.length;
	for(var i = 0; i < usedTagsLength; i++) {
		createNewTagElement(usedTags[i]);
	}

	// Setup karma lock slider.
	var $slider = $(".karma-lock-slider");
	var $text = $(".karma-lock-text");
	$slider.slider({
		min: 0,
		max: $slider.attr("max"),
		step: Math.max(1, Math.round($slider.attr("max") / 100.0)),
		value: +$text.text(),
		slide: function(event, ui) {
			$text.text(ui.value);
		},
	});
});
