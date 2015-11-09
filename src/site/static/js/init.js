"use strict";

RegExp.escape = function(s) {
    return s.replace(/[-\/\\^$*+?.()|[\]{}]/g, "\\$&");
};

// Prevent Enter key from submitting the form.
// TODO: we should do this for each form that needs it in the place where the form is created.
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

// Setup things correctly.
window.addEventListener("load", function () {
	// Adjust the footer position.
	// TODO: fix this via CSS or something.
	var $footer = $(".page-footer");
	if ($footer.length > 0) {
		var spacerHeight = $(document).height() - $footer.outerHeight() - 1;
		if (spacerHeight > 0) {
			$footer.offset({top: spacerHeight, left: $footer.offset().left});
		}
	}
});

// Return the value of the sParam from the URL.
// TODO: use $location instead.
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
	var lens = getUrlParameter("lens");
	if (lens) {
		lens = "&lens=" + encodeURIComponent(lens);
	} else {
		lens = "";
	}
	window.location.href = window.location.pathname + lens + (hash ? "#" + hash : "");
}

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
var submitForm = function($form, url, data, success, error) {
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
		type: "POST",
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
		if (error) error();
	});
}
