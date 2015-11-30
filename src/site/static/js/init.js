"use strict";

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
	var $successText = $form.find(".submit-form-success");
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
		$successText.show();
		success(r);
	}).fail(function(r) {
		// Want to show an error even on invisible submit.
		$errorText.show();
		$errorText.text(r.statusText + ": " + r.responseText);
		$successText.hide();
		console.log(r);
		if (error) error();
	});
}
