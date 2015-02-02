$(document).ready(function() {
	$(".newClaimForm").on("submit", function(event) {
		var data = {};
		$.each($(event.target).serializeArray(), function(i, field) {
			data[field.name] = field.value;
		});
		$.ajax({
			type: 'POST',
			url: '/newClaim/',
			data: JSON.stringify(data),
		})
		.done(function(url, textStatus) {
			window.location.replace(url);
		})
		.fail(function(r) {
			console.log("fail: " + JSON.stringify(r));
			$("#newClaimError").text("An error has occured.");
		});
		return false;
	});
});
