"use strict";

// Add various helper functions.
$(function() {
  if (!String.prototype.format) {
    String.prototype.format = function() {
      var args = arguments;
      return this.replace(/{(\d+)}/g, function(match, number) { 
        return typeof args[number] != 'undefined' ? args[number] : match;
      });
    };
  }
});

// Prevent Enter key from submitting the form.
$(function() {
	$(window).keydown(function(event){
		if(event.keyCode == 13 && !$(event.target).is("textarea")) {
			event.preventDefault();
			return false;
		}
	});
});

// Setup event handlers.
$(function() {
	$("#logout").click(function() {
		$.removeCookie("zanaduu", {path: "/"});
	});

	$(".undo-page-delete").on("click", function(event) {
		var data = {
			pageId: $("body").attr("page-id"),
			undoDelete: true,
		};
		$.ajax({
			type: 'POST',
			url: '/deletePage/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
			smartPageReload();
		});
		return false;
	});
});

// Setup things correctly.
$(function() {
	$("#vote-slider-input.example-slider").bootstrapSlider({
		handle: "square",
	});
});

// Setup things correctly.
window.addEventListener('load', function () {
	var $footer = $(".page-footer");
	if ($footer.length > 0) {
		var spacerHeight = $(document).height() - $footer.outerHeight() - 1;
		if (spacerHeight > 0) {
			$footer.offset({top: spacerHeight, left: $footer.offset().left});
		}
	}
});

// Reload the page with a lastVisit parameter so we can pretend that we are
// looking at a page at that time. This way new/updated markers are displayed
// correctly.
function smartPageReload() {
	var url = $("body").attr("page-url");
	var lastVisit = encodeURIComponent($("body").attr("last-visit"));
	window.location.replace(url + "?lastVisit=" + lastVisit);
}
// We don't want to display lastVisit in the URL bar, so we'll erase it.
$(function(){
	if (window.location.href.indexOf("lastVisit") >= 0) {
		history.replaceState(null, document.title, $("body").attr("page-url"));
	}
});

// submitForm handles the common functionality in submitting a form like
// showing/hiding UI elements and doing the AJAX call.
var submitForm = function($target, url, data, success) {
	var $errorText = $target.find(".alert");
	$target.find("[toggle-on-submit]").toggle();

	$.each($target.serializeArray(), function(i, field) {
		data[field.name] = field.value;
	});
	console.log(data);

	$.ajax({
		type: 'POST',
		url: url,
		data: JSON.stringify(data),
	})
	.always(function(r) {
		$target.find("[toggle-on-submit]").toggle();
	}).success(function(r) {
		$errorText.hide();
		success(r);
		console.log(r);
	}).fail(function(r) {
		$errorText.show();
		$errorText.text(r.statusText + ": " + r.responseText);
		console.log(r);
	});
}
