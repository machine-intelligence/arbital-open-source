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

// Helpful function to serialize an object
$.fn.serializeObject = function() {
  var o = {};
  var a = this.serializeArray();
  $.each(a, function() {
    if (o[this.name] !== undefined) {
      if (!o[this.name].push) {
        o[this.name] = [o[this.name]];
      }
      o[this.name].push(this.value || '');
    } else {
      o[this.name] = this.value || '';
    }
  });
  return o;
};

// Helpful handler to make automatic POST requests with form's contents
var handleFormPostSubmit = function(event) {
  event.preventDefault();
  var formData = $('input').serializeObject();
  $.ajax({
    type: 'POST',
    //TODO: url: $('input').attr('action'),
    url: '/update_contest/',
    data: JSON.stringify(formData),
    dataType: 'json',
    contentType : 'application/json'
  }).always(function(data) {
    $('#update-contest-result').text(data.responseText);
    $('input').trigger("reset");
  });
};
