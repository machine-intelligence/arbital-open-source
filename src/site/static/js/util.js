"use strict";

// Used to escape regexp symbols in a string to make it safe for injecting into a regexp
RegExp.escape = function(s) {
    return s.replace(/[-\/\\^$*+?.()|[\]{}]/g, "\\$&");
};

// Turn a callback function into a cleverly throttled version.
// Callback parameter should return true if the lock is to be set.
// Basically, we want:
// 1) Instant callback if the delay is met
// 2) Otherwise, wait to call the callback until delay is met
// 3) If we are waiting on the delay, don't stack additional calls
var createThrottledCallback = function(callback, delay) {
	// waitLock is set when we are waiting on the delay.
	var waitLock = false;
	// Timeout is set when we need to do the callback after the delay
	var timeout = undefined;

	var result = function() {
		if (waitLock) {
			if (!timeout) {
				timeout = window.setTimeout(function() {
					timeout = undefined;
					result();
				}, delay);
			}
			return;
		}
		if (callback()) {
			waitLock = true;
			window.setTimeout(function() {
				waitLock = false;
			}, delay);
		}
	};
	return result;
};

// submitForm handles the common functionality in submitting a form like
// showing/hiding UI elements and doing the AJAX call.
var submitForm = function($form, url, data, success, error) {
	var $errorText = $form.find(".submit-form-error");
	var $successText = $form.find(".submit-form-success");
	var invisibleSubmit = data["__invisibleSubmit"];
	if (!invisibleSubmit) {
		$form.find("[toggle-on-submit]").toggle();
	}

	console.log("Submitting form to " + url + ":"); console.log(data);
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
		console.error(r);
		if (error) error();
	});
}

