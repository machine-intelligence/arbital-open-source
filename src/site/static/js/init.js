$(document).ready(function() {
  if (!String.prototype.format) {
    String.prototype.format = function() {
      var args = arguments;
      return this.replace(/{(\d+)}/g, function(match, number) { 
        return typeof args[number] != 'undefined' ? args[number] : match;
      });
    };
  }

	$("#logout").click(function() {
		$.removeCookie("zanaduu", {path: "/"});
	});
});

var submitForm = function($target, url, data, success) {
	var $submitButton = $target.find(":submit");
	var $cancelLink = $target.find(".cancelLink");
	var $loadingIndicator = $target.find(".loadingIndicator");
	var $errorText = $target.find(".errorText");
	$cancelLink.hide();
	$submitButton.hide();
	$loadingIndicator.show();

	$.each($target.serializeArray(), function(i, field) {
		data[field.name] = field.value;
	});

	$.ajax({
		type: 'POST',
		url: url,
		data: JSON.stringify(data),
	})
	.always(function(r) {
		$cancelLink.show();
		$submitButton.show();
		$loadingIndicator.hide();
	}).success(function(r) {
		$errorText.hide();
		success(r);
		console.log(data);
	}).fail(function(r) {
		$errorText.show();
		$errorText.text(r.statusText);
		console.log(data);
	});
}
