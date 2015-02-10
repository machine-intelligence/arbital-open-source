$(document).ready(function() {
	$(".new-claim-form").on("submit", function(event) {
		var data = {};
		submitForm($(event.target), "/newClaim/", data, function(r) {
			window.location.replace(r);
		});
		return false;
	});
});
