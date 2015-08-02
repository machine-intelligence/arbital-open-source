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
		if(event.keyCode == 13) {
			var target = $(event.target);
			if(!(target.is("textarea") || target.closest("#new-link-modal").length > 0)) {
				event.preventDefault();
				return false;
			}
		}
	});
});

// Setup search via navbar.
$(function() {
	var $navSearch = $("#nav-search");
	if ($navSearch.length <= 0) return;
  $navSearch.autocomplete({
		source: "/json/search",
		minLength: 4,
		delay: 500,
		focus: function (event, ui) {
			return false;
		},
		select: function (event, ui) {
			window.location.href = "/pages/" + ui.item.value;
			return false;
		},
  });
	$navSearch.data("ui-autocomplete")._renderItem = function(ul, item) {
		var group = item.label.groupName ? "[" + item.label.groupName + "] " : "";
		var alias = !+item.label.alias ? " (" + item.label.alias + ")" : "";
		var title = item.label.title ? item.label.title : "COMMENT";
	  return $("<li>")
	    .attr("data-value", item.value)
	    .append(group + title + alias)
	    .appendTo(ul);
	};
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
window.addEventListener("load", function () {
	// Adjust the footer position.
	var $footer = $(".page-footer");
	if ($footer.length > 0) {
		var spacerHeight = $(document).height() - $footer.outerHeight() - 1;
		if (spacerHeight > 0) {
			$footer.offset({top: spacerHeight, left: $footer.offset().left});
		}
	}
});

// Return the value of the sParam from the URL.
function getUrlParameter(sParam) {
	var sPageURL = window.location.search.substring(1);
	var sURLVariables = sPageURL.split('&');
	for (var i = 0; i < sURLVariables.length; i++) {
		var sParameterName = sURLVariables[i].split('=');
		if (sParameterName[0] == sParam) {
			return decodeURIComponent(sParameterName[1]);
		}
	}
} 

// Reload the page with a lastVisit parameter so we can pretend that we are
// looking at a page at that time. This way new/updated markers are displayed
// correctly.
function smartPageReload(hash) {
	var lastVisit = encodeURIComponent($("body").attr("last-visit"));
	if (!lastVisit) {
		lastVisit = "0";
	}
	window.location.href = window.location.pathname + "?lastVisit=" + lastVisit + (hash ? "#" + hash : "");
}
// We don't want certain url parameters cluttering up the url, so we'll erase them.
$(function(){
	var lastVisit = getUrlParameter("lastVisit");
	if (lastVisit) {
		$("body").attr("last-visit", lastVisit);
		history.replaceState(null, document.title,
			window.location.origin + window.location.pathname + window.location.hash);
	}
});

// serializeFormData takes input values from the given form and returns them as
// a map. Optionally, data can have pre-existing map values.
var serializeFormData = function($form, data) {
	if (data === undefined) data = {};
	$.each($form.serializeArray(), function(i, field) {
		data[field.name] = field.value;
	});
	data["__formSerialized"] = true;
	return data;
}

// submitForm handles the common functionality in submitting a form like
// showing/hiding UI elements and doing the AJAX call.
var submitForm = function($form, url, data, success) {
	var $errorText = $form.find(".submit-form-error");
	var invisibleSubmit = data["__invisibleSubmit"];
	if (!invisibleSubmit) {
		$form.find("[toggle-on-submit]").toggle();
	}

	if (!("__formSerialized" in data)) {
		serializeFormData($form, data);
	}
	console.log("Sending POST to " + url + ":"); console.log(data);

	$.ajax({
		type: 'POST',
		url: url,
		data: JSON.stringify(data),
	})
	.always(function(r) {
		if (!invisibleSubmit) {
			$form.find("[toggle-on-submit]").toggle();
		}
	}).success(function(r) {
		$errorText.hide();
		success(r);
	}).fail(function(r) {
		// Want to show an error even on invisible submit.
		$errorText.show();
		$errorText.text(r.statusText + ": " + r.responseText);
		console.log(r);
	});
}
