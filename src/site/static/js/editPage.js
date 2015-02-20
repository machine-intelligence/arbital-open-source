function createNewTagElement(value) {
	var $template = $(".tag.template");
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
	var editor = new Markdown.Editor(converter);
	editor.run();
});

// Setup triggers.
$(function() {
	// Helper function for calling the pageHandler
	var callPageHandler = function(isDraft, $body, callback) {
		var tagIds = [];
		$body.find(".tag:not(.template)").each(function(index, element) {
			tagIds.push(+$(element).attr("id"));
		});
		var privacyKey = $body.attr("privacy-key");
		if (privacyKey.length <= 0) { privacyKey = "on"; }
		var data = {
			pageId: $body.attr("page-id"),
			isDraft: isDraft,
			tagIds: tagIds,
			privacyKey: $("input[name='private']").is(":checked") ? privacyKey : "",
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
			var id = (/^\/pages\/([0-9]+).*$/g).exec(r)[1];
			$loadingText.show().text("Saved!");
			$body.attr("page-id", id);
		});
		return false;
	});

	// Show correct options when the type of the page changes.
	$(".type-select").on("change", function(event) {
		$(".type-help").children().hide();
		$(".type-help-" + this.value).show();
		$(".karma-lock-form-group").toggle(this.value !== "blog");
		$(".answers-form-group").toggle(this.value === "question");
	});

	// Setup autocomplete for tags.
	$(".tag-input").autocomplete({
		source: availableTags,
		select: function (event, ui) {
			createNewTagElement(ui.item.value);
			$(event.target).val("");
			return false;
		}
	});

	// Deleting tags.
	$(".tag").on("click", function(event) {
		var $target = $(event.target);
		availableTags.push($target.text());
		$target.remove();
		return false;
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
